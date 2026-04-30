package msgqueue

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStreamsUsesTopicScopedConsumerGroups(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	queue := NewRedisStreams(client, "workpal:test:messages", "workpal-search")
	upsertCh := make(chan string, 1)
	deleteCh := make(chan string, 1)

	require.NoError(t, queue.SubscribeWithOptions("message.upserted", SubscribeOptions{
		Consumer:     "test-upsert-indexer",
		ClaimMinIdle: time.Hour,
	}, func(data []byte) error {
		upsertCh <- string(data)
		return nil
	}))
	require.NoError(t, queue.SubscribeWithOptions("message.deleted", SubscribeOptions{
		Consumer:     "test-delete-indexer",
		ClaimMinIdle: time.Hour,
	}, func(data []byte) error {
		deleteCh <- string(data)
		return nil
	}))

	ctx := context.Background()
	require.NoError(t, queue.Publish(ctx, "message.upserted", []byte(`{"id":1}`)))
	require.NoError(t, queue.Publish(ctx, "message.deleted", []byte(`{"id":1}`)))

	assert.Equal(t, `{"id":1}`, receiveString(t, upsertCh))
	assert.Equal(t, `{"id":1}`, receiveString(t, deleteCh))

	assertPendingCount(t, client, "workpal:test:messages", defaultGroupName("workpal-search", "message.upserted"), 0)
	assertPendingCount(t, client, "workpal:test:messages", defaultGroupName("workpal-search", "message.deleted"), 0)
}

func receiveString(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for redis stream handler")
		return ""
	}
}

func assertPendingCount(t *testing.T, client *redis.Client, stream, group string, want int64) {
	t.Helper()
	require.Eventually(t, func() bool {
		info, err := client.XPending(context.Background(), stream, group).Result()
		return err == nil && info.Count == want
	}, 2*time.Second, 20*time.Millisecond)
}
