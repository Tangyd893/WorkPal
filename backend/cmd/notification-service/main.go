package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	"github.com/Tangyd893/WorkPal/backend/internal/notification/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/notification/model"
	"github.com/Tangyd893/WorkPal/backend/internal/notification/repo"
	"github.com/Tangyd893/WorkPal/backend/internal/notification/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "notification-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&model.Notification{}, &audit.Log{}); err != nil {
		log.Fatalf("migrate notification service schema: %v", err)
	}

	r := repo.NewRepo(db)
	svc := service.NewService(r)
	h := handler.NewHandler(svc)

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[notification-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "notification-service", map[string]string{
				"domain": "notification",
			})
			if redisErr != nil {
				log.Printf("[notification-service] register service instance: %v", redisErr)
			}
		}
	}

	router := platform.NewRouter(cfg, "notification-service")
	platform.RegisterHealth(router, "notification-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := router.Group("/api/v1")
	h.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("notification-service", cfg.Services.NotificationPort, router, func() {
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
		log.Fatalf("notification service stopped: %v", err)
	}
}
