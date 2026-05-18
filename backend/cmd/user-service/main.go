package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	userHandler "github.com/Tangyd893/WorkPal/backend/internal/user/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	userRepo "github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	userService "github.com/Tangyd893/WorkPal/backend/internal/user/service"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "user-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(
		&model.User{}, &model.Department{}, &model.Employee{},
		&model.Role{}, &model.Permission{}, &model.RolePermission{},
		&model.UserRole{}, &model.ProjectRole{}, &model.ProjectRolePermission{},
		&model.ProjectMember{},
	); err != nil {
		log.Fatalf("migrate user service schema: %v", err)
	}

	userRepoInst := userRepo.NewUserRepo(db)
	rbacRepoInst := userRepo.NewRBACRepo(db)
	if cfg.Server.Mode != "release" {
		if err := platform.EnsureDevelopmentUsers(context.Background(), db, userRepoInst); err != nil {
			log.Fatalf("seed development users: %v", err)
		}
		if err := platform.EnsureRBACSeed(context.Background(), db); err != nil {
			log.Printf("seed rbac data: %v", err)
		}
		log.Printf("development users ensured (%d accounts)", platform.DevelopmentUserCount())
	}

	authSvc := userService.NewAuthService(userRepoInst, cfg.Server.JWTExpiryHours)
	userSvc := userService.NewUserService(userRepoInst)
	rbacSvc := userService.NewRBACService(rbacRepoInst)
	userHdlr := userHandler.NewUserHandler(userSvc, authSvc)
	rbacHdlr := userHandler.NewRBACHandler(rbacSvc)

	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		registryRedis, redisErr := platform.OpenRedis(cfg)
		if redisErr != nil {
			log.Printf("[user-service] service registry unavailable: %v", redisErr)
		} else {
			registryRedisCloser = registryRedis
			registry, registryStop, redisErr = platform.StartServiceRegistration(cfg, registryRedis, "user-service", map[string]string{
				"domain": "identity",
			})
			if redisErr != nil {
				log.Printf("[user-service] register service instance: %v", redisErr)
			}
		}
	}

	r := platform.NewRouter(cfg, "user-service")
	platform.RegisterHealth(r, "user-service", platform.SQLHealthCheck("postgres", sqlDB))
	apiV1 := r.Group("/api/v1")
	userHdlr.RegisterRoutes(apiV1)
	rbacHdlr.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("user-service", cfg.Services.UserPort, r, func() {
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
		log.Fatalf("user service stopped: %v", err)
	}
}
