package service

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"
)

type OutboxRepository interface {
	ClaimPending(ctx context.Context, limit int) ([]*model.MessageOutbox, error)
	MarkDelivered(ctx context.Context, id int64) error
	MarkRetry(ctx context.Context, id int64, retryCount int, lastError string, nextAttemptAt time.Time) error
	ResetPublishing(ctx context.Context, threshold time.Time) (int64, error)
}

type OutboxPublisher struct {
	repo              OutboxRepository
	queue             msgqueue.Interface
	interval          time.Duration
	batchSize         int
	maxBatchesPerTick int
	recoverAfter      time.Duration
}

func NewOutboxPublisher(repo OutboxRepository, queue msgqueue.Interface) *OutboxPublisher {
	return &OutboxPublisher{
		repo:              repo,
		queue:             queue,
		interval:          2 * time.Second,
		batchSize:         20,
		maxBatchesPerTick: 5,
		recoverAfter:      30 * time.Second,
	}
}

func (p *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	_ = p.PublishPending(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.PublishPending(ctx)
		}
	}
}

func (p *OutboxPublisher) PublishPending(ctx context.Context) error {
	if _, err := p.repo.ResetPublishing(ctx, time.Now().Add(-p.recoverAfter)); err != nil {
		return err
	}

	for i := 0; i < p.maxBatchesPerTick; i++ {
		events, err := p.repo.ClaimPending(ctx, p.batchSize)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		for _, event := range events {
			if err := p.publishOne(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *OutboxPublisher) publishOne(ctx context.Context, event *model.MessageOutbox) error {
	if err := p.queue.Publish(ctx, event.Topic, []byte(event.Payload)); err != nil {
		retryCount := event.RetryCount + 1
		return p.repo.MarkRetry(ctx, event.ID, retryCount, err.Error(), time.Now().Add(backoffForRetry(retryCount)))
	}
	return p.repo.MarkDelivered(ctx, event.ID)
}

func backoffForRetry(retryCount int) time.Duration {
	if retryCount < 1 {
		return time.Second
	}
	if retryCount > 5 {
		retryCount = 5
	}
	return time.Duration(1<<uint(retryCount-1)) * time.Second
}
