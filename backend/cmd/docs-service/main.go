package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/docs/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/docs/model"
	"github.com/Tangyd893/WorkPal/backend/internal/docs/repo"
	"github.com/Tangyd893/WorkPal/backend/internal/docs/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "docs-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&model.Document{}, &model.DocumentRevision{}); err != nil {
		log.Fatalf("migrate docs service schema: %v", err)
	}

	repoInst := repo.NewRepo(db)
	svc := service.NewService(repoInst)
	h := handler.NewHandler(svc)

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[docs-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "docs-service", map[string]string{
				"domain": "docs",
			})
			if redisErr != nil {
				log.Printf("[docs-service] register service instance: %v", redisErr)
			}
		}
	}

	r := platform.NewRouter(cfg, "docs-service")
	platform.RegisterHealth(r, "docs-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := r.Group("/api/v1")
	h.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("docs-service", cfg.Services.DocsPort, r, func() {
		if registry != nil {
			_ = registry.Deregister(context.Background())
		}
		if registryStop != nil {
			registryStop()
		}
		if registryRedisCloser != nil {
			_ = registryRedisCloser.Close()
		}
	}); err != nil {
		log.Fatalf("docs service stopped: %v", err)
	}
}
