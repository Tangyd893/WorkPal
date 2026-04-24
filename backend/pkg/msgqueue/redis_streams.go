package msgqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStreams Redis Streams 实现
type RedisStreams struct {
	client *redis.Client
	key    string
	group  string
}

var _ Interface = (*RedisStreams)(nil)

// NewRedisStreams 创建 Redis Streams 队列
func NewRedisStreams(client *redis.Client, key, group string) *RedisStreams {
	return &RedisStreams{
		client: client,
		key:    key,
		group:  group,
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
	// 确保消费者组存在
	r.client.XGroupCreateMkStream(context.Background(), r.key, r.group, "0")

	go r.consume(topic, handler)
	return nil
}

func (r *RedisStreams) consume(topic string, handler func([]byte)) {
	ctx := context.Background()
	for {
		// 读取新消息
		streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    r.group,
			Consumer: fmt.Sprintf("consumer-%d", time.Now().UnixNano()),
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
				topicVal, ok := msg.Values["topic"].(string)
				if !ok {
					continue
				}
				if topicVal == topic {
					dataStr, ok := msg.Values["data"].(string)
					if !ok {
						continue
					}
					data := []byte(dataStr)
					go handler(data)
				}
				// ACK 消息
				r.client.XAck(ctx, r.key, r.group, msg.ID)
			}
		}
	}
}

// Close 关闭
func (r *RedisStreams) Close() error {
	return nil
}

// GetPending 获取待处理消息数
func (r *RedisStreams) GetPending() (int64, error) {
	ctx := context.Background()
	info, err := r.client.XPending(ctx, r.key, r.group).Result()
	if err != nil {
		return 0, err
	}
	return info.Count, nil
}

// Heal 修复消费者组（将所有pending消息重新激活）
func (r *RedisStreams) Heal() error {
	ctx := context.Background()
	// 读取所有待处理消息
	pending, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: r.key,
		Group:  r.group,
		Start:  "-+",
		Count:  100,
	}).Result()
	if err != nil {
		return err
	}
	// 重新claim这些消息（idle时间设为0，视为重新激活）
	for _, p := range pending {
		r.client.XClaim(ctx, &redis.XClaimArgs{
			Stream: r.key,
			Group:  r.group,
			MinIdle: 0,
			Messages: []string{p.ID},
		})
	}
	return nil
}
