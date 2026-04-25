package service

import (
	"context"
	"testing"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
)

// mockMessageRepo implements MessageRepository interface in-memory for testing
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

// errMsgNotFound matches the repo-level error that services compare with errors.Is
var errMsgNotFound = &testAppError{Code: 40400, Message: "资源不存在"}

type testAppError struct {
	Code    int
	Message string
}

func (e *testAppError) Error() string { return e.Message }

func TestMessageService_Send(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())

	t.Run("发送文字消息成功", func(t *testing.T) {
		msg, err := svc.Send(context.Background(), 1, 100, 1, "hello", nil, 0)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "hello", msg.Content)
		assert.Equal(t, int64(1), msg.ConvID)
		assert.Equal(t, int64(100), msg.SenderID)
	})

	t.Run("空内容且非文字类型被拒绝", func(t *testing.T) {
		msg, err := svc.Send(context.Background(), 1, 100, 2, "", nil, 0)
		assert.Error(t, err)
		assert.Nil(t, msg)
	})

	t.Run("带 metadata 发送", func(t *testing.T) {
		meta := map[string]interface{}{"key": "value"}
		msg, err := svc.Send(context.Background(), 1, 100, 1, "hello", meta, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.Metadata)
	})
}

func TestMessageService_Edit(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())
	ctx := context.Background()

	// 先发一条消息
	msg, _ := svc.Send(ctx, 1, 100, 1, "original", nil, 0)

	t.Run("本人编辑成功", func(t *testing.T) {
		updated, err := svc.Edit(ctx, msg.ID, 100, "updated content")
		assert.NoError(t, err)
		assert.Equal(t, "updated content", updated.Content)
	})

	t.Run("非本人编辑被拒绝", func(t *testing.T) {
		_, err := svc.Edit(ctx, msg.ID, 999, "hacked")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "只能编辑自己发送的消息")
	})

	t.Run("编辑不存在的消息", func(t *testing.T) {
		_, err := svc.Edit(ctx, 9999, 100, "nope")
		assert.Error(t, err)
	})
}

func TestMessageService_Recall(t *testing.T) {
	svc := newMessageService(newMockMessageRepo())
	ctx := context.Background()

	msg, _ := svc.Send(ctx, 1, 100, 1, "to be deleted", nil, 0)

	t.Run("本人撤回成功", func(t *testing.T) {
		err := svc.Recall(ctx, msg.ID, 100)
		assert.NoError(t, err)
		// 撤回后消息应不存在
		_, err = svc.GetByID(ctx, msg.ID)
		assert.Error(t, err)
	})

	t.Run("非本人撤回被拒绝", func(t *testing.T) {
		msg2, _ := svc.Send(ctx, 1, 100, 1, "another", nil, 0)
		err := svc.Recall(ctx, msg2.ID, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "只能撤回自己发送的消息")
	})
}

func TestMessageService_GetHistory(t *testing.T) {
	repo := newMockMessageRepo()
	svc := newMessageService(repo)
	ctx := context.Background()

	// 灌入 5 条消息
	for i := 1; i <= 5; i++ {
		repo.msgs[int64(i)] = &model.Message{ID: int64(i), ConvID: 1, SenderID: 100, Content: "msg"}
	}

	t.Run("默认 limit=20", func(t *testing.T) {
		msgs, err := svc.GetHistory(ctx, 1, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(msgs))
	})

	t.Run("limit 上限 100", func(t *testing.T) {
		msgs, err := svc.GetHistory(ctx, 1, 0, 200)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(msgs), 100)
	})
}
