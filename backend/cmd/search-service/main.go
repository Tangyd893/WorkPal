package main

import (
	"encoding/json"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/events"
	imModel "github.com/Tangyd893/WorkPal/backend/internal/im/model"
	imRepo "github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	imService "github.com/Tangyd893/WorkPal/backend/internal/im/service"
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

	db, sqlDB, err := platform.OpenDB(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

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

	convRepo := imRepo.NewConversationRepo(db)
	convSvc := imService.NewConversationService(convRepo)
	searchHdlr := searchHandler.NewSearchHandler(searchService, convSvc)

	r := platform.NewRouter(cfg, "search-service")
	platform.RegisterHealth(r, sqlDB, redisClient)
	apiV1 := r.Group("/api/v1")
	searchHdlr.RegisterRoutes(apiV1)

	if err := platform.RunHTTP("search-service", cfg.Services.SearchPort, r, nil); err != nil {
		log.Fatalf("search service stopped: %v", err)
	}
}

func subscribeMessageEvents(queue msgqueue.Interface, searchService *searchSvc.SearchService) {
	if err := queue.Subscribe(events.TopicMessageUpserted, func(data []byte) {
		var msg imModel.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("decode message upsert event: %v", err)
			return
		}
		if err := searchService.IndexMessage(&msg); err != nil {
			log.Printf("index message %d: %v", msg.ID, err)
		}
	}); err != nil {
		log.Printf("subscribe message upsert events: %v", err)
	}

	if err := queue.Subscribe(events.TopicMessageDeleted, func(data []byte) {
		var event events.MessageDeletedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("decode message delete event: %v", err)
			return
		}
		if err := searchService.DeleteMessage(event.ID); err != nil {
			log.Printf("delete message %d from search index: %v", event.ID, err)
		}
	}); err != nil {
		log.Printf("subscribe message delete events: %v", err)
	}
}
