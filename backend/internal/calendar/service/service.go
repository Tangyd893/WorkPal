package service

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/calendar/model"
)

type Repository interface {
	ListEvents(ctx context.Context, from, to *time.Time, organizerID *int64) ([]*model.CalendarEvent, error)
	GetEvent(ctx context.Context, eventID int64) (*model.CalendarEvent, error)
	CreateEvent(ctx context.Context, event *model.CalendarEvent) error
	UpdateEvent(ctx context.Context, event *model.CalendarEvent) error
	DeleteEvent(ctx context.Context, eventID int64) error
	ListAttendees(ctx context.Context, eventID int64) ([]*model.CalendarAttendee, error)
	AddAttendee(ctx context.Context, att *model.CalendarAttendee) error
	RemoveAttendee(ctx context.Context, eventID, userID int64) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

type EventDTO struct {
	ID             int64           `json:"id"`
	ProjectID      *int64          `json:"project_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	StartsAt       string          `json:"starts_at"`
	EndsAt         string          `json:"ends_at"`
	IsAllDay       bool            `json:"is_all_day"`
	Location       string          `json:"location"`
	OrganizerID    int64           `json:"organizer_id"`
	RecurrenceRule string          `json:"recurrence_rule"`
	Attendees      []AttendeeDTO   `json:"attendees,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

type EventInput struct {
	ProjectID   *int64 `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	IsAllDay    bool   `json:"is_all_day"`
	Location    string `json:"location"`
	AttendeeIDs []int64 `json:"attendee_ids"`
}

type AttendeeDTO struct {
	ID      int64  `json:"id"`
	UserID  int64  `json:"user_id"`
	Status  string `json:"status"`
}

func (s *Service) ListEvents(ctx context.Context, from, to *string, organizerID *int64) ([]EventDTO, error) {
	var f, t *time.Time
	if from != nil {
		parsed, _ := time.Parse(time.RFC3339, *from)
		f = &parsed
	}
	if to != nil {
		parsed, _ := time.Parse(time.RFC3339, *to)
		t = &parsed
	}
	events, err := s.repo.ListEvents(ctx, f, t, organizerID)
	if err != nil {
		return nil, err
	}
	out := make([]EventDTO, 0, len(events))
	for _, ev := range events {
		dto := eventToDTO(ev)
		atts, _ := s.repo.ListAttendees(ctx, ev.ID)
		for _, a := range atts {
			dto.Attendees = append(dto.Attendees, AttendeeDTO{ID: a.ID, UserID: a.UserID, Status: a.Status})
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID int64) (EventDTO, error) {
	ev, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return EventDTO{}, err
	}
	dto := eventToDTO(ev)
	atts, _ := s.repo.ListAttendees(ctx, ev.ID)
	for _, a := range atts {
		dto.Attendees = append(dto.Attendees, AttendeeDTO{ID: a.ID, UserID: a.UserID, Status: a.Status})
	}
	return dto, nil
}

func (s *Service) CreateEvent(ctx context.Context, organizerID int64, input EventInput) (EventDTO, error) {
	startsAt, _ := time.Parse(time.RFC3339, input.StartsAt)
	endsAt, _ := time.Parse(time.RFC3339, input.EndsAt)
	ev := &model.CalendarEvent{
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Description: input.Description,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		IsAllDay:    input.IsAllDay,
		Location:    input.Location,
		OrganizerID: organizerID,
	}
	if err := s.repo.CreateEvent(ctx, ev); err != nil {
		return EventDTO{}, err
	}
	for _, uid := range input.AttendeeIDs {
		_ = s.repo.AddAttendee(ctx, &model.CalendarAttendee{EventID: ev.ID, UserID: uid, Status: "pending"})
	}
	return s.GetEvent(ctx, ev.ID)
}

func (s *Service) UpdateEvent(ctx context.Context, eventID int64, input EventInput) (EventDTO, error) {
	ev, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return EventDTO{}, err
	}
	if input.Title != "" { ev.Title = input.Title }
	if input.Description != "" { ev.Description = input.Description }
	if input.StartsAt != "" {
		startsAt, _ := time.Parse(time.RFC3339, input.StartsAt)
		ev.StartsAt = startsAt
	}
	if input.EndsAt != "" {
		endsAt, _ := time.Parse(time.RFC3339, input.EndsAt)
		ev.EndsAt = endsAt
	}
	ev.Location = input.Location
	ev.IsAllDay = input.IsAllDay
	if err := s.repo.UpdateEvent(ctx, ev); err != nil {
		return EventDTO{}, err
	}
	return s.GetEvent(ctx, ev.ID)
}

func (s *Service) DeleteEvent(ctx context.Context, eventID int64) error {
	return s.repo.DeleteEvent(ctx, eventID)
}

func eventToDTO(ev *model.CalendarEvent) EventDTO {
	return EventDTO{
		ID:             ev.ID,
		ProjectID:      ev.ProjectID,
		Title:          ev.Title,
		Description:    ev.Description,
		StartsAt:       ev.StartsAt.Format(time.RFC3339),
		EndsAt:         ev.EndsAt.Format(time.RFC3339),
		IsAllDay:       ev.IsAllDay,
		Location:       ev.Location,
		OrganizerID:    ev.OrganizerID,
		RecurrenceRule: ev.RecurrenceRule,
		CreatedAt:      ev.CreatedAt.Format(time.RFC3339),
	}
}
