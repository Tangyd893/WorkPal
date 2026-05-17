package model

import "time"

type Notification struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	Type       string    `gorm:"size:50;not null" json:"type"`
	Title      string    `gorm:"size:255;not null" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	EntityType string    `gorm:"size:50" json:"entity_type"`
	EntityID   string    `gorm:"size:100" json:"entity_id"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}
