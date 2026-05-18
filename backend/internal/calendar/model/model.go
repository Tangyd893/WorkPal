package model

import (
	"time"

	"gorm.io/gorm"
)

type CalendarEvent struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID      *int64         `gorm:"index" json:"project_id"`
	Title          string         `gorm:"size:500;not null" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	StartsAt       time.Time      `gorm:"not null" json:"starts_at"`
	EndsAt         time.Time      `gorm:"not null" json:"ends_at"`
	IsAllDay       bool           `gorm:"default:false" json:"is_all_day"`
	Location       string         `gorm:"size:255" json:"location"`
	OrganizerID    int64          `gorm:"not null;index" json:"organizer_id"`
	RecurrenceRule string         `gorm:"size:255" json:"recurrence_rule"`
	ParentEventID  *int64         `json:"parent_event_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type CalendarAttendee struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID   int64     `gorm:"not null;uniqueIndex:idx_attendee_event" json:"event_id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_attendee_event" json:"user_id"`
	Status    string    `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
