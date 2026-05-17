package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/project/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListProjects(ctx context.Context) ([]*model.Project, error) {
	var projects []*model.Project
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

func (r *Repo) CreateProject(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *Repo) GetProject(ctx context.Context, projectID int64) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).
		Where("id = ?", projectID).
		First(&project).Error
	return &project, err
}

func (r *Repo) SaveProject(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *Repo) DeleteProject(ctx context.Context, projectID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", projectID).
		Delete(&model.Project{}).Error
}

func (r *Repo) NextIssueKeySeq(ctx context.Context, projectID int64) (int64, error) {
	project, err := r.GetProject(ctx, projectID)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Issue{}).
		Where("project_id = ?", projectID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	_ = project
	return count + 1, nil
}

func (r *Repo) ListIssues(ctx context.Context, projectID int64) ([]*model.Issue, error) {
	var issues []*model.Issue
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("sort_order ASC, created_at DESC").
		Find(&issues).Error
	return issues, err
}

func (r *Repo) CreateIssue(ctx context.Context, issue *model.Issue) error {
	return r.db.WithContext(ctx).Create(issue).Error
}

func (r *Repo) GetIssue(ctx context.Context, issueID int64) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.WithContext(ctx).
		Where("id = ?", issueID).
		First(&issue).Error
	return &issue, err
}

func (r *Repo) SaveIssue(ctx context.Context, issue *model.Issue) error {
	return r.db.WithContext(ctx).Save(issue).Error
}

func (r *Repo) DeleteIssue(ctx context.Context, issueID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", issueID).
		Delete(&model.Issue{}).Error
}

func (r *Repo) CreateChangelog(ctx context.Context, log *model.IssueChangelog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *Repo) ListChangelogs(ctx context.Context, issueID int64) ([]*model.IssueChangelog, error) {
	var logs []*model.IssueChangelog
	err := r.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

func (r *Repo) CreateAssociation(ctx context.Context, assoc *model.Association) error {
	return r.db.WithContext(ctx).Create(assoc).Error
}

func (r *Repo) ListAssociations(ctx context.Context, sourceType string, sourceID int64) ([]*model.Association, error) {
	var assocs []*model.Association
	err := r.db.WithContext(ctx).
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Find(&assocs).Error
	return assocs, err
}

func (r *Repo) ListIssueTypes(ctx context.Context, projectID int64) ([]*model.IssueType, error) {
	var types []*model.IssueType
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&types).Error
	return types, err
}
