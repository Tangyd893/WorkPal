package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/notification/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context, userID int64) ([]*model.Notification, error) {
	var items []*model.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&items).Error
	return items, err
}

func (r *Repo) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *Repo) MarkRead(ctx context.Context, userID, notifID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true).Error
}

func (r *Repo) MarkAllRead(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true).Error
}

func (r *Repo) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}
