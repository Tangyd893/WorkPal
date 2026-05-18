package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/docs/model"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListDocuments(ctx context.Context, projectID *int64, parentID *int64) ([]*model.Document, error) {
	var docs []*model.Document
	q := r.db.WithContext(ctx)
	if projectID != nil {
		q = q.Where("project_id = ?", *projectID)
	}
	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	} else {
		q = q.Where("parent_id IS NULL")
	}
	err := q.Order("is_folder DESC, sort_order ASC, created_at DESC").Find(&docs).Error
	return docs, err
}

func (r *Repo) GetDocument(ctx context.Context, docID int64) (*model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).Where("id = ?", docID).First(&doc).Error
	return &doc, err
}

func (r *Repo) CreateDocument(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *Repo) UpdateDocument(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *Repo) DeleteDocument(ctx context.Context, docID int64) error {
	return r.db.WithContext(ctx).Where("id = ?", docID).Delete(&model.Document{}).Error
}

func (r *Repo) GetLatestRevision(ctx context.Context, docID int64) (*model.DocumentRevision, error) {
	var rev model.DocumentRevision
	err := r.db.WithContext(ctx).
		Where("document_id = ?", docID).
		Order("version DESC").
		First(&rev).Error
	return &rev, err
}

func (r *Repo) CreateRevision(ctx context.Context, rev *model.DocumentRevision) error {
	return r.db.WithContext(ctx).Create(rev).Error
}

func (r *Repo) ListRevisions(ctx context.Context, docID int64) ([]*model.DocumentRevision, error) {
	var revs []*model.DocumentRevision
	err := r.db.WithContext(ctx).
		Where("document_id = ?", docID).
		Order("version DESC").
		Find(&revs).Error
	return revs, err
}
