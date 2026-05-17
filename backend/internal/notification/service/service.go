package service

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/notification/model"
)

type Repository interface {
	List(ctx context.Context, userID int64) ([]*model.Notification, error)
	Create(ctx context.Context, n *model.Notification) error
	MarkRead(ctx context.Context, userID, notifID int64) error
	MarkAllRead(ctx context.Context, userID int64) error
	CountUnread(ctx context.Context, userID int64) (int64, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type NotificationDTO struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}

func (s *Service) List(ctx context.Context, userID int64) ([]NotificationDTO, error) {
	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]NotificationDTO, 0, len(items))
	for _, n := range items {
		out = append(out, toDTO(n))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, userID int64, typ, title, content, entityType, entityID string) error {
	n := &model.Notification{
		UserID:     userID,
		Type:       typ,
		Title:      title,
		Content:    content,
		EntityType: entityType,
		EntityID:   entityID,
	}
	return s.repo.Create(ctx, n)
}

func (s *Service) MarkRead(ctx context.Context, userID, notifID int64) error {
	return s.repo.MarkRead(ctx, userID, notifID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *Service) CountUnread(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}

func toDTO(n *model.Notification) NotificationDTO {
	return NotificationDTO{
		ID:         n.ID,
		Type:       n.Type,
		Title:      n.Title,
		Content:    n.Content,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		IsRead:     n.IsRead,
		CreatedAt:  n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
