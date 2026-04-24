package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	fileHandler "github.com/Tangyd893/WorkPal/backend/internal/file/handler"
	fileSvc "github.com/Tangyd893/WorkPal/backend/internal/file/service"
	fileRepo "github.com/Tangyd893/WorkPal/backend/internal/file/repo"
	fileModel "github.com/Tangyd893/WorkPal/backend/internal/file/model"
	imHandler "github.com/Tangyd893/WorkPal/backend/internal/im/handler"
	imModel "github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	imWS "github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	searchHandler "github.com/Tangyd893/WorkPal/backend/internal/search/handler"
	searchSvc "github.com/Tangyd893/WorkPal/backend/internal/search/service"
	userHandler "github.com/Tangyd893/WorkPal/backend/internal/user/handler"
	userRepo "github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	userService "github.com/Tangyd893/WorkPal/backend/internal/user/service"
	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/Tangyd893/WorkPal/backend/pkg/cache"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// Prometheus 指标
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	wsConnectionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)
	messagesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "messages_total",
			Help: "Total number of messages sent",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, wsConnectionsGauge, messagesTotal)
}

func main() {
	// 1. 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置 JWT 密钥
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

	// 3. 初始化 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	if err := cache.Init(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB); err != nil {
		log.Fatalf("初始化 cache 失败: %v", err)
	}
	log.Println("✓ Redis 连接成功")

	// 4. 初始化 Redis Streams 消息队列
	mq := msgqueue.NewRedisStreams(redisClient, cfg.Redis.StreamsKey, "workpal-consumers")
	msgqueue.Init(mq)
	log.Println("✓ Redis Streams 消息队列初始化完成")

	// 5. 初始化 Bleve 搜索服务
	var searchService *searchSvc.SearchService
	searchService, err = searchSvc.NewSearchService(cfg.Search.Bleve.IndexPath)
	if err != nil {
		log.Printf("⚠ Bleve 搜索服务初始化失败: %v (搜索功能将不可用)", err)
		searchService = nil
	} else {
		log.Println("✓ Bleve 搜索服务初始化完成")
	}

	// 6. 自动迁移 Schema
	if err := autoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("✓ 数据库迁移完成")

	// 7. 初始化依赖

	// User 模块
	userRepoInst := userRepo.NewUserRepo(db)
	authSvc := userService.NewAuthService(userRepoInst, cfg.Server.JWTExpiryHours)
	userSvc := userService.NewUserService(userRepoInst)
	userHdlr := userHandler.NewUserHandler(userSvc, authSvc)

	// IM 模块
	convRepoInst := repo.NewConversationRepo(db)
	msgRepoInst := repo.NewMessageRepo(db)
	convSvc := service.NewConversationService(convRepoInst)
	msgSvc := service.NewMessageService(msgRepoInst)
	_ = service.NewPresenceService()

	// WebSocket Hub
	hub := imWS.InitHub()

	convHandler := imHandler.NewConversationHandler(convSvc)
	msgHandler := imHandler.NewMessageHandler(msgSvc, convSvc, hub, searchService)

	// File 模块
	fRepo := fileRepo.NewFileRepo(db)
	var fStore fileSvc.FileStore
	var fStoreErr error
	if cfg.File.StoreType == "minio" {
		fStore, fStoreErr = fileSvc.NewMinIOFileStore(
			cfg.File.MinIO.Endpoint,
			cfg.File.MinIO.AccessKey,
			cfg.File.MinIO.SecretKey,
			cfg.File.MinIO.Bucket,
			cfg.File.MinIO.UseSSL,
		)
		if fStoreErr != nil {
			log.Printf("⚠ MinIO 初始化失败: %v，使用本地存储", fStoreErr)
			fStore = fileSvc.NewLocalFileStore(cfg.File.LocalBaseDir)
			log.Println("✓ 文件存储初始化完成（本地模式）")
		} else {
			log.Println("✓ 文件存储初始化完成（MinIO 模式）")
		}
	} else {
		fStore = fileSvc.NewLocalFileStore(cfg.File.LocalBaseDir)
		log.Println("✓ 文件存储初始化完成（本地模式）")
	}
	fileService := fileSvc.NewFileService(fRepo, fStore, cfg.File.MaxFileSizeMB)
	fHdlr := fileHandler.NewFileHandler(fileService)

	// Search 模块
	var searchHdlr *searchHandler.SearchHandler
	if searchService != nil {
		searchHdlr = searchHandler.NewSearchHandler(searchService, convSvc)
	}

	// 8. 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 9. 创建 Gin Engine
	r := gin.New()

	// Prometheus 中间件
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path))
		c.Next()
		timer.ObserveDuration()
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	})

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 根路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "WorkPal",
			"version": "0.2.0",
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
	userHdlr.RegisterRoutes(apiV1)

	// IM 路由
	imGroup := apiV1.Group("")
	imGroup.Use(middleware.AuthRequired())
	convHandler.RegisterRoutes(imGroup)
	msgHandler.RegisterRoutes(imGroup)
	fHdlr.RegisterRoutes(imGroup)

	// 搜索路由
	if searchHdlr != nil {
		searchHdlr.RegisterRoutes(imGroup)
	}

	// WebSocket 端点
	wsHandler := imHandler.NewWebSocketHandler(hub)
	r.GET("/ws", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "missing token"})
			return
		}
		tokenStr := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenStr = authHeader[7:]
		}
		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		wsHandler.Handle(c)
	})

	// 10. 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("收到退出信号，正在关闭...")
		if searchService != nil {
			searchService.Close()
		}
		sqlDB.Close()
		os.Exit(0)
	}()

	log.Printf("🚀 WorkPal 服务启动中: http://localhost%s", addr)
	log.Printf("📊 Prometheus 指标: http://localhost%s/metrics", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// autoMigrate 自动迁移所有表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Department{},
		&imModel.Conversation{},
		&imModel.ConversationMember{},
		&imModel.Message{},
		&imModel.MessageRead{},
		&fileModel.File{},
	)
}
