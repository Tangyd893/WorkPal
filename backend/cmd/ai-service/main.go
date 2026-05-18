package main

import (
	"context"
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/ai/handler"
	"github.com/Tangyd893/WorkPal/backend/internal/ai/service"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

type noopSearch struct{}

func (n *noopSearch) SearchMessages(ctx context.Context, query string, convID *int64, page, pageSize int) ([]service.MessageResult, error) {
	return nil, nil
}

func (n *noopSearch) SearchIssues(ctx context.Context, query string, projectID *int64) ([]service.IssueResult, error) {
	return nil, nil
}

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil { log.Fatalf("load config: %v", err) }
	svc := service.NewService(&noopSearch{})
	h := handler.NewHandler(svc)
	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	var registryRedisCloser interface{ Close() error }
	if cfg.Registry.Enabled {
		if redis, err := platform.OpenRedis(cfg); err == nil {
			registryRedisCloser = redis
			registry, registryStop, err = platform.StartServiceRegistration(cfg, redis, "ai-service", map[string]string{"domain": "ai"})
			if err != nil { log.Printf("[ai] register: %v", err) }
		}
	}
	r := platform.NewRouter(cfg, "ai-service")
	platform.RegisterHealth(r, "ai-service", platform.HealthCheck{Name: "ai", Check: func(ctx context.Context) error { return nil }})
	h.RegisterRoutes(r.Group("/api/v1"))
	if err := platform.RunHTTP("ai-service", cfg.Services.AIPort, r, func() {
		if registry != nil { _ = registry.Deregister(context.Background()) }
		if registryStop != nil { registryStop() }
		if registryRedisCloser != nil { _ = registryRedisCloser.Close() }
	}); err != nil { log.Fatalf("stopped: %v", err) }
}
