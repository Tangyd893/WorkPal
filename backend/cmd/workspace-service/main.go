package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	workspaceHandler "github.com/Tangyd893/WorkPal/backend/internal/workspace/handler"
	workspaceModel "github.com/Tangyd893/WorkPal/backend/internal/workspace/model"
	workspaceRepo "github.com/Tangyd893/WorkPal/backend/internal/workspace/repo"
	workspaceService "github.com/Tangyd893/WorkPal/backend/internal/workspace/service"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "workspace-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&workspaceModel.Task{}, &workspaceModel.ScheduleEvent{}); err != nil {
		log.Fatalf("migrate workspace service schema: %v", err)
	}

	repo := workspaceRepo.NewRepo(db)
	svc := workspaceService.NewService(repo)
	handler := workspaceHandler.NewHandler(svc)

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[workspace-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "workspace-service", map[string]string{
				"domain": "workspace",
			})
			if redisErr != nil {
				log.Printf("[workspace-service] register service instance: %v", redisErr)
			}
		}
	}

	r := platform.NewRouter(cfg, "workspace-service")
	platform.RegisterHealth(r, "workspace-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := r.Group("/api/v1")
	handler.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("workspace-service", cfg.Services.WorkspacePort, r, func() {
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
		log.Fatalf("workspace service stopped: %v", err)
	}
}
