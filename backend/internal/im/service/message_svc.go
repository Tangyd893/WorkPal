package service

import (
	"context"
	"encoding/json"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
)

// MessageRepository 接口，便于测试注入 mock
type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByConvID(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error)
	Update(ctx context.Context, msg *model.Message) error
	SoftDelete(ctx context.Context, msgID int64) error
	CountUnread(ctx context.Context, convID, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID, convID int64) error
}

type MessageService struct {
	msgRepo MessageRepository
}

func NewMessageService(msgRepo MessageRepository) *MessageService {
	return &MessageService{msgRepo: msgRepo}
}

func newMessageService(msgRepo MessageRepository) *MessageService {
	return NewMessageService(msgRepo)
}

// Send 发送消息（创建消息记录）
func (s *MessageService) Send(ctx context.Context, convID, senderID int64, msgType int8, content string, metadata map[string]interface{}, replyTo int64) (*model.Message, error) {
	if content == "" && msgType != model.MessageTypeText {
		return nil, apperrors.ErrContentEmpty
	}

	metaJSON := ""
	if metadata != nil {
		data, _ := json.Marshal(metadata)
		metaJSON = string(data)
	}

	msg := &model.Message{
		ConvID:    convID,
		SenderID:  senderID,
		Type:      msgType,
		Content:   content,
		Metadata:  metaJSON,
		ReplyTo:   replyTo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// GetByID 获取消息
func (s *MessageService) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	return s.msgRepo.GetByID(ctx, id)
}

// GetHistory 获取历史消息（分页）
func (s *MessageService) GetHistory(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.msgRepo.GetByConvID(ctx, convID, beforeID, limit)
}

// CountUnread 统计未读数
func (s *MessageService) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	return s.msgRepo.CountUnread(ctx, convID, userID)
}

// MarkRead 标记已读
func (s *MessageService) MarkRead(ctx context.Context, userID, convID int64) error {
	return s.msgRepo.MarkRead(ctx, userID, convID)
}

// Edit 编辑消息（只能修改自己发送的）
func (s *MessageService) Edit(ctx context.Context, msgID, senderID int64, content string) (*model.Message, error) {
	msg, err := s.msgRepo.GetByID(ctx, msgID)
	if err != nil {
		return nil, err
	}
	if msg.SenderID != senderID {
		return nil, apperrors.ErrCannotEditOthersMsg
	}
	msg.Content = content
	msg.UpdatedAt = time.Now()
	if err := s.msgRepo.Update(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Recall 撤回消息（软删除）
func (s *MessageService) Recall(ctx context.Context, msgID, senderID int64) error {
	msg, err := s.msgRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}
	if msg.SenderID != senderID {
		return apperrors.ErrCannotRecallOthers
	}
	return s.msgRepo.SoftDelete(ctx, msgID)
}
