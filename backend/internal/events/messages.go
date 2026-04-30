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

func MarshalMessageUpserted(msg *model.Message) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	return json.Marshal(msg)
}

func MarshalMessageDeleted(id int64) ([]byte, error) {
	return json.Marshal(MessageDeletedEvent{ID: id})
}

func PublishMessageUpserted(ctx context.Context, msg *model.Message) error {
	data, err := MarshalMessageUpserted(msg)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return msgqueue.Publish(ctx, TopicMessageUpserted, data)
}

func PublishMessageDeleted(ctx context.Context, id int64) error {
	data, err := MarshalMessageDeleted(id)
	if err != nil {
		return err
	}
	return msgqueue.Publish(ctx, TopicMessageDeleted, data)
}
