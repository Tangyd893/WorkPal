package msgqueue

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStreams Redis Streams 实现
type RedisStreams struct {
	client *redis.Client
	key    string
	group  string
	mu     sync.RWMutex
	groups map[string]struct{}
}

var _ Interface = (*RedisStreams)(nil)

// NewRedisStreams 创建 Redis Streams 队列
func NewRedisStreams(client *redis.Client, key, group string) *RedisStreams {
	return &RedisStreams{
		client: client,
		key:    key,
		group:  group,
		groups: make(map[string]struct{}),
	}
}

// Publish 发布消息
func (r *RedisStreams) Publish(ctx context.Context, topic string, msg []byte) error {
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.key,
		Values: map[string]interface{}{
			"topic": topic,
			"data":  string(msg),
			"time":  time.Now().UnixMilli(),
		},
	}).Err()
}

// Subscribe 订阅主题
func (r *RedisStreams) Subscribe(topic string, handler func([]byte)) error {
	return r.SubscribeWithOptions(topic, SubscribeOptions{}, func(data []byte) error {
		handler(data)
		return nil
	})
}

func (r *RedisStreams) SubscribeWithOptions(topic string, options SubscribeOptions, handler Handler) error {
	options = r.normalizeOptions(topic, options)
	if err := r.ensureGroup(context.Background(), options.Group); err != nil {
		return err
	}
	r.trackGroup(options.Group)

	go r.consume(topic, options, handler)
	go r.reclaimPending(topic, options, handler)
	return nil
}

func (r *RedisStreams) consume(topic string, options SubscribeOptions, handler Handler) {
	ctx := context.Background()
	for {
		streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    options.Group,
			Consumer: options.Consumer,
			Streams:  []string{r.key, ">"},
			Count:    10,
			Block:    time.Second,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				r.handleMessage(ctx, topic, options, msg, handler)
			}
		}
	}
}

func (r *RedisStreams) reclaimPending(topic string, options SubscribeOptions, handler Handler) {
	ticker := time.NewTicker(options.ClaimMinIdle)
	defer ticker.Stop()

	ctx := context.Background()
	for range ticker.C {
		pending, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: r.key,
			Group:  options.Group,
			Start:  "-",
			End:    "+",
			Count:  20,
		}).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("[redis-streams] read pending messages: %v", err)
			}
			continue
		}

		for _, pendingMsg := range pending {
			if pendingMsg.Idle < options.ClaimMinIdle {
				continue
			}
			if int64(pendingMsg.RetryCount) >= options.MaxRetries {
				r.movePendingToDeadLetter(ctx, pendingMsg.ID, options)
				continue
			}

			claimed, err := r.client.XClaim(ctx, &redis.XClaimArgs{
				Stream:   r.key,
				Group:    options.Group,
				Consumer: options.Consumer,
				MinIdle:  options.ClaimMinIdle,
				Messages: []string{pendingMsg.ID},
			}).Result()
			if err != nil {
				log.Printf("[redis-streams] claim pending message %s: %v", pendingMsg.ID, err)
				continue
			}
			for _, msg := range claimed {
				r.handleMessage(ctx, topic, options, msg, handler)
			}
		}
	}
}

func (r *RedisStreams) handleMessage(ctx context.Context, topic string, options SubscribeOptions, msg redis.XMessage, handler Handler) {
	topicVal, ok := msg.Values["topic"].(string)
	if !ok {
		_ = r.client.XAck(ctx, r.key, options.Group, msg.ID).Err()
		return
	}
	if topicVal != topic {
		if err := r.client.XAck(ctx, r.key, options.Group, msg.ID).Err(); err != nil {
			log.Printf("[redis-streams] ack skipped topic=%s id=%s group=%s: %v", topicVal, msg.ID, options.Group, err)
		}
		return
	}

	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		_ = r.client.XAck(ctx, r.key, options.Group, msg.ID).Err()
		return
	}
	if err := handler([]byte(dataStr)); err != nil {
		log.Printf("[redis-streams] handler failed topic=%s id=%s: %v", topic, msg.ID, err)
		return
	}
	if err := r.client.XAck(ctx, r.key, options.Group, msg.ID).Err(); err != nil {
		log.Printf("[redis-streams] ack failed topic=%s id=%s: %v", topic, msg.ID, err)
	}
}

func (r *RedisStreams) movePendingToDeadLetter(ctx context.Context, messageID string, options SubscribeOptions) {
	claimed, err := r.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   r.key,
		Group:    options.Group,
		Consumer: options.Consumer,
		MinIdle:  0,
		Messages: []string{messageID},
	}).Result()
	if err != nil {
		log.Printf("[redis-streams] claim dead-letter message %s: %v", messageID, err)
		return
	}
	for _, msg := range claimed {
		values := map[string]interface{}{
			"stream":      r.key,
			"group":       options.Group,
			"source_id":   msg.ID,
			"topic":       msg.Values["topic"],
			"data":        msg.Values["data"],
			"failed_at":   time.Now().UnixMilli(),
			"max_retries": options.MaxRetries,
		}
		if err := r.client.XAdd(ctx, &redis.XAddArgs{Stream: options.DeadLetterKey, Values: values}).Err(); err != nil {
			log.Printf("[redis-streams] add dead-letter message %s: %v", msg.ID, err)
			continue
		}
		if err := r.client.XAck(ctx, r.key, options.Group, msg.ID).Err(); err != nil {
			log.Printf("[redis-streams] ack dead-letter message %s: %v", msg.ID, err)
		}
	}
}

func (r *RedisStreams) normalizeOptions(topic string, options SubscribeOptions) SubscribeOptions {
	if options.Group == "" {
		options.Group = defaultGroupName(r.group, topic)
	}
	if options.Consumer == "" {
		options.Consumer = defaultConsumerName(options.Group, topic)
	}
	if options.MaxRetries <= 0 {
		options.MaxRetries = 5
	}
	if options.ClaimMinIdle <= 0 {
		options.ClaimMinIdle = 30 * time.Second
	}
	if options.DeadLetterKey == "" {
		options.DeadLetterKey = r.key + ":dead"
	}
	return options
}

func (r *RedisStreams) ensureGroup(ctx context.Context, group string) error {
	err := r.client.XGroupCreateMkStream(ctx, r.key, group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (r *RedisStreams) trackGroup(group string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[group] = struct{}{}
}

func (r *RedisStreams) trackedGroups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.groups) == 0 {
		return []string{r.group}
	}
	groups := make([]string, 0, len(r.groups))
	for group := range r.groups {
		groups = append(groups, group)
	}
	return groups
}

func defaultGroupName(group, topic string) string {
	topic = strings.NewReplacer(".", "-", ":", "-").Replace(topic)
	return fmt.Sprintf("%s:%s", group, topic)
}

func defaultConsumerName(group, topic string) string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	topic = strings.NewReplacer(".", "-", ":", "-").Replace(topic)
	return fmt.Sprintf("%s-%s-%s-%d", group, topic, host, os.Getpid())
}

// Close 关闭
func (r *RedisStreams) Close() error {
	return nil
}

// GetPending 获取待处理消息数
func (r *RedisStreams) GetPending() (int64, error) {
	ctx := context.Background()
	var total int64
	for _, group := range r.trackedGroups() {
		info, err := r.client.XPending(ctx, r.key, group).Result()
		if err != nil {
			if strings.Contains(strings.ToUpper(err.Error()), "NOGROUP") {
				continue
			}
			return 0, err
		}
		total += info.Count
	}
	return total, nil
}

// Heal 修复消费者组（将所有pending消息重新激活）
func (r *RedisStreams) Heal() error {
	ctx := context.Background()
	for _, group := range r.trackedGroups() {
		pending, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: r.key,
			Group:  group,
			Start:  "-",
			End:    "+",
			Count:  100,
		}).Result()
		if err != nil {
			if strings.Contains(strings.ToUpper(err.Error()), "NOGROUP") {
				continue
			}
			return err
		}
		for _, p := range pending {
			r.client.XClaim(ctx, &redis.XClaimArgs{
				Stream:   r.key,
				Group:    group,
				Consumer: defaultConsumerName(group, "heal"),
				MinIdle:  0,
				Messages: []string{p.ID},
			})
		}
	}
	return nil
}
