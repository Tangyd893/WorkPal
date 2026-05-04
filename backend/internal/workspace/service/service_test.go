package service

import (
	"context"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/workspace/model"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	tasks  []*model.Task
	events []*model.ScheduleEvent
	nextID int64

	getTaskErr     error
	getEventErr    error
	saveTaskErr    error
	saveEventErr   error
	createTaskErr  error
	createEventErr error
	deleteTaskErr  error
	deleteEventErr error
	listTaskErr    error
	listEventErr   error
}

func newMockRepo() *mockRepo {
	return &mockRepo{nextID: 1}
}

func (m *mockRepo) ListTasks(ctx context.Context, userID int64) ([]*model.Task, error) {
	if m.listTaskErr != nil {
		return nil, m.listTaskErr
	}
	var out []*model.Task
	for _, t := range m.tasks {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *mockRepo) CreateTask(ctx context.Context, task *model.Task) error {
	if m.createTaskErr != nil {
		return m.createTaskErr
	}
	task.ID = m.nextID
	m.nextID++
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *mockRepo) GetTask(ctx context.Context, userID, taskID int64) (*model.Task, error) {
	if m.getTaskErr != nil {
		return nil, m.getTaskErr
	}
	for _, t := range m.tasks {
		if t.ID == taskID && t.UserID == userID {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockRepo) SaveTask(ctx context.Context, task *model.Task) error {
	if m.saveTaskErr != nil {
		return m.saveTaskErr
	}
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			return nil
		}
	}
	return nil
}

func (m *mockRepo) DeleteTask(ctx context.Context, userID, taskID int64) error {
	if m.deleteTaskErr != nil {
		return m.deleteTaskErr
	}
	for i, t := range m.tasks {
		if t.ID == taskID && t.UserID == userID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockRepo) ListEvents(ctx context.Context, userID int64) ([]*model.ScheduleEvent, error) {
	if m.listEventErr != nil {
		return nil, m.listEventErr
	}
	var out []*model.ScheduleEvent
	for _, e := range m.events {
		if e.UserID == userID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *mockRepo) CreateEvent(ctx context.Context, event *model.ScheduleEvent) error {
	if m.createEventErr != nil {
		return m.createEventErr
	}
	event.ID = m.nextID
	m.nextID++
	m.events = append(m.events, event)
	return nil
}

func (m *mockRepo) GetEvent(ctx context.Context, userID, eventID int64) (*model.ScheduleEvent, error) {
	if m.getEventErr != nil {
		return nil, m.getEventErr
	}
	for _, e := range m.events {
		if e.ID == eventID && e.UserID == userID {
			return e, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockRepo) SaveEvent(ctx context.Context, event *model.ScheduleEvent) error {
	if m.saveEventErr != nil {
		return m.saveEventErr
	}
	for i, e := range m.events {
		if e.ID == event.ID {
			m.events[i] = event
			return nil
		}
	}
	return nil
}

func (m *mockRepo) DeleteEvent(ctx context.Context, userID, eventID int64) error {
	if m.deleteEventErr != nil {
		return m.deleteEventErr
	}
	for i, e := range m.events {
		if e.ID == eventID && e.UserID == userID {
			m.events = append(m.events[:i], m.events[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestNewService(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	assert.NotNil(t, svc)
}

func TestListTasks(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10, Title: "Task A", Status: "planned"},
		{ID: 2, UserID: 10, Title: "Task B", Status: "in-progress"},
		{ID: 3, UserID: 20, Title: "Task C", Status: "done"},
	}
	svc := NewService(repo)

	tasks, err := svc.ListTasks(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "ws-1", tasks[0].ID)
	assert.Equal(t, "ws-2", tasks[1].ID)
}

func TestListTasks_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tasks, err := svc.ListTasks(context.Background(), 99)
	assert.NoError(t, err)
	assert.Len(t, tasks, 0)
}

func TestListTasks_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.listTaskErr = assert.AnError
	svc := NewService(repo)

	_, err := svc.ListTasks(context.Background(), 10)
	assert.Error(t, err)
}

func TestCreateTask(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	task, err := svc.CreateTask(context.Background(), 10, TaskInput{
		Title:         "New Task",
		Summary:       "Task summary",
		Project:       "Alpha",
		OwnerUsername: "alice",
		Teammates:     []string{"bob", "charlie"},
		DueDate:       "2026-06-01",
		Priority:      "high",
	})
	assert.NoError(t, err)
	assert.Equal(t, "ws-1", task.ID)
	assert.Equal(t, "New Task", task.Title)
	assert.Equal(t, "Alpha", task.Project)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, "planned", task.Status)
	assert.Equal(t, 0, task.SharedCount)
	assert.Equal(t, "custom", task.Source)

	assert.Len(t, repo.tasks, 1)
	assert.Equal(t, "high", repo.tasks[0].Priority)
}

func TestCreateTask_Defaults(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	task, err := svc.CreateTask(context.Background(), 10, TaskInput{
		Title: "Minimal Task",
	})
	assert.NoError(t, err)
	assert.Equal(t, "General", task.Project)
	assert.Equal(t, "medium", task.Priority)
	assert.Equal(t, "planned", task.Status)
}

func TestCreateTask_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.createTaskErr = assert.AnError
	svc := NewService(repo)

	_, err := svc.CreateTask(context.Background(), 10, TaskInput{
		Title: "Will Fail",
	})
	assert.Error(t, err)
}

func TestUpdateTaskStatus(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10, Title: "Task", Status: "planned"},
	}
	svc := NewService(repo)

	task, err := svc.UpdateTaskStatus(context.Background(), 10, 1, "in-progress")
	assert.NoError(t, err)
	assert.Equal(t, "in-progress", task.Status)
}

func TestUpdateTaskStatus_EmptyStatus_KeepsPlanned(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10, Title: "Task", Status: "review"},
	}
	svc := NewService(repo)

	task, err := svc.UpdateTaskStatus(context.Background(), 10, 1, "")
	assert.NoError(t, err)
	assert.Equal(t, "planned", task.Status)
}

func TestUpdateTaskStatus_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.UpdateTaskStatus(context.Background(), 10, 999, "done")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestUpdateTaskStatus_WrongUser(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10, Title: "Task"},
	}
	svc := NewService(repo)

	_, err := svc.UpdateTaskStatus(context.Background(), 99, 1, "done")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestShareTask(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10, Title: "Task", SharedCount: 2},
	}
	svc := NewService(repo)

	task, err := svc.ShareTask(context.Background(), 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, task.SharedCount)
}

func TestShareTask_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.ShareTask(context.Background(), 10, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteTask(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10},
	}
	svc := NewService(repo)

	err := svc.DeleteTask(context.Background(), 10, 1)
	assert.NoError(t, err)
	assert.Len(t, repo.tasks, 0)
}

func TestDeleteTask_WrongUser_NoError(t *testing.T) {
	repo := newMockRepo()
	repo.tasks = []*model.Task{
		{ID: 1, UserID: 10},
	}
	svc := NewService(repo)

	err := svc.DeleteTask(context.Background(), 99, 1)
	assert.NoError(t, err)
	assert.Len(t, repo.tasks, 1)
}

func TestListEvents(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	repo := newMockRepo()
	repo.events = []*model.ScheduleEvent{
		{ID: 1, UserID: 10, Title: "Meeting A", StartsAt: now},
		{ID: 2, UserID: 10, Title: "Meeting B", StartsAt: now.Add(time.Hour)},
		{ID: 3, UserID: 20, Title: "Meeting C", StartsAt: now},
	}
	svc := NewService(repo)

	events, err := svc.ListEvents(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "ws-1", events[0].ID)
	assert.Equal(t, "ws-2", events[1].ID)
}

func TestListEvents_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	events, err := svc.ListEvents(context.Background(), 99)
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestListEvents_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.listEventErr = assert.AnError
	svc := NewService(repo)

	_, err := svc.ListEvents(context.Background(), 10)
	assert.Error(t, err)
}

func TestCreateEvent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	event, err := svc.CreateEvent(context.Background(), 10, EventInput{
		Title:           "Standup",
		Detail:          "Daily standup meeting",
		OwnerUsername:   "alice",
		StartsAt:        "2026-06-01T09:00:00Z",
		DurationMinutes: 15,
		Attendees:       []string{"bob", "charlie"},
		Room:            "Room A",
	})
	assert.NoError(t, err)
	assert.Equal(t, "ws-1", event.ID)
	assert.Equal(t, "Standup", event.Title)
	assert.Equal(t, 15, event.DurationMinutes)
	assert.Equal(t, "Room A", event.Room)
	assert.Equal(t, 0, event.SharedCount)
	assert.Equal(t, "custom", event.Source)

	assert.Len(t, repo.events, 1)
}

func TestCreateEvent_InvalidDate_UsesNow(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	event, err := svc.CreateEvent(context.Background(), 10, EventInput{
		Title:    "Bad Date",
		StartsAt: "not-a-date",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, event.StartsAt)
}

func TestCreateEvent_ZeroDuration_UsesDefault(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	event, err := svc.CreateEvent(context.Background(), 10, EventInput{
		Title:           "No Duration",
		StartsAt:        "2026-06-01T09:00:00Z",
		DurationMinutes: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, 30, event.DurationMinutes)
}

func TestCreateEvent_NegativeDuration_UsesDefault(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	event, err := svc.CreateEvent(context.Background(), 10, EventInput{
		Title:           "Bad Duration",
		StartsAt:        "2026-06-01T09:00:00Z",
		DurationMinutes: -5,
	})
	assert.NoError(t, err)
	assert.Equal(t, 30, event.DurationMinutes)
}

func TestCreateEvent_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.createEventErr = assert.AnError
	svc := NewService(repo)

	_, err := svc.CreateEvent(context.Background(), 10, EventInput{
		Title:    "Will Fail",
		StartsAt: "2026-06-01T09:00:00Z",
	})
	assert.Error(t, err)
}

func TestShareEvent(t *testing.T) {
	repo := newMockRepo()
	repo.events = []*model.ScheduleEvent{
		{ID: 1, UserID: 10, Title: "Event", SharedCount: 1},
	}
	svc := NewService(repo)

	event, err := svc.ShareEvent(context.Background(), 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, event.SharedCount)
}

func TestShareEvent_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.ShareEvent(context.Background(), 10, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteEvent(t *testing.T) {
	repo := newMockRepo()
	repo.events = []*model.ScheduleEvent{
		{ID: 1, UserID: 10},
	}
	svc := NewService(repo)

	err := svc.DeleteEvent(context.Background(), 10, 1)
	assert.NoError(t, err)
	assert.Len(t, repo.events, 0)
}

func TestDeleteEvent_WrongUser_NoError(t *testing.T) {
	repo := newMockRepo()
	repo.events = []*model.ScheduleEvent{
		{ID: 1, UserID: 10},
	}
	svc := NewService(repo)

	err := svc.DeleteEvent(context.Background(), 99, 1)
	assert.NoError(t, err)
	assert.Len(t, repo.events, 1)
}

func TestFormatID(t *testing.T) {
	assert.Equal(t, "ws-1", formatID(1))
	assert.Equal(t, "ws-999", formatID(999))
}

func TestParseID(t *testing.T) {
	id, err := ParseID("ws-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestParseID_NoPrefix(t *testing.T) {
	id, err := ParseID("42")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestParseID_Invalid(t *testing.T) {
	_, err := ParseID("ws-abc")
	assert.Error(t, err)
}

func TestMarshalStringSlice(t *testing.T) {
	result := marshalStringSlice([]string{"a", "b", "c"})
	assert.Equal(t, `["a","b","c"]`, result)
}

func TestMarshalStringSlice_Empty(t *testing.T) {
	result := marshalStringSlice([]string{})
	assert.Equal(t, `[]`, result)
}

func TestUnmarshalStringSlice(t *testing.T) {
	result := unmarshalStringSlice(`["x","y"]`)
	assert.Equal(t, []string{"x", "y"}, result)
}

func TestUnmarshalStringSlice_Empty(t *testing.T) {
	assert.Equal(t, []string{}, unmarshalStringSlice(""))
	assert.Equal(t, []string{}, unmarshalStringSlice("invalid-json"))
}

func TestFallback(t *testing.T) {
	assert.Equal(t, "hello", fallback("hello", "world"))
	assert.Equal(t, "world", fallback("", "world"))
}

func TestTaskDTO_Teammates(t *testing.T) {
	task := &model.Task{
		ID:        1,
		Title:     "T",
		Teammates: `["alice","bob"]`,
		Status:    "done",
		Priority:  "high",
	}
	dto := taskDTO(task)
	assert.Equal(t, []string{"alice", "bob"}, dto.Teammates)
	assert.Equal(t, "done", dto.Status)
	assert.Equal(t, "high", dto.Priority)
}

func TestEventDTO_Attendees(t *testing.T) {
	now := time.Now()
	event := &model.ScheduleEvent{
		ID:              1,
		Title:           "E",
		StartsAt:        now,
		DurationMinutes: 60,
		Attendees:       `["x","y"]`,
		Room:            "R1",
	}
	dto := eventDTO(event)
	assert.Equal(t, []string{"x", "y"}, dto.Attendees)
	assert.Equal(t, "R1", dto.Room)
	assert.Equal(t, 60, dto.DurationMinutes)
	assert.Equal(t, now.Format(time.RFC3339), dto.StartsAt)
}
