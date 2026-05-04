package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOutboxRepo struct {
	pending      []*model.MessageOutbox
	deliveredIDs []int64
	retries      map[int64]int
	resetCalls   int
	failClaim    error
}

func newMockOutboxRepo(events ...*model.MessageOutbox) *mockOutboxRepo {
	return &mockOutboxRepo{
		pending: events,
		retries: make(map[int64]int),
	}
}

func (r *mockOutboxRepo) ClaimPending(ctx context.Context, limit int) ([]*model.MessageOutbox, error) {
	if r.failClaim != nil {
		return nil, r.failClaim
	}
	if len(r.pending) == 0 {
		return nil, nil
	}
	if limit > len(r.pending) {
		limit = len(r.pending)
	}
	batch := make([]*model.MessageOutbox, limit)
	copy(batch, r.pending[:limit])
	r.pending = r.pending[limit:]
	return batch, nil
}

func (r *mockOutboxRepo) MarkDelivered(ctx context.Context, id int64) error {
	r.deliveredIDs = append(r.deliveredIDs, id)
	return nil
}

func (r *mockOutboxRepo) MarkRetry(ctx context.Context, id int64, retryCount int, lastError string, nextAttemptAt time.Time) error {
	r.retries[id] = retryCount
	return nil
}

func (r *mockOutboxRepo) ResetPublishing(ctx context.Context, threshold time.Time) (int64, error) {
	r.resetCalls++
	return 0, nil
}

type mockQueue struct {
	published []string
	failOnce  map[string]bool
}

func newMockQueue() *mockQueue {
	return &mockQueue{failOnce: make(map[string]bool)}
}

func (q *mockQueue) Publish(ctx context.Context, topic string, msg []byte) error {
	if q.failOnce[topic] {
		delete(q.failOnce, topic)
		return errors.New("temporary publish failure")
	}
	q.published = append(q.published, topic+":"+string(msg))
	return nil
}

func (q *mockQueue) Subscribe(topic string, handler func([]byte)) error {
	return nil
}

func (q *mockQueue) SubscribeWithOptions(topic string, options msgqueue.SubscribeOptions, handler msgqueue.Handler) error {
	return nil
}

func (q *mockQueue) Close() error {
	return nil
}

func TestOutboxPublisherPublishesPendingEvents(t *testing.T) {
	repo := newMockOutboxRepo(
		&model.MessageOutbox{ID: 1, Topic: "message.upserted", Payload: `{"id":1}`},
		&model.MessageOutbox{ID: 2, Topic: "message.deleted", Payload: `{"id":2}`},
	)
	queue := newMockQueue()
	publisher := NewOutboxPublisher(repo, queue)
	publisher.batchSize = 10

	err := publisher.PublishPending(context.Background())
	require.NoError(t, err)
	assert.Len(t, queue.published, 2)
	assert.Equal(t, []int64{1, 2}, repo.deliveredIDs)
	assert.Equal(t, 1, repo.resetCalls)
}

func TestOutboxPublisherMarksRetryOnFailure(t *testing.T) {
	repo := newMockOutboxRepo(
		&model.MessageOutbox{ID: 1, Topic: "message.upserted", Payload: `{"id":1}`},
	)
	queue := newMockQueue()
	queue.failOnce["message.upserted"] = true

	publisher := NewOutboxPublisher(repo, queue)
	err := publisher.PublishPending(context.Background())
	require.NoError(t, err)
	assert.Empty(t, repo.deliveredIDs)
	assert.Equal(t, 1, repo.retries[1])
}

func TestOutboxPublisherContinuesAfterRetriablePublishFailure(t *testing.T) {
	repo := newMockOutboxRepo(
		&model.MessageOutbox{ID: 1, Topic: "message.upserted", Payload: `{"id":1}`},
		&model.MessageOutbox{ID: 2, Topic: "message.deleted", Payload: `{"id":2}`},
	)
	queue := newMockQueue()
	queue.failOnce["message.upserted"] = true

	publisher := NewOutboxPublisher(repo, queue)
	publisher.batchSize = 10

	err := publisher.PublishPending(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, repo.retries[1])
	assert.Equal(t, []int64{2}, repo.deliveredIDs)
	assert.Equal(t, []string{`message.deleted:{"id":2}`}, queue.published)
}
