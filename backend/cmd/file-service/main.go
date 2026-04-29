package main

import (
	"log"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	fileHandler "github.com/Tangyd893/WorkPal/backend/internal/file/handler"
	fileModel "github.com/Tangyd893/WorkPal/backend/internal/file/model"
	fileRepo "github.com/Tangyd893/WorkPal/backend/internal/file/repo"
	fileSvc "github.com/Tangyd893/WorkPal/backend/internal/file/service"
	imRepo "github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	imService "github.com/Tangyd893/WorkPal/backend/internal/im/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
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

	if err := db.AutoMigrate(&fileModel.File{}); err != nil {
		log.Fatalf("migrate file service schema: %v", err)
	}

	store := newFileStore(cfg)
	fileRepoInst := fileRepo.NewFileRepo(db)
	fileService := fileSvc.NewFileService(fileRepoInst, store, cfg.File.MaxFileSizeMB)

	convRepo := imRepo.NewConversationRepo(db)
	convSvc := imService.NewConversationService(convRepo)
	fileHdlr := fileHandler.NewFileHandler(fileService, convSvc)

	r := platform.NewRouter(cfg, "file-service")
	platform.RegisterHealth(r, sqlDB, nil)
	apiV1 := r.Group("/api/v1")
	fileHdlr.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("file-service", cfg.Services.FilePort, r, nil); err != nil {
		log.Fatalf("file service stopped: %v", err)
	}
}

func newFileStore(cfg *config.Config) fileSvc.FileStore {
	if cfg.File.StoreType == "minio" {
		store, err := fileSvc.NewMinIOFileStore(
			cfg.File.MinIO.Endpoint,
			cfg.File.MinIO.AccessKey,
			cfg.File.MinIO.SecretKey,
			cfg.File.MinIO.Bucket,
			cfg.File.MinIO.UseSSL,
		)
		if err == nil {
			log.Println("file storage initialized in MinIO mode")
			return store
		}
		log.Printf("minio initialization failed, falling back to local storage: %v", err)
	}

	log.Println("file storage initialized in local mode")
	return fileSvc.NewLocalFileStore(cfg.File.LocalBaseDir)
}
