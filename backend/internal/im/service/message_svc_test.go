package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/events"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
)

type mockMessageRepo struct {
	msgs map[int64]*model.Message
	next int64
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		msgs: make(map[int64]*model.Message),
		next: 1,
	}
}

func (r *mockMessageRepo) Create(ctx context.Context, msg *model.Message) error {
	msg.ID = r.next
	r.next++
	r.msgs[msg.ID] = msg
	return nil
}

func (r *mockMessageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	if msg, ok := r.msgs[id]; ok {
		return msg, nil
	}
	return nil, errMsgNotFound
}

func (r *mockMessageRepo) GetByConvID(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error) {
	var result []*model.Message
	for _, m := range r.msgs {
		if m.ConvID == convID && (beforeID == 0 || m.ID < beforeID) {
			result = append(result, m)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *mockMessageRepo) GetByConvIDAndTimeRange(ctx context.Context, convID int64, startAt, endAt time.Time, limit int) ([]*model.Message, error) {
	var result []*model.Message
	for _, m := range r.msgs {
		if m.ConvID != convID {
			continue
		}
		if !startAt.IsZero() && m.CreatedAt.Before(startAt) {
			continue
		}
		if !endAt.IsZero() && m.CreatedAt.After(endAt) {
			continue
		}
		result = append(result, m)
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *mockMessageRepo) GetByIdempotencyKey(ctx context.Context, convID, senderID int64, key string, validAfter time.Time) (*model.Message, error) {
	for _, m := range r.msgs {
		if m.ConvID == convID && m.SenderID == senderID && m.IdempotencyKey == key && !m.CreatedAt.Before(validAfter) {
			return m, nil
		}
	}
	return nil, apperrors.ErrMessageNotFound
}

func (r *mockMessageRepo) ClearExpiredIdempotencyKeys(ctx context.Context, before time.Time) error {
	for _, m := range r.msgs {
		if !m.CreatedAt.IsZero() && m.CreatedAt.Before(before) {
			m.IdempotencyKey = ""
		}
	}
	return nil
}

func (r *mockMessageRepo) Update(ctx context.Context, msg *model.Message) error {
	if _, ok := r.msgs[msg.ID]; !ok {
		return errMsgNotFound
	}
	r.msgs[msg.ID] = msg
	return nil
}

func (r *mockMessageRepo) SoftDelete(ctx context.Context, msgID int64) error {
	if _, ok := r.msgs[msgID]; !ok {
		return errMsgNotFound
	}
	delete(r.msgs, msgID)
	return nil
}

func (r *mockMessageRepo) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	return 0, nil
}

func (r *mockMessageRepo) MarkRead(ctx context.Context, userID, convID int64) error {
	return nil
}

var errMsgNotFound = &testAppError{Code: 40400, Message: "资源不存在"}

type testAppError struct {
	Code    int
	Message string
}

func (e *testAppError) Error() string { return e.Message }

type mockCommandStore struct {
	repo   *mockMessageRepo
	outbox []*model.MessageOutbox
}

func newMockCommandStore(repo *mockMessageRepo) *mockCommandStore {
	return &mockCommandStore{repo: repo}
}

func (s *mockCommandStore) WithinTx(ctx context.Context, fn func(MessageTxStore) error) error {
	return fn(&mockTxStore{repo: s.repo, outbox: &s.outbox})
}

type mockTxStore struct {
	repo   *mockMessageRepo
	outbox *[]*model.MessageOutbox
}

func (s *mockTxStore) CreateMessage(msg *model.Message) error {
	return s.repo.Create(context.Background(), msg)
}

func (s *mockTxStore) GetMessage(id int64) (*model.Message, error) {
	return s.repo.GetByID(context.Background(), id)
}

func (s *mockTxStore) UpdateMessage(msg *model.Message) error {
	return s.repo.Update(context.Background(), msg)
}

func (s *mockTxStore) SoftDeleteMessage(msgID int64) error {
	return s.repo.SoftDelete(context.Background(), msgID)
}

func (s *mockTxStore) EnqueueOutbox(event *model.MessageOutbox) error {
	copied := *event
	if copied.NextAttemptAt.IsZero() {
		copied.NextAttemptAt = time.Now()
	}
	*s.outbox = append(*s.outbox, &copied)
	return nil
}

func TestMessageServiceSend(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())

	t.Run("sends a text message", func(t *testing.T) {
		msg, err := svc.Send(context.Background(), 1, 100, 1, "hello", nil, 0)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "hello", msg.Content)
		assert.Equal(t, int64(1), msg.ConvID)
		assert.Equal(t, int64(100), msg.SenderID)
	})

	t.Run("rejects empty non-text content", func(t *testing.T) {
		msg, err := svc.Send(context.Background(), 1, 100, 2, "", nil, 0)
		assert.Error(t, err)
		assert.Nil(t, msg)
	})

	t.Run("serializes metadata", func(t *testing.T) {
		meta := map[string]interface{}{"key": "value"}
		msg, err := svc.Send(context.Background(), 1, 100, 1, "hello", meta, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.Metadata)
	})
}

func TestMessageServiceSendWithIdempotency(t *testing.T) {
	repo := newMockMessageRepo()
	svc := newMessageService(repo)
	ctx := context.Background()

	first, err := svc.SendWithIdempotency(ctx, 1, 100, model.MessageTypeText, "hello", nil, 0, "idem-001")
	assert.NoError(t, err)

	second, err := svc.SendWithIdempotency(ctx, 1, 100, model.MessageTypeText, "hello again", nil, 0, "idem-001")
	assert.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "hello", second.Content)
	assert.Len(t, repo.msgs, 1)
}

func TestMessageServiceEdit(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())
	ctx := context.Background()

	msg, _ := svc.Send(ctx, 1, 100, 1, "original", nil, 0)

	t.Run("allows the sender to edit", func(t *testing.T) {
		updated, err := svc.Edit(ctx, msg.ID, 100, "updated content")
		assert.NoError(t, err)
		assert.Equal(t, "updated content", updated.Content)
	})

	t.Run("rejects editing others messages", func(t *testing.T) {
		_, err := svc.Edit(ctx, msg.ID, 999, "hacked")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "只能编辑自己发送的消息")
	})

	t.Run("returns an error when the message is missing", func(t *testing.T) {
		_, err := svc.Edit(ctx, 9999, 100, "nope")
		assert.Error(t, err)
	})
}

func TestMessageServiceRecall(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())
	ctx := context.Background()

	msg, _ := svc.Send(ctx, 1, 100, 1, "to be deleted", nil, 0)

	t.Run("allows the sender to recall", func(t *testing.T) {
		err := svc.Recall(ctx, msg.ID, 100)
		assert.NoError(t, err)
		_, err = svc.GetByID(ctx, msg.ID)
		assert.Error(t, err)
	})

	t.Run("rejects recalling others messages", func(t *testing.T) {
		msg2, _ := svc.Send(ctx, 1, 100, 1, "another", nil, 0)
		err := svc.Recall(ctx, msg2.ID, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "只能撤回自己发送的消息")
	})
}

func TestMessageServiceGetHistory(t *testing.T) {
	repo := newMockMessageRepo()
	svc := newMessageService(repo)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		repo.msgs[int64(i)] = &model.Message{ID: int64(i), ConvID: 1, SenderID: 100, Content: "msg"}
	}

	t.Run("uses the default limit", func(t *testing.T) {
		msgs, err := svc.GetHistory(ctx, 1, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(msgs))
	})

	t.Run("caps the limit at 100", func(t *testing.T) {
		msgs, err := svc.GetHistory(ctx, 1, 0, 200)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(msgs), 100)
	})
}

func TestReliableMessageServiceEnqueuesOutbox(t *testing.T) {
	repo := newMockMessageRepo()
	commandStore := newMockCommandStore(repo)
	svc := NewReliableMessageService(repo, commandStore)
	ctx := context.Background()

	msg, err := svc.Send(ctx, 99, 7, model.MessageTypeText, "hello reliable world", nil, 0)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Len(t, commandStore.outbox, 1)
	assert.Equal(t, events.TopicMessageUpserted, commandStore.outbox[0].Topic)
	assert.Equal(t, model.OutboxStatusPending, commandStore.outbox[0].Status)

	var payload model.Message
	err = json.Unmarshal([]byte(commandStore.outbox[0].Payload), &payload)
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, payload.ID)
	assert.Equal(t, "hello reliable world", payload.Content)

	updated, err := svc.Edit(ctx, msg.ID, 7, "edited")
	assert.NoError(t, err)
	assert.Equal(t, "edited", updated.Content)
	assert.Len(t, commandStore.outbox, 2)
	assert.Equal(t, events.TopicMessageUpserted, commandStore.outbox[1].Topic)

	err = svc.Recall(ctx, msg.ID, 7)
	assert.NoError(t, err)
	assert.Len(t, commandStore.outbox, 3)
	assert.Equal(t, events.TopicMessageDeleted, commandStore.outbox[2].Topic)

	var deleted events.MessageDeletedEvent
	err = json.Unmarshal([]byte(commandStore.outbox[2].Payload), &deleted)
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, deleted.ID)
}
