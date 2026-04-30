package msgqueue

import (
	"context"
	"sync"
	"time"
)

// Message 消息结构
type Message struct {
	Topic string
	Data  []byte
}

type Handler func([]byte) error

type SubscribeOptions struct {
	Group         string
	Consumer      string
	MaxRetries    int64
	DeadLetterKey string
	ClaimMinIdle  time.Duration
}

// Interface 消息队列接口
type Interface interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Subscribe(topic string, handler func([]byte)) error
	SubscribeWithOptions(topic string, options SubscribeOptions, handler Handler) error
	Close() error
}

// memQueue 内存队列实现（单进程）
type memQueue struct {
	handlers map[string][]func([]byte)
	mu       sync.RWMutex
}

var _ Interface = (*memQueue)(nil)

// NewMemQueue 创建内存队列
func NewMemQueue() *memQueue {
	return &memQueue{
		handlers: make(map[string][]func([]byte)),
	}
}

// Publish 发布消息
func (m *memQueue) Publish(ctx context.Context, topic string, msg []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers[topic] {
		go h(msg)
	}
	return nil
}

// Subscribe 订阅主题
func (m *memQueue) Subscribe(topic string, handler func([]byte)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[topic] = append(m.handlers[topic], handler)
	return nil
}

func (m *memQueue) SubscribeWithOptions(topic string, options SubscribeOptions, handler Handler) error {
	return m.Subscribe(topic, func(data []byte) {
		_ = handler(data)
	})
}

// Close 关闭
func (m *memQueue) Close() error {
	return nil
}

var globalQueue Interface

// Init 初始化全局队列
func Init(q Interface) {
	globalQueue = q
}

// Publish 全局发布
func Publish(ctx context.Context, topic string, msg []byte) error {
	if globalQueue != nil {
		return globalQueue.Publish(ctx, topic, msg)
	}
	return nil
}

// Subscribe 全局订阅
func Subscribe(topic string, handler func([]byte)) error {
	if globalQueue != nil {
		return globalQueue.Subscribe(topic, handler)
	}
	return nil
}

func SubscribeWithOptions(topic string, options SubscribeOptions, handler Handler) error {
	if globalQueue != nil {
		return globalQueue.SubscribeWithOptions(topic, options, handler)
	}
	return nil
}
