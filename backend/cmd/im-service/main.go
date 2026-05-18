package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	"github.com/Tangyd893/WorkPal/backend/internal/im/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	imWS "github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, sqlDB, err := platform.OpenServiceDB(cfg, "im-service")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	redisClient, err := platform.OpenRedis(cfg)
	if err != nil {
		log.Fatalf("open redis: %v", err)
	}
	defer redisClient.Close()
	if err := platform.InitCache(cfg); err != nil {
		log.Fatalf("init cache: %v", err)
	}
	msgqueue.Init(msgqueue.NewRedisStreams(redisClient, cfg.Redis.StreamsKey, "workpal-im"))

	if err := db.AutoMigrate(
		&model.Conversation{},
		&model.ConversationMember{},
		&model.Message{},
		&model.MessageOutbox{},
		&model.MessageRead{},
		&model.Channel{},
		&model.ChannelMember{},
		&model.Thread{},
		&audit.Log{},
	); err != nil {
		log.Fatalf("migrate im service schema: %v", err)
	}

	convRepo := repo.NewConversationRepo(db)
	msgRepo := repo.NewMessageRepo(db)
	outboxRepo := repo.NewOutboxRepo(db)
	convSvc := service.NewConversationService(convRepo)
	commandStore := service.NewGormMessageCommandStore(db, msgRepo, outboxRepo)
	msgSvc := service.NewReliableMessageService(msgRepo, commandStore)
	_ = service.NewPresenceService()
	outboxPublisher := service.NewOutboxPublisher(outboxRepo, msgqueue.NewRedisStreams(redisClient, cfg.Redis.StreamsKey, "workpal-im-outbox"))
	outboxCtx, outboxCancel := context.WithCancel(context.Background())
	go outboxPublisher.Run(outboxCtx)
	hub := imWS.InitHub()
	clusterBroadcaster := imWS.NewClusterBroadcaster(redisClient, hub, "")
	hub.SetClusterBroadcaster(clusterBroadcaster)
	clusterCtx, clusterCancel := context.WithCancel(context.Background())
	go clusterBroadcaster.Run(clusterCtx)

	registry, registryStop, registryErr := platform.StartServiceRegistration(cfg, redisClient, "im-service", map[string]string{
		"domain":   "im",
		"realtime": "websocket",
	})
	if registryErr != nil {
		log.Printf("[im-service] register service instance: %v", registryErr)
	}

	auditRecorder := audit.NewRecorder(db)
	convHandler := handler.NewConversationHandler(convSvc, auditRecorder)
	msgHandler := handler.NewMessageHandler(msgSvc, convSvc, hub, nil, clusterBroadcaster, auditRecorder)
	wsHandler := handler.NewWebSocketHandler(hub, convSvc)

	r := platform.NewRouter(cfg, "im-service")
	platform.RegisterHealth(
		r,
		"im-service",
		platform.SQLHealthCheck("postgres", sqlDB),
		platform.RedisHealthCheck("redis", redisClient),
	)
	apiV1 := r.Group("/api/v1")
	channelRepo := repo.NewChannelRepo(db)
	channelHandler := handler.NewChannelHandler(channelRepo)
	convHandler.RegisterRoutes(apiV1)
	msgHandler.RegisterRoutes(apiV1)
	channelHandler.RegisterRoutes(apiV1)
	convHandler.RegisterInternalRoutes(r.Group(""), cfg.Server.InternalToken)
	registerWebSocket(r, wsHandler)

	if err := platform.RunHTTP("im-service", cfg.Services.IMPort, r, func() {
		if registry != nil {
			_ = registry.Deregister(context.Background())
		}
		if registryStop != nil {
			registryStop()
		}
		outboxCancel()
		clusterCancel()
		hub.Stop()
	}); err != nil {
		log.Fatalf("im service stopped: %v", err)
	}
}

func registerWebSocket(r *gin.Engine, wsHandler *handler.WebSocketHandler) {
	r.GET("/ws", func(c *gin.Context) {
		var tokenStr string
		if t := c.Query("token"); t != "" {
			tokenStr = t
		} else if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			tokenStr = authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		wsHandler.Handle(c)
	})
}
