package events

import (
	"context"
	"encoding/json"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/pkg/msgqueue"
)

const (
	TopicMessageUpserted = "message.upserted"
	TopicMessageDeleted  = "message.deleted"
)

type MessageDeletedEvent struct {
	ID int64 `json:"id"`
}

func PublishMessageUpserted(ctx context.Context, msg *model.Message) error {
	if msg == nil {
		return nil
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return msgqueue.Publish(ctx, TopicMessageUpserted, data)
}

func PublishMessageDeleted(ctx context.Context, id int64) error {
	data, err := json.Marshal(MessageDeletedEvent{ID: id})
	if err != nil {
		return err
	}
	return msgqueue.Publish(ctx, TopicMessageDeleted, data)
}
