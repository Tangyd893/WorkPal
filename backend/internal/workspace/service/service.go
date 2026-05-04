package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/workspace/model"
)

var ErrNotFound = errors.New("workspace item not found")

type Repository interface {
	ListTasks(ctx context.Context, userID int64) ([]*model.Task, error)
	CreateTask(ctx context.Context, task *model.Task) error
	GetTask(ctx context.Context, userID, taskID int64) (*model.Task, error)
	SaveTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, userID, taskID int64) error
	ListEvents(ctx context.Context, userID int64) ([]*model.ScheduleEvent, error)
	CreateEvent(ctx context.Context, event *model.ScheduleEvent) error
	GetEvent(ctx context.Context, userID, eventID int64) (*model.ScheduleEvent, error)
	SaveEvent(ctx context.Context, event *model.ScheduleEvent) error
	DeleteEvent(ctx context.Context, userID, eventID int64) error
}

type Service struct {
	repo Repository
}

type TaskDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Project       string   `json:"project"`
	OwnerUsername string   `json:"ownerUsername"`
	Teammates     []string `json:"teammates"`
	DueDate       string   `json:"dueDate"`
	Priority      string   `json:"priority"`
	Status        string   `json:"status"`
	SharedCount   int      `json:"sharedCount"`
	Source        string   `json:"source"`
}

type TaskInput struct {
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Project       string   `json:"project"`
	OwnerUsername string   `json:"ownerUsername"`
	Teammates     []string `json:"teammates"`
	DueDate       string   `json:"dueDate"`
	Priority      string   `json:"priority"`
}

type EventDTO struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Detail          string   `json:"detail"`
	OwnerUsername   string   `json:"ownerUsername"`
	StartsAt        string   `json:"startsAt"`
	DurationMinutes int      `json:"durationMinutes"`
	Attendees       []string `json:"attendees"`
	Room            string   `json:"room"`
	SharedCount     int      `json:"sharedCount"`
	Source          string   `json:"source"`
}

type EventInput struct {
	Title           string   `json:"title"`
	Detail          string   `json:"detail"`
	OwnerUsername   string   `json:"ownerUsername"`
	StartsAt        string   `json:"startsAt"`
	DurationMinutes int      `json:"durationMinutes"`
	Attendees       []string `json:"attendees"`
	Room            string   `json:"room"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTasks(ctx context.Context, userID int64) ([]TaskDTO, error) {
	tasks, err := s.repo.ListTasks(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]TaskDTO, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, taskDTO(task))
	}
	return out, nil
}

func (s *Service) CreateTask(ctx context.Context, userID int64, input TaskInput) (TaskDTO, error) {
	task := &model.Task{
		UserID:        userID,
		Title:         input.Title,
		Summary:       input.Summary,
		Project:       fallback(input.Project, "General"),
		OwnerUsername: input.OwnerUsername,
		Teammates:     marshalStringSlice(input.Teammates),
		DueDate:       input.DueDate,
		Priority:      fallback(input.Priority, "medium"),
		Status:        "planned",
		SharedCount:   0,
		Source:        "custom",
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return TaskDTO{}, err
	}
	return taskDTO(task), nil
}

func (s *Service) UpdateTaskStatus(ctx context.Context, userID, taskID int64, status string) (TaskDTO, error) {
	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return TaskDTO{}, ErrNotFound
	}
	task.Status = fallback(status, "planned")
	if err := s.repo.SaveTask(ctx, task); err != nil {
		return TaskDTO{}, err
	}
	return taskDTO(task), nil
}

func (s *Service) ShareTask(ctx context.Context, userID, taskID int64) (TaskDTO, error) {
	task, err := s.repo.GetTask(ctx, userID, taskID)
	if err != nil {
		return TaskDTO{}, ErrNotFound
	}
	task.SharedCount++
	if err := s.repo.SaveTask(ctx, task); err != nil {
		return TaskDTO{}, err
	}
	return taskDTO(task), nil
}

func (s *Service) DeleteTask(ctx context.Context, userID, taskID int64) error {
	return s.repo.DeleteTask(ctx, userID, taskID)
}

func (s *Service) ListEvents(ctx context.Context, userID int64) ([]EventDTO, error) {
	events, err := s.repo.ListEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]EventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, eventDTO(event))
	}
	return out, nil
}

func (s *Service) CreateEvent(ctx context.Context, userID int64, input EventInput) (EventDTO, error) {
	startsAt, err := time.Parse(time.RFC3339, input.StartsAt)
	if err != nil {
		startsAt = time.Now()
	}
	duration := input.DurationMinutes
	if duration <= 0 {
		duration = 30
	}
	event := &model.ScheduleEvent{
		UserID:          userID,
		Title:           input.Title,
		Detail:          input.Detail,
		OwnerUsername:   input.OwnerUsername,
		StartsAt:        startsAt,
		DurationMinutes: duration,
		Attendees:       marshalStringSlice(input.Attendees),
		Room:            input.Room,
		SharedCount:     0,
		Source:          "custom",
	}
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return EventDTO{}, err
	}
	return eventDTO(event), nil
}

func (s *Service) ShareEvent(ctx context.Context, userID, eventID int64) (EventDTO, error) {
	event, err := s.repo.GetEvent(ctx, userID, eventID)
	if err != nil {
		return EventDTO{}, ErrNotFound
	}
	event.SharedCount++
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		return EventDTO{}, err
	}
	return eventDTO(event), nil
}

func (s *Service) DeleteEvent(ctx context.Context, userID, eventID int64) error {
	return s.repo.DeleteEvent(ctx, userID, eventID)
}

func taskDTO(task *model.Task) TaskDTO {
	return TaskDTO{
		ID:            formatID(task.ID),
		Title:         task.Title,
		Summary:       task.Summary,
		Project:       task.Project,
		OwnerUsername: task.OwnerUsername,
		Teammates:     unmarshalStringSlice(task.Teammates),
		DueDate:       task.DueDate,
		Priority:      fallback(task.Priority, "medium"),
		Status:        fallback(task.Status, "planned"),
		SharedCount:   task.SharedCount,
		Source:        fallback(task.Source, "custom"),
	}
}

func eventDTO(event *model.ScheduleEvent) EventDTO {
	return EventDTO{
		ID:              formatID(event.ID),
		Title:           event.Title,
		Detail:          event.Detail,
		OwnerUsername:   event.OwnerUsername,
		StartsAt:        event.StartsAt.Format(time.RFC3339),
		DurationMinutes: event.DurationMinutes,
		Attendees:       unmarshalStringSlice(event.Attendees),
		Room:            event.Room,
		SharedCount:     event.SharedCount,
		Source:          fallback(event.Source, "custom"),
	}
}

func formatID(id int64) string {
	return "ws-" + strconv.FormatInt(id, 10)
}

func ParseID(id string) (int64, error) {
	id = strings.TrimPrefix(id, "ws-")
	return strconv.ParseInt(id, 10, 64)
}

func marshalStringSlice(value []string) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func unmarshalStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return []string{}
	}
	return out
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
