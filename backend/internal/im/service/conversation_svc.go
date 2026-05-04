package service

import (
	"context"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
)

type ConversationRepository interface {
	Create(ctx context.Context, conv *model.Conversation) error
	GetByID(ctx context.Context, id int64) (*model.Conversation, error)
	FindPrivateConv(ctx context.Context, uid1, uid2 int64) (*model.Conversation, error)
	AddMember(ctx context.Context, m *model.ConversationMember) error
	IsMember(ctx context.Context, convID, userID int64) (bool, error)
	RemoveMember(ctx context.Context, convID, userID int64) error
	GetMembers(ctx context.Context, convID int64) ([]int64, error)
	ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, error)
	CountByUser(ctx context.Context, userID int64) (int64, error)
	Update(ctx context.Context, conv *model.Conversation) error
	Delete(ctx context.Context, convID int64) error
}

type ConversationService struct {
	convRepo ConversationRepository
}

func NewConversationService(convRepo ConversationRepository) *ConversationService {
	return &ConversationService{convRepo: convRepo}
}

func newConversationService(convRepo ConversationRepository) *ConversationService {
	return NewConversationService(convRepo)
}

func (s *ConversationService) CreatePrivateConv(ctx context.Context, userID, targetID int64) (*model.Conversation, error) {
	if userID == targetID {
		return nil, apperrors.ErrCannotChatWithSelf
	}

	conv, err := s.convRepo.FindPrivateConv(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}
	if conv != nil {
		return conv, nil
	}

	conv = &model.Conversation{
		Type:      model.ConversationTypePrivate,
		OwnerID:   userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}

	if err := s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: conv.ID, UserID: userID, JoinedAt: time.Now()}); err != nil {
		return nil, err
	}
	if err := s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: conv.ID, UserID: targetID, JoinedAt: time.Now()}); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *ConversationService) CreateGroup(ctx context.Context, name string, ownerID int64, memberIDs []int64) (*model.Conversation, error) {
	if name == "" {
		name = "群聊"
	}

	conv := &model.Conversation{
		Type:      model.ConversationTypeGroup,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}

	seenMembers := map[int64]struct{}{ownerID: {}}
	allMembers := []int64{ownerID}
	for _, memberID := range memberIDs {
		if memberID <= 0 {
			continue
		}
		if _, exists := seenMembers[memberID]; exists {
			continue
		}
		seenMembers[memberID] = struct{}{}
		allMembers = append(allMembers, memberID)
	}

	for _, userID := range allMembers {
		if err := s.convRepo.AddMember(ctx, &model.ConversationMember{
			ConvID:   conv.ID,
			UserID:   userID,
			JoinedAt: time.Now(),
		}); err != nil {
			return nil, err
		}
	}

	return conv, nil
}

func (s *ConversationService) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	return s.convRepo.GetByID(ctx, id)
}

func (s *ConversationService) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, int64, error) {
	convs, err := s.convRepo.ListByUser(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.convRepo.CountByUser(ctx, userID)
	return convs, total, err
}

func (s *ConversationService) ListUserConversationIDs(ctx context.Context, userID int64) ([]int64, error) {
	convs, _, err := s.ListByUser(ctx, userID, 0, 1000)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(convs))
	for _, conv := range convs {
		ids = append(ids, conv.ID)
	}
	return ids, nil
}

func (s *ConversationService) Update(ctx context.Context, conv *model.Conversation) error {
	conv.UpdatedAt = time.Now()
	return s.convRepo.Update(ctx, conv)
}

func (s *ConversationService) UpdateAnnouncement(ctx context.Context, convID, userID int64, announcement string) (*model.Conversation, error) {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv.Type == model.ConversationTypePrivate {
		return nil, apperrors.ErrPrivateChatImmutable
	}
	if conv.OwnerID != userID {
		return nil, apperrors.ErrNotGroupOwner
	}

	now := time.Now()
	conv.Announcement = announcement
	conv.AnnouncementUpdatedAt = &now
	conv.UpdatedAt = now
	if err := s.convRepo.Update(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *ConversationService) Delete(ctx context.Context, convID, userID int64) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return err
	}
	if conv.OwnerID != userID {
		return apperrors.ErrNotGroupOwner
	}
	return s.convRepo.Delete(ctx, convID)
}

func (s *ConversationService) AddMember(ctx context.Context, convID, userID int64) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return err
	}
	if conv.Type == model.ConversationTypePrivate {
		return apperrors.ErrPrivateChatImmutable
	}
	ok, err := s.convRepo.IsMember(ctx, convID, userID)
	if err != nil {
		return err
	}
	if ok {
		return apperrors.ErrMemberAlreadyInConv
	}
	return s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: convID, UserID: userID, JoinedAt: time.Now()})
}

func (s *ConversationService) RemoveMember(ctx context.Context, convID, userID int64) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return err
	}
	if conv.Type == model.ConversationTypePrivate {
		return apperrors.ErrPrivateChatImmutable
	}
	return s.convRepo.RemoveMember(ctx, convID, userID)
}

func (s *ConversationService) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	return s.convRepo.GetMembers(ctx, convID)
}

func (s *ConversationService) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	return s.convRepo.IsMember(ctx, convID, userID)
}
