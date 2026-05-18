package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/analytics"
	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	projectHandler "github.com/Tangyd893/WorkPal/backend/internal/project/handler"
	projectModel "github.com/Tangyd893/WorkPal/backend/internal/project/model"
	projectRepo "github.com/Tangyd893/WorkPal/backend/internal/project/repo"
	projectService "github.com/Tangyd893/WorkPal/backend/internal/project/service"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "project-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(
		&projectModel.Project{},
		&projectModel.IssueType{},
		&projectModel.Issue{},
		&projectModel.Board{},
		&projectModel.Version{},
		&projectModel.IssueChangelog{},
		&projectModel.Association{},
		&projectModel.Workflow{},
		&projectModel.CustomFieldDef{},
		&projectModel.CustomFieldValue{},
		&audit.Log{},
	); err != nil {
		log.Fatalf("migrate project service schema: %v", err)
	}

	repo := projectRepo.NewRepo(db)
	svc := projectService.NewService(repo)
	analyticsSvc := analytics.NewService(db, repo)
	handler := projectHandler.NewHandler(svc, analyticsSvc, audit.NewRecorder(db))

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[project-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "project-service", map[string]string{
				"domain": "project",
			})
			if redisErr != nil {
				log.Printf("[project-service] register service instance: %v", redisErr)
			}
		}
	}

	r := platform.NewRouter(cfg, "project-service")
	platform.RegisterHealth(r, "project-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := r.Group("/api/v1")
	handler.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("project-service", cfg.Services.ProjectPort, r, func() {
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
		log.Fatalf("project service stopped: %v", err)
	}
}
