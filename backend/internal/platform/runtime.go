package platform

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/Tangyd893/WorkPal/backend/pkg/cache"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const Version = "0.4.0"

type HealthCheck struct {
	Name  string
	Check func(context.Context) error
}

func ResolveConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	candidates := []string{
		filepath.Join("configs", "config.yaml"),
		filepath.Join("configs", "config.example.yaml"),
		filepath.Join("backend", "configs", "config.yaml"),
		filepath.Join("backend", "configs", "config.example.yaml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join("configs", "config.yaml")
}

func LoadConfig() (*config.Config, error) {
	cfg, err := config.Load(ResolveConfigPath())
	if err != nil {
		return nil, err
	}
	if cfg.Server.JWTSecret == "" {
		return nil, fmt.Errorf("server.jwtSecret cannot be empty")
	}
	auth.SetSecret(cfg.Server.JWTSecret)
	return cfg, nil
}

func OpenServiceDB(cfg *config.Config, serviceName string) (*gorm.DB, *sql.DB, error) {
	dbName, err := cfg.Database.ServiceDBName(serviceName)
	if err != nil {
		return nil, nil, err
	}
	if err := ensureDatabaseExists(cfg, dbName); err != nil {
		return nil, nil, err
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN(dbName)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	return db, sqlDB, nil
}

func OpenRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func InitCache(cfg *config.Config) error {
	return cache.Init(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
}

func NewRouter(cfg *config.Config, serviceName string) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "WorkPal",
			"service": serviceName,
			"version": Version,
			"status":  "running",
		})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return r
}

func RegisterHealth(r *gin.Engine, serviceName string, checks ...HealthCheck) {
	r.GET("/health", func(c *gin.Context) {
		components := make(map[string]string, len(checks))
		errorsByComponent := map[string]string{}
		statusCode := http.StatusOK
		statusText := "ok"

		for _, check := range checks {
			if check.Check == nil {
				components[check.Name] = "ok"
				continue
			}
			if err := check.Check(c.Request.Context()); err != nil {
				components[check.Name] = "degraded"
				errorsByComponent[check.Name] = err.Error()
				statusCode = http.StatusServiceUnavailable
				statusText = "degraded"
				continue
			}
			components[check.Name] = "ok"
		}

		body := gin.H{
			"status":     statusText,
			"service":    serviceName,
			"components": components,
		}
		if len(errorsByComponent) > 0 {
			body["errors"] = errorsByComponent
		}
		c.JSON(statusCode, body)
	})
}

func SQLHealthCheck(name string, sqlDB *sql.DB) HealthCheck {
	return HealthCheck{
		Name: name,
		Check: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
	}
}

func RedisHealthCheck(name string, client *redis.Client) HealthCheck {
	return HealthCheck{
		Name: name,
		Check: func(ctx context.Context) error {
			return client.Ping(ctx).Err()
		},
	}
}

func HTTPHealthCheck(name, targetURL string, client *http.Client) HealthCheck {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return HealthCheck{
		Name: name,
		Check: func(ctx context.Context) error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
			if err != nil {
				return err
			}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("health endpoint returned HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}
}

func NamedHealthCheck(name string, check func(context.Context) error) HealthCheck {
	return HealthCheck{Name: name, Check: check}
}

func RunHTTP(serviceName string, port int, handler http.Handler, onShutdown func()) error {
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr, Handler: handler}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Printf("[%s] shutting down", serviceName)
		if onShutdown != nil {
			onShutdown()
		}
		_ = srv.Close()
	}()

	log.Printf("[%s] listening on http://localhost%s", serviceName, addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func ensureDatabaseExists(cfg *config.Config, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	adminName := cfg.Database.AdminDBName
	if adminName == "" {
		adminName = "postgres"
	}

	adminDB, err := sql.Open("pgx", cfg.Database.DSN(adminName))
	if err != nil {
		return fmt.Errorf("open admin database connection: %w", err)
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	if err := adminDB.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence for %s: %w", dbName, err)
	}
	if exists {
		return nil
	}

	query := fmt.Sprintf(`CREATE DATABASE "%s"`, strings.ReplaceAll(dbName, `"`, `""`))
	if _, err := adminDB.ExecContext(ctx, query); err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}
	return nil
}
