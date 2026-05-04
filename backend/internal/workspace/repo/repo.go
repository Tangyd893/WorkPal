package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/workspace/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListTasks(ctx context.Context, userID int64) ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

func (r *Repo) CreateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *Repo) GetTask(ctx context.Context, userID, taskID int64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", taskID, userID).
		First(&task).Error
	return &task, err
}

func (r *Repo) SaveTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *Repo) DeleteTask(ctx context.Context, userID, taskID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", taskID, userID).
		Delete(&model.Task{}).Error
}

func (r *Repo) ListEvents(ctx context.Context, userID int64) ([]*model.ScheduleEvent, error) {
	var events []*model.ScheduleEvent
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("starts_at ASC").
		Find(&events).Error
	return events, err
}

func (r *Repo) CreateEvent(ctx context.Context, event *model.ScheduleEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *Repo) GetEvent(ctx context.Context, userID, eventID int64) (*model.ScheduleEvent, error) {
	var event model.ScheduleEvent
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", eventID, userID).
		First(&event).Error
	return &event, err
}

func (r *Repo) SaveEvent(ctx context.Context, event *model.ScheduleEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *Repo) DeleteEvent(ctx context.Context, userID, eventID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", eventID, userID).
		Delete(&model.ScheduleEvent{}).Error
}
