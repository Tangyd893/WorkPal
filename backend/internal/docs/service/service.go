package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/docs/model"
)

var ErrNotFound = errors.New("document not found")

type Repository interface {
	ListDocuments(ctx context.Context, projectID *int64, parentID *int64) ([]*model.Document, error)
	GetDocument(ctx context.Context, docID int64) (*model.Document, error)
	CreateDocument(ctx context.Context, doc *model.Document) error
	UpdateDocument(ctx context.Context, doc *model.Document) error
	DeleteDocument(ctx context.Context, docID int64) error
	GetLatestRevision(ctx context.Context, docID int64) (*model.DocumentRevision, error)
	CreateRevision(ctx context.Context, rev *model.DocumentRevision) error
	ListRevisions(ctx context.Context, docID int64) ([]*model.DocumentRevision, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type DocumentDTO struct {
	ID        int64  `json:"id"`
	ProjectID *int64 `json:"project_id"`
	ParentID  *int64 `json:"parent_id"`
	Title     string `json:"title"`
	CreatedBy int64  `json:"created_by"`
	UpdatedBy int64  `json:"updated_by"`
	IsFolder  bool   `json:"is_folder"`
	SortOrder int    `json:"sort_order"`
	Content   string `json:"content,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type DocumentInput struct {
	ProjectID *int64 `json:"project_id"`
	ParentID  *int64 `json:"parent_id"`
	Title     string `json:"title"`
	IsFolder  bool   `json:"is_folder"`
	Content   string `json:"content"`
}

type RevisionDTO struct {
	ID         int64  `json:"id"`
	DocumentID int64  `json:"document_id"`
	Version    int    `json:"version"`
	Content    string `json:"content"`
	CreatedBy  int64  `json:"created_by"`
	CreatedAt  string `json:"created_at"`
}

func (s *Service) ListDocuments(ctx context.Context, projectID, parentID *int64) ([]DocumentDTO, error) {
	docs, err := s.repo.ListDocuments(ctx, projectID, parentID)
	if err != nil {
		return nil, err
	}
	out := make([]DocumentDTO, 0, len(docs))
	for _, d := range docs {
		out = append(out, DocumentDTO{
			ID:        d.ID,
			ProjectID: d.ProjectID,
			ParentID:  d.ParentID,
			Title:     d.Title,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
			IsFolder:  d.IsFolder,
			SortOrder: d.SortOrder,
			CreatedAt: d.CreatedAt.Format(time.RFC3339),
			UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *Service) GetDocument(ctx context.Context, docID int64) (DocumentDTO, error) {
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return DocumentDTO{}, ErrNotFound
	}
	dto := DocumentDTO{
		ID:        doc.ID,
		ProjectID: doc.ProjectID,
		ParentID:  doc.ParentID,
		Title:     doc.Title,
		CreatedBy: doc.CreatedBy,
		UpdatedBy: doc.UpdatedBy,
		IsFolder:  doc.IsFolder,
		SortOrder: doc.SortOrder,
		CreatedAt: doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt: doc.UpdatedAt.Format(time.RFC3339),
	}
	rev, err := s.repo.GetLatestRevision(ctx, docID)
	if err == nil {
		dto.Content = rev.Content
	}
	return dto, nil
}

func (s *Service) CreateDocument(ctx context.Context, userID int64, input DocumentInput) (DocumentDTO, error) {
	doc := &model.Document{
		ProjectID: input.ProjectID,
		ParentID:  input.ParentID,
		Title:     input.Title,
		CreatedBy: userID,
		UpdatedBy: userID,
		IsFolder:  input.IsFolder,
	}
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return DocumentDTO{}, err
	}
	if !input.IsFolder {
		rev := &model.DocumentRevision{
			DocumentID: doc.ID,
			Version:    1,
			Content:    input.Content,
			CreatedBy:  userID,
		}
		_ = s.repo.CreateRevision(ctx, rev)
	}
	return s.GetDocument(ctx, doc.ID)
}

func (s *Service) UpdateDocument(ctx context.Context, docID int64, userID int64, input DocumentInput) (DocumentDTO, error) {
	doc, err := s.repo.GetDocument(ctx, docID)
	if err != nil {
		return DocumentDTO{}, ErrNotFound
	}
	if input.Title != "" {
		doc.Title = input.Title
	}
	doc.UpdatedBy = userID
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return DocumentDTO{}, err
	}
	if input.Content != "" && !doc.IsFolder {
		latest, _ := s.repo.GetLatestRevision(ctx, docID)
		nextVersion := 1
		if latest != nil {
			nextVersion = latest.Version + 1
		}
		rev := &model.DocumentRevision{
			DocumentID: docID,
			Version:    nextVersion,
			Content:    input.Content,
			CreatedBy:  userID,
		}
		if err := s.repo.CreateRevision(ctx, rev); err != nil {
			return DocumentDTO{}, fmt.Errorf("创建文档版本失败: %w", err)
		}
	}
	return s.GetDocument(ctx, docID)
}

func (s *Service) DeleteDocument(ctx context.Context, docID int64) error {
	return s.repo.DeleteDocument(ctx, docID)
}

func (s *Service) ListRevisions(ctx context.Context, docID int64) ([]RevisionDTO, error) {
	revs, err := s.repo.ListRevisions(ctx, docID)
	if err != nil {
		return nil, err
	}
	out := make([]RevisionDTO, 0, len(revs))
	for _, r := range revs {
		out = append(out, RevisionDTO{
			ID:         r.ID,
			DocumentID: r.DocumentID,
			Version:    r.Version,
			Content:    r.Content,
			CreatedBy:  r.CreatedBy,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}
