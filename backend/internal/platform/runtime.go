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
	"syscall"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/Tangyd893/WorkPal/backend/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const Version = "0.3.0"

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

func OpenDB(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
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

func RegisterHealth(r *gin.Engine, sqlDB *sql.DB, redisClient *redis.Client) {
	r.GET("/health", func(c *gin.Context) {
		components := map[string]string{}
		if sqlDB != nil {
			if err := sqlDB.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "degraded",
					"error":  "postgres unreachable",
				})
				return
			}
			components["postgres"] = "ok"
		}
		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "degraded",
					"error":  "redis unreachable",
				})
				return
			}
			components["redis"] = "ok"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    "workpal",
			"components": components,
		})
	})
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
