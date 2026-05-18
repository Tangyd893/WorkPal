package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/approval/model"
	"gorm.io/gorm"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) ListTemplates(ctx context.Context, projectID *int64) ([]*model.ApprovalTemplate, error) {
	var ts []*model.ApprovalTemplate
	q := r.db.WithContext(ctx)
	if projectID != nil { q = q.Where("project_id = ?", *projectID) }
	err := q.Find(&ts).Error
	return ts, err
}

func (r *Repo) GetTemplate(ctx context.Context, id int64) (*model.ApprovalTemplate, error) {
	var t model.ApprovalTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *Repo) CreateTemplate(ctx context.Context, t *model.ApprovalTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repo) ListInstances(ctx context.Context, submitterID *int64, status *string) ([]*model.ApprovalInstance, error) {
	var insts []*model.ApprovalInstance
	q := r.db.WithContext(ctx)
	if submitterID != nil { q = q.Where("submitter_id = ?", *submitterID) }
	if status != nil { q = q.Where("status = ?", *status) }
	err := q.Order("created_at DESC").Find(&insts).Error
	return insts, err
}

func (r *Repo) GetInstance(ctx context.Context, id int64) (*model.ApprovalInstance, error) {
	var i model.ApprovalInstance
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&i).Error
	return &i, err
}

func (r *Repo) CreateInstance(ctx context.Context, i *model.ApprovalInstance) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *Repo) UpdateInstance(ctx context.Context, i *model.ApprovalInstance) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *Repo) CreateAction(ctx context.Context, a *model.ApprovalAction) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repo) ListActions(ctx context.Context, instanceID int64) ([]*model.ApprovalAction, error) {
	var acts []*model.ApprovalAction
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at ASC").Find(&acts).Error
	return acts, err
}
