package main

import (
	"context"
	"log"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/clients"
	fileHandler "github.com/Tangyd893/WorkPal/backend/internal/file/handler"
	fileModel "github.com/Tangyd893/WorkPal/backend/internal/file/model"
	fileRepo "github.com/Tangyd893/WorkPal/backend/internal/file/repo"
	fileSvc "github.com/Tangyd893/WorkPal/backend/internal/file/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "file-service")
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

	convSvc := clients.NewIMClient(cfg.Services.IMURL, cfg.Server.InternalToken)
	fileHdlr := fileHandler.NewFileHandler(fileService, convSvc)

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[file-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "file-service", map[string]string{
				"domain": "files",
				"store":  cfg.File.StoreType,
			})
			if redisErr != nil {
				log.Printf("[file-service] register service instance: %v", redisErr)
			}
		}
	}

	r := platform.NewRouter(cfg, "file-service")
	platform.RegisterHealth(r, "file-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := r.Group("/api/v1")
	fileHdlr.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("file-service", cfg.Services.FilePort, r, func() {
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
