package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/clients"
	"github.com/Tangyd893/WorkPal/backend/internal/events"
	imModel "github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	searchHandler "github.com/Tangyd893/WorkPal/backend/internal/search/handler"
	searchSvc "github.com/Tangyd893/WorkPal/backend/internal/search/service"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	redisClient, err := platform.OpenRedis(cfg)
	if err != nil {
		log.Fatalf("open redis: %v", err)
	}
	defer redisClient.Close()

	searchService, err := searchSvc.NewSearchService(cfg.Search.Bleve.IndexPath)
	if err != nil {
		log.Fatalf("open search index: %v", err)
	}
	defer searchService.Close()

	queue := msgqueue.NewRedisStreams(redisClient, cfg.Redis.StreamsKey, "workpal-search")
	msgqueue.Init(queue)
	subscribeMessageEvents(queue, searchService)

	registry, registryStop, registryErr := platform.StartServiceRegistration(cfg, redisClient, "search-service", map[string]string{
		"domain": "search",
		"index":  cfg.Search.Engine,
	})
	if registryErr != nil {
		log.Printf("[search-service] register service instance: %v", registryErr)
	}

	convSvc := clients.NewIMClient(cfg.Services.IMURL, cfg.Server.InternalToken)
	searchHdlr := searchHandler.NewSearchHandler(searchService, convSvc)

	r := platform.NewRouter(cfg, "search-service")
	platform.RegisterHealth(
		r,
		"search-service",
		platform.RedisHealthCheck("redis", redisClient),
		platform.NamedHealthCheck("bleve", func(ctx context.Context) error {
			_, err := searchService.Search(ctx, "health", 1, 1)
			return err
		}),
	)
	apiV1 := r.Group("/api/v1")
	searchHdlr.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("search-service", cfg.Services.SearchPort, r, func() {
		if registry != nil {
			_ = registry.Deregister(context.Background())
		}
		if registryStop != nil {
			registryStop()
		}
	}); err != nil {
		log.Fatalf("search service stopped: %v", err)
	}
}

func subscribeMessageEvents(queue msgqueue.Interface, searchService *searchSvc.SearchService) {
	if err := queue.SubscribeWithOptions(events.TopicMessageUpserted, msgqueue.SubscribeOptions{
		Consumer:      "search-message-upsert-indexer",
		MaxRetries:    5,
		DeadLetterKey: "workpal:streams:messages:dead",
	}, func(data []byte) error {
		var msg imModel.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		if err := searchService.IndexMessage(&msg); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("subscribe message upsert events: %v", err)
	}

	if err := queue.SubscribeWithOptions(events.TopicMessageDeleted, msgqueue.SubscribeOptions{
		Consumer:      "search-message-delete-indexer",
		MaxRetries:    5,
		DeadLetterKey: "workpal:streams:messages:dead",
	}, func(data []byte) error {
		var event events.MessageDeletedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		if err := searchService.DeleteMessage(event.ID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("subscribe message delete events: %v", err)
	}
}
