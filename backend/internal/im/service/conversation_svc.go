package service

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/repo"
)

type ConversationService struct {
	convRepo *repo.ConversationRepo
}

func NewConversationService(convRepo *repo.ConversationRepo) *ConversationService {
	return &ConversationService{convRepo: convRepo}
}

// CreatePrivateConv 创建或获取私聊会话
func (s *ConversationService) CreatePrivateConv(ctx context.Context, userID, targetID int64) (*model.Conversation, error) {
	if userID == targetID {
		return nil, errors.New(40001, "不能和自己聊天")
	}

	// 先查找是否已有私聊
	conv, err := s.convRepo.FindPrivateConv(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}
	if conv != nil {
		return conv, nil
	}

	// 创建新私聊
	conv = &model.Conversation{
		Type:     model.ConversationTypePrivate,
		OwnerID:  userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}

	// 双方都加入会话
	if err := s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: conv.ID, UserID: userID, JoinedAt: time.Now()}); err != nil {
		return nil, err
	}
	if err := s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: conv.ID, UserID: targetID, JoinedAt: time.Now()}); err != nil {
		return nil, err
	}

	return conv, nil
}

// CreateGroup 创建群聊
func (s *ConversationService) CreateGroup(ctx context.Context, name string, ownerID int64, memberIDs []int64) (*model.Conversation, error) {
	if name == "" {
		name = "群聊"
	}
	conv := &model.Conversation{
		Type:     model.ConversationTypeGroup,
		Name:     name,
		OwnerID:  ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}

	// 所有者+所有成员加入
	allMembers := append([]int64{ownerID}, memberIDs...)
	for _, uid := range allMembers {
		if err := s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: conv.ID, UserID: uid, JoinedAt: time.Now()}); err != nil {
			return nil, err
		}
	}

	return conv, nil
}

// GetByID 获取会话
func (s *ConversationService) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	return s.convRepo.GetByID(ctx, id)
}

// ListByUser 获取用户会话列表
func (s *ConversationService) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, int64, error) {
	convs, err := s.convRepo.ListByUser(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.convRepo.CountByUser(ctx, userID)
	return convs, total, err
}

// Update 更新会话
func (s *ConversationService) Update(ctx context.Context, conv *model.Conversation) error {
	conv.UpdatedAt = time.Now()
	return s.convRepo.Update(ctx, conv)
}

// Delete 解散会话
func (s *ConversationService) Delete(ctx context.Context, convID, userID int64) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return err
	}
	if conv.OwnerID != userID {
		return errors.New(40301, "只有群主可以解散群聊")
	}
	return s.convRepo.Delete(ctx, convID)
}

// AddMember 添加成员
func (s *ConversationService) AddMember(ctx context.Context, convID, userID int64) error {
	ok, err := s.convRepo.IsMember(ctx, convID, userID)
	if err != nil {
		return err
	}
	if ok {
		return errors.New(40901, "用户已在群中")
	}
	return s.convRepo.AddMember(ctx, &model.ConversationMember{ConvID: convID, UserID: userID, JoinedAt: time.Now()})
}

// RemoveMember 移除成员
func (s *ConversationService) RemoveMember(ctx context.Context, convID, userID int64) error {
	conv, err := s.convRepo.GetByID(ctx, convID)
	if err != nil {
		return err
	}
	if conv.Type == model.ConversationTypePrivate {
		return errors.New(40001, "私聊无法移除成员")
	}
	return s.convRepo.RemoveMember(ctx, convID, userID)
}

// GetMembers 获取会话成员ID列表
func (s *ConversationService) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	return s.convRepo.GetMembers(ctx, convID)
}

// IsMember 检查是否是成员
func (s *ConversationService) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	return s.convRepo.IsMember(ctx, convID, userID)
}
