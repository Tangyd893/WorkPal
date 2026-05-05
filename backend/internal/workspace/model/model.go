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

type TaskSaga struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID       int64      `gorm:"not null;index" json:"task_id"`
	UserID       int64      `gorm:"not null;index" json:"user_id"`
	SagaType     string     `gorm:"size:64;not null;index" json:"saga_type"`
	Status       string     `gorm:"size:32;not null;index" json:"status"`
	CurrentStep  string     `gorm:"size:128;not null" json:"current_step"`
	Compensation string     `gorm:"type:text" json:"compensation"`
	LastError    string     `gorm:"type:text" json:"last_error"`
	NextRunAt    *time.Time `gorm:"index" json:"next_run_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (TaskSaga) TableName() string {
	return "task_sagas"
}
