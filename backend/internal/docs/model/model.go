package model

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   *int64         `gorm:"index" json:"project_id"`
	ParentID    *int64         `gorm:"index" json:"parent_id"`
	Title       string         `gorm:"size:500;not null" json:"title"`
	CreatedBy   int64          `gorm:"not null" json:"created_by"`
	UpdatedBy   int64          `gorm:"not null" json:"updated_by"`
	IsFolder    bool           `gorm:"default:false" json:"is_folder"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type DocumentRevision struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	DocumentID int64     `gorm:"not null;index" json:"document_id"`
	Version    int       `gorm:"not null" json:"version"`
	Content    string    `gorm:"type:jsonb;not null;default:'{}'" json:"content"`
	CreatedBy  int64     `gorm:"not null" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}
