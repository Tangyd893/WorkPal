package repo

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxRepo struct {
	db *gorm.DB
}

func NewOutboxRepo(db *gorm.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) CreateWithTx(tx *gorm.DB, event *model.MessageOutbox) error {
	return tx.Create(event).Error
}

func (r *OutboxRepo) ClaimPending(ctx context.Context, limit int) ([]*model.MessageOutbox, error) {
	if limit <= 0 {
		limit = 20
	}

	now := time.Now()
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var events []*model.MessageOutbox
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND next_attempt_at <= ?", model.OutboxStatusPending, now).
		Order("id ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(events) == 0 {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return nil, nil
	}

	ids := make([]int64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
		event.Status = model.OutboxStatusPublishing
	}
	if err := tx.Model(&model.MessageOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"status":     model.OutboxStatusPublishing,
			"updated_at": now,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OutboxRepo) MarkDelivered(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.MessageOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.OutboxStatusDelivered,
			"delivered_at": &now,
			"last_error":   "",
			"updated_at":   now,
		}).Error
}

func (r *OutboxRepo) MarkRetry(ctx context.Context, id int64, retryCount int, lastError string, nextAttemptAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.MessageOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":          model.OutboxStatusPending,
			"retry_count":     retryCount,
			"last_error":      truncateError(lastError),
			"next_attempt_at": nextAttemptAt,
		}).Error
}

func (r *OutboxRepo) ResetPublishing(ctx context.Context, threshold time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.MessageOutbox{}).
		Where("status = ? AND updated_at < ?", model.OutboxStatusPublishing, threshold).
		Updates(map[string]any{
			"status": model.OutboxStatusPending,
		})
	return result.RowsAffected, result.Error
}

func truncateError(message string) string {
	const maxLen = 1024
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen]
}
