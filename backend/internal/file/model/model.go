package model

import (
	"time"
)

// File 文件元信息
type File struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"not null;index" json:"user_id"`
	ConvID      int64     `gorm:"index;default:0" json:"conv_id"` // 关联会话（可选）
	Name        string    `gorm:"size:256;not null" json:"name"`
	Key         string    `gorm:"size:512;not null" json:"key"` // 对象存储 key
	Size        int64     `gorm:"not null" json:"size"`         // bytes
	ContentType string    `gorm:"size:128" json:"content_type"`
	MimeType    string    `gorm:"size:128" json:"mime_type"`
	CreatedAt   time.Time `json:"created_at"`
}

func (File) TableName() string {
	return "files"
}
