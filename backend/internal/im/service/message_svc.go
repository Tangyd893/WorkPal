package service

import (
	"context"
	"encoding/json"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/events"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByConvID(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error)
	Update(ctx context.Context, msg *model.Message) error
	SoftDelete(ctx context.Context, msgID int64) error
	CountUnread(ctx context.Context, convID, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID, convID int64) error
}

type MessageCommandStore interface {
	WithinTx(ctx context.Context, fn func(MessageTxStore) error) error
}

type MessageTxStore interface {
	CreateMessage(msg *model.Message) error
	GetMessage(id int64) (*model.Message, error)
	UpdateMessage(msg *model.Message) error
	SoftDeleteMessage(msgID int64) error
	EnqueueOutbox(event *model.MessageOutbox) error
}

type MessageService struct {
	msgRepo      MessageRepository
	commandStore MessageCommandStore
}

func NewMessageService(msgRepo MessageRepository) *MessageService {
	return &MessageService{msgRepo: msgRepo}
}

func NewReliableMessageService(msgRepo MessageRepository, commandStore MessageCommandStore) *MessageService {
	return &MessageService{
		msgRepo:      msgRepo,
		commandStore: commandStore,
	}
}

func newMessageService(msgRepo MessageRepository) *MessageService {
	return NewMessageService(msgRepo)
}

func (s *MessageService) Send(ctx context.Context, convID, senderID int64, msgType int8, content string, metadata map[string]interface{}, replyTo int64) (*model.Message, error) {
	msg, err := buildMessage(convID, senderID, msgType, content, metadata, replyTo)
	if err != nil {
		return nil, err
	}

	if s.commandStore == nil {
		if err := s.msgRepo.Create(ctx, msg); err != nil {
			return nil, err
		}
		return msg, nil
	}

	if err := s.commandStore.WithinTx(ctx, func(tx MessageTxStore) error {
		if err := tx.CreateMessage(msg); err != nil {
			return err
		}
		return enqueueUpsertEvent(tx, msg)
	}); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *MessageService) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	return s.msgRepo.GetByID(ctx, id)
}

func (s *MessageService) GetHistory(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.msgRepo.GetByConvID(ctx, convID, beforeID, limit)
}

func (s *MessageService) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	return s.msgRepo.CountUnread(ctx, convID, userID)
}

func (s *MessageService) MarkRead(ctx context.Context, userID, convID int64) error {
	return s.msgRepo.MarkRead(ctx, userID, convID)
}

func (s *MessageService) Edit(ctx context.Context, msgID, senderID int64, content string) (*model.Message, error) {
	if s.commandStore == nil {
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

	var updated *model.Message
	if err := s.commandStore.WithinTx(ctx, func(tx MessageTxStore) error {
		msg, err := tx.GetMessage(msgID)
		if err != nil {
			return err
		}
		if msg.SenderID != senderID {
			return apperrors.ErrCannotEditOthersMsg
		}
		msg.Content = content
		msg.UpdatedAt = time.Now()
		if err := tx.UpdateMessage(msg); err != nil {
			return err
		}
		if err := enqueueUpsertEvent(tx, msg); err != nil {
			return err
		}
		updated = msg
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *MessageService) Recall(ctx context.Context, msgID, senderID int64) error {
	if s.commandStore == nil {
		msg, err := s.msgRepo.GetByID(ctx, msgID)
		if err != nil {
			return err
		}
		if msg.SenderID != senderID {
			return apperrors.ErrCannotRecallOthers
		}
		return s.msgRepo.SoftDelete(ctx, msgID)
	}

	return s.commandStore.WithinTx(ctx, func(tx MessageTxStore) error {
		msg, err := tx.GetMessage(msgID)
		if err != nil {
			return err
		}
		if msg.SenderID != senderID {
			return apperrors.ErrCannotRecallOthers
		}
		if err := tx.SoftDeleteMessage(msgID); err != nil {
			return err
		}
		return enqueueDeleteEvent(tx, msgID)
	})
}

func buildMessage(convID, senderID int64, msgType int8, content string, metadata map[string]interface{}, replyTo int64) (*model.Message, error) {
	if content == "" && msgType != model.MessageTypeText {
		return nil, apperrors.ErrContentEmpty
	}

	metaJSON := ""
	if metadata != nil {
		data, _ := json.Marshal(metadata)
		metaJSON = string(data)
	}

	return &model.Message{
		ConvID:    convID,
		SenderID:  senderID,
		Type:      msgType,
		Content:   content,
		Metadata:  metaJSON,
		ReplyTo:   replyTo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func enqueueUpsertEvent(tx MessageTxStore, msg *model.Message) error {
	data, err := events.MarshalMessageUpserted(msg)
	if err != nil {
		return err
	}
	return tx.EnqueueOutbox(&model.MessageOutbox{
		Topic:         events.TopicMessageUpserted,
		Payload:       string(data),
		Status:        model.OutboxStatusPending,
		NextAttemptAt: time.Now(),
	})
}

func enqueueDeleteEvent(tx MessageTxStore, msgID int64) error {
	data, err := events.MarshalMessageDeleted(msgID)
	if err != nil {
		return err
	}
	return tx.EnqueueOutbox(&model.MessageOutbox{
		Topic:         events.TopicMessageDeleted,
		Payload:       string(data),
		Status:        model.OutboxStatusPending,
		NextAttemptAt: time.Now(),
	})
}
