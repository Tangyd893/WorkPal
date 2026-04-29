package main

import (
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

	db, sqlDB, err := platform.OpenDB(cfg)
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

	r := platform.NewRouter(cfg, "workspace-service")
	platform.RegisterHealth(r, sqlDB, nil)
	apiV1 := r.Group("/api/v1")
	handler.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("workspace-service", cfg.Services.WorkspacePort, r, nil); err != nil {
		log.Fatalf("workspace service stopped: %v", err)
	}
}
