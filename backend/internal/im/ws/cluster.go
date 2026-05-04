package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultClusterChannel = "workpal:im:fanout"

type ClusterBroadcaster struct {
	client     *redis.Client
	hub        *Hub
	channel    string
	instanceID string
}

type clusterEvent struct {
	Kind      string  `json:"kind"`
	Origin    string  `json:"origin"`
	ConvID    int64   `json:"conv_id,omitempty"`
	FromID    int64   `json:"from_id,omitempty"`
	UserIDs   []int64 `json:"user_ids,omitempty"`
	Content   []byte  `json:"content"`
	CreatedAt string  `json:"created_at"`
}

func NewClusterBroadcaster(client *redis.Client, hub *Hub, channel string) *ClusterBroadcaster {
	if channel == "" {
		channel = defaultClusterChannel
	}
	host, _ := os.Hostname()
	if host == "" {
		host = "localhost"
	}
	return &ClusterBroadcaster{
		client:     client,
		hub:        hub,
		channel:    channel,
		instanceID: fmt.Sprintf("%s-%d", host, time.Now().UnixNano()),
	}
}

func (b *ClusterBroadcaster) BroadcastUsers(ctx context.Context, userIDs []int64, content []byte) error {
	if b.hub != nil {
		b.hub.SendToUsers(userIDs, content)
	}
	if b.client == nil || len(userIDs) == 0 {
		return nil
	}

	return b.publish(ctx, clusterEvent{
		Kind:      "users",
		Origin:    b.instanceID,
		UserIDs:   append([]int64(nil), userIDs...),
		Content:   append([]byte(nil), content...),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (b *ClusterBroadcaster) BroadcastRoom(ctx context.Context, convID int64, fromID int64, content []byte) error {
	if b.client == nil || convID <= 0 {
		return nil
	}

	return b.publish(ctx, clusterEvent{
		Kind:      "room",
		Origin:    b.instanceID,
		ConvID:    convID,
		FromID:    fromID,
		Content:   append([]byte(nil), content...),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (b *ClusterBroadcaster) Run(ctx context.Context) {
	if b.client == nil {
		return
	}

	pubsub := b.client.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf("[ws-cluster] subscribe %s: %v", b.channel, err)
		return
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-ch:
			if !ok {
				return
			}
			if err := b.handle(message.Payload); err != nil {
				log.Printf("[ws-cluster] handle event: %v", err)
			}
		}
	}
}

func (b *ClusterBroadcaster) publish(ctx context.Context, event clusterEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, b.channel, payload).Err()
}

func (b *ClusterBroadcaster) handle(payload string) error {
	var event clusterEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return err
	}
	if event.Origin == b.instanceID {
		return nil
	}

	switch event.Kind {
	case "users":
		if b.hub != nil {
			b.hub.SendToUsers(event.UserIDs, event.Content)
		}
	case "room":
		if b.hub != nil {
			b.hub.deliverMessage(&BroadcastMsg{
				ConvID:  event.ConvID,
				FromID:  event.FromID,
				Content: event.Content,
			})
		}
	default:
		return fmt.Errorf("unknown cluster event kind %q", event.Kind)
	}
	return nil
}
