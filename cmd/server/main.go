package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Tangyd893/WorkPal/configs"
	"github.com/Tangyd893/WorkPal/internal/user/handler"
	"github.com/Tangyd893/WorkPal/internal/user/repo"
	"github.com/Tangyd893/WorkPal/internal/user/service"
	"github.com/Tangyd893/WorkPal/internal/user/model"
	"github.com/Tangyd893/WorkPal/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := configs.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置 JWT 密钥（供 auth pkg 使用）
	auth.SetSecret(cfg.Server.JWTSecret)

	// 2. 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	// 3. 自动迁移 Schema
	if err := autoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("✓ 数据库迁移完成")

	// 4. 初始化依赖（Repo → Service → Handler）
	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.Server.JWTExpiryHours)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, authSvc)

	// 5. 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 6. 创建 Gin Engine
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 根路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "WorkPal",
			"version": "0.1.0",
			"status":  "running",
		})
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 路由组
	apiV1 := r.Group("/api/v1")
	userHandler.RegisterRoutes(apiV1)

	// 7. 启动
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🚀 服务启动中: http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// autoMigrate 自动迁移所有表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&usermodel.User{},
		&usermodel.Department{},
	)
}
