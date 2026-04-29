package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64          `gorm:"not null;index" json:"user_id"`
	Title         string         `gorm:"size:255;not null" json:"title"`
	Summary       string         `gorm:"type:text" json:"summary"`
	Project       string         `gorm:"size:128" json:"project"`
	OwnerUsername string         `gorm:"size:64;index" json:"ownerUsername"`
	Teammates     string         `gorm:"type:jsonb" json:"-"`
	DueDate       string         `gorm:"size:32" json:"dueDate"`
	Priority      string         `gorm:"size:32;default:medium" json:"priority"`
	Status        string         `gorm:"size:32;default:planned" json:"status"`
	SharedCount   int            `gorm:"default:0" json:"sharedCount"`
	Source        string         `gorm:"size:32;default:custom" json:"source"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ScheduleEvent struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64          `gorm:"not null;index" json:"user_id"`
	Title           string         `gorm:"size:255;not null" json:"title"`
	Detail          string         `gorm:"type:text" json:"detail"`
	OwnerUsername   string         `gorm:"size:64;index" json:"ownerUsername"`
	StartsAt        time.Time      `gorm:"not null;index" json:"startsAt"`
	DurationMinutes int            `gorm:"default:30" json:"durationMinutes"`
	Attendees       string         `gorm:"type:jsonb" json:"-"`
	Room            string         `gorm:"size:128" json:"room"`
	SharedCount     int            `gorm:"default:0" json:"sharedCount"`
	Source          string         `gorm:"size:32;default:custom" json:"source"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
