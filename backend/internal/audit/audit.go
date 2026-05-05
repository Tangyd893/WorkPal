package audit

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"not null;index" json:"user_id"`
	Action       string    `gorm:"size:128;not null;index" json:"action"`
	ResourceType string    `gorm:"size:128;not null;index" json:"resource_type"`
	ResourceID   string    `gorm:"size:128;not null;index" json:"resource_id"`
	IP           string    `gorm:"size:128" json:"ip"`
	Timestamp    time.Time `gorm:"not null;index" json:"timestamp"`
}

func (Log) TableName() string {
	return "audit_logs"
}

type Recorder struct {
	db *gorm.DB
}

func NewRecorder(db *gorm.DB) *Recorder {
	if db == nil {
		return nil
	}
	return &Recorder{db: db}
}

func (r *Recorder) Record(ctx context.Context, userID int64, action, resourceType, resourceID, ip string) {
	if r == nil || r.db == nil {
		return
	}
	entry := &Log{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IP:           ip,
		Timestamp:    time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		slog.WarnContext(ctx, "审计日志写入失败", "user_id", userID, "action", action, "resource_type", resourceType, "resource_id", resourceID, "error", err)
	}
}
