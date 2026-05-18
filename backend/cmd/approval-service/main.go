package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/approval/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/approval/model"
	"github.com/Tangyd893/WorkPal/backend/internal/approval/repo"
	"github.com/Tangyd893/WorkPal/backend/internal/approval/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil { log.Fatalf("load config: %v", err) }
	db, sqlDB, err := platform.OpenServiceDB(cfg, "approval-service")
	if err != nil { log.Fatalf("open database: %v", err) }
	defer sqlDB.Close()
	if err := db.AutoMigrate(&model.ApprovalTemplate{}, &model.ApprovalInstance{}, &model.ApprovalAction{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	svc := service.NewService(repo.NewRepo(db))
	h := handler.NewHandler(svc)
	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		if redis, err := platform.OpenRedis(cfg); err == nil {
			registryRedisCloser = redis
			registry, registryStop, err = platform.StartServiceRegistration(cfg, redis, "approval-service", map[string]string{"domain": "approval"})
			if err != nil { log.Printf("[approval] register: %v", err) }
		}
	}
	r := platform.NewRouter(cfg, "approval-service")
	platform.RegisterHealth(r, "approval-service", platform.SQLHealthCheck("postgres", sqlDB))
	h.RegisterRoutes(r.Group("/api/v1"))
	if err := platform.RunHTTP("approval-service", cfg.Services.ApprovalPort, r, func() {
		if registry != nil { _ = registry.Deregister(context.Background()) }
		if registryStop != nil { registryStop() }
		if registryRedisCloser != nil { _ = registryRedisCloser.Close() }
	}); err != nil { log.Fatalf("stopped: %v", err) }
}
