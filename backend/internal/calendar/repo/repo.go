package repo

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/calendar/model"
	"gorm.io/gorm"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) ListEvents(ctx context.Context, from, to *time.Time, organizerID *int64) ([]*model.CalendarEvent, error) {
	var events []*model.CalendarEvent
	q := r.db.WithContext(ctx)
	if from != nil {
		q = q.Where("starts_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("starts_at <= ?", *to)
	}
	if organizerID != nil {
		q = q.Where("organizer_id = ?", *organizerID)
	}
	err := q.Order("starts_at ASC").Find(&events).Error
	return events, err
}

func (r *Repo) GetEvent(ctx context.Context, eventID int64) (*model.CalendarEvent, error) {
	var e model.CalendarEvent
	err := r.db.WithContext(ctx).Where("id = ?", eventID).First(&e).Error
	return &e, err
}

func (r *Repo) CreateEvent(ctx context.Context, event *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *Repo) UpdateEvent(ctx context.Context, event *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *Repo) DeleteEvent(ctx context.Context, eventID int64) error {
	return r.db.WithContext(ctx).Where("id = ?", eventID).Delete(&model.CalendarEvent{}).Error
}

func (r *Repo) ListAttendees(ctx context.Context, eventID int64) ([]*model.CalendarAttendee, error) {
	var atts []*model.CalendarAttendee
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&atts).Error
	return atts, err
}

func (r *Repo) AddAttendee(ctx context.Context, att *model.CalendarAttendee) error {
	return r.db.WithContext(ctx).Create(att).Error
}

func (r *Repo) RemoveAttendee(ctx context.Context, eventID, userID int64) error {
	return r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&model.CalendarAttendee{}).Error
}
