package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/redis/go-redis/v9"
)

type ServiceInstance struct {
	ID        string            `json:"id"`
	Service   string            `json:"service"`
	BaseURL   string            `json:"base_url"`
	HealthURL string            `json:"health_url"`
	Version   string            `json:"version"`
	StartedAt string            `json:"started_at"`
	UpdatedAt string            `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ServiceRegistry struct {
	client      *redis.Client
	namespace   string
	ttl         time.Duration
	heartbeat   time.Duration
	instance    ServiceInstance
	instanceKey string
	indexKey    string
	now         func() time.Time
	mu          sync.Mutex
}

func StartServiceRegistration(cfg *config.Config, client *redis.Client, serviceName string, metadata map[string]string) (*ServiceRegistry, context.CancelFunc, error) {
	if cfg == nil || !cfg.Registry.Enabled || client == nil {
		return nil, nil, nil
	}

	registry, err := NewServiceRegistry(cfg, client, serviceName, metadata)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go registry.HeartbeatLoop(ctx)
	return registry, cancel, nil
}

func NewServiceRegistry(cfg *config.Config, client *redis.Client, serviceName string, metadata map[string]string) (*ServiceRegistry, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	baseURL, err := cfg.Services.BaseURLFor(serviceName)
	if err != nil {
		return nil, err
	}

	host, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s-%s-%d", serviceName, sanitizeRegistryName(host), time.Now().UnixNano())
	startedAt := time.Now().UTC().Format(time.RFC3339)
	namespace := strings.TrimSpace(cfg.Registry.Namespace)
	if namespace == "" {
		namespace = "workpal:registry"
	}
	ttl := time.Duration(cfg.Registry.TTLMS) * time.Millisecond
	if ttl <= 0 {
		ttl = 15 * time.Second
	}
	heartbeat := time.Duration(cfg.Registry.HeartbeatMS) * time.Millisecond
	if heartbeat <= 0 {
		heartbeat = ttl / 3
	}
	if heartbeat <= 0 {
		heartbeat = 5 * time.Second
	}

	instanceKey := fmt.Sprintf("%s:service:%s:%s", namespace, serviceName, instanceID)
	indexKey := fmt.Sprintf("%s:index:%s", namespace, serviceName)
	return &ServiceRegistry{
		client:    client,
		namespace: namespace,
		ttl:       ttl,
		heartbeat: heartbeat,
		instance: ServiceInstance{
			ID:        instanceID,
			Service:   serviceName,
			BaseURL:   strings.TrimRight(baseURL, "/"),
			HealthURL: strings.TrimRight(baseURL, "/") + "/health",
			Version:   Version,
			StartedAt: startedAt,
			UpdatedAt: startedAt,
			Metadata:  cloneStringMap(metadata),
		},
		instanceKey: instanceKey,
		indexKey:    indexKey,
		now:         time.Now,
	}, nil
}

func (r *ServiceRegistry) Register(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.instance.UpdatedAt = r.now().UTC().Format(time.RFC3339)
	payload, err := json.Marshal(r.instance)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.instanceKey, payload, r.ttl)
	pipe.SAdd(ctx, r.indexKey, r.instanceKey)
	pipe.Expire(ctx, r.indexKey, r.ttl*4)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *ServiceRegistry) HeartbeatLoop(ctx context.Context) {
	if err := r.Register(ctx); err != nil {
		log.Printf("[%s] register service instance: %v", r.instance.Service, err)
	}

	ticker := time.NewTicker(r.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.Register(ctx); err != nil {
				log.Printf("[%s] refresh service registration: %v", r.instance.Service, err)
			}
		}
	}
}

func (r *ServiceRegistry) Deregister(ctx context.Context) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, r.instanceKey)
	pipe.SRem(ctx, r.indexKey, r.instanceKey)
	_, err := pipe.Exec(ctx)
	return err
}

func ListServiceInstances(ctx context.Context, cfg *config.Config, client *redis.Client, serviceName string) ([]ServiceInstance, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	namespace := strings.TrimSpace(cfg.Registry.Namespace)
	if namespace == "" {
		namespace = "workpal:registry"
	}
	indexKey := fmt.Sprintf("%s:index:%s", namespace, serviceName)
	keys, err := client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return []ServiceInstance{}, nil
	}

	values, err := client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	instances := make([]ServiceInstance, 0, len(values))
	missingKeys := make([]string, 0)
	for idx, raw := range values {
		if raw == nil {
			missingKeys = append(missingKeys, keys[idx])
			continue
		}

		var encoded string
		switch value := raw.(type) {
		case string:
			encoded = value
		case []byte:
			encoded = string(value)
		default:
			missingKeys = append(missingKeys, keys[idx])
			continue
		}

		var instance ServiceInstance
		if err := json.Unmarshal([]byte(encoded), &instance); err != nil {
			missingKeys = append(missingKeys, keys[idx])
			continue
		}
		instances = append(instances, instance)
	}

	if len(missingKeys) > 0 {
		_ = client.SRem(ctx, indexKey, missingKeys).Err()
	}
	return instances, nil
}

func cloneStringMap(value map[string]string) map[string]string {
	if len(value) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

func sanitizeRegistryName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "localhost"
	}
	replacer := strings.NewReplacer(":", "-", "/", "-", "\\", "-", " ", "-", ".", "-")
	return replacer.Replace(value)
}
