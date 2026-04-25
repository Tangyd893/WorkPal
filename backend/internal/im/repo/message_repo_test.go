package repo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// mockMessageRepo 内存实现，用于测试
type mockMessageRepo struct {
	mu   sync.RWMutex
	msgs map[int64]*model.Message
	next int64
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{msgs: make(map[int64]*model.Message), next: 1}
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *model.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg.ID = m.next
	m.next++
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	m.msgs[msg.ID] = msg
	return nil
}

func (m *mockMessageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if msg, ok := m.msgs[id]; ok {
		return msg, nil
	}
	return nil, ErrMessageNotFound
}

func (m *mockMessageRepo) GetByConvID(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.Message
	for _, msg := range m.msgs {
		if msg.ConvID == convID && msg.DeletedAt.Valid == false {
			if beforeID == 0 || msg.ID < beforeID {
				result = append(result, msg)
			}
		}
	}
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ID < result[j].ID {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockMessageRepo) Update(ctx context.Context, msg *model.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.msgs[msg.ID]; !ok {
		return ErrMessageNotFound
	}
	m.msgs[msg.ID] = msg
	return nil
}

func (m *mockMessageRepo) SoftDelete(ctx context.Context, msgID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if msg, ok := m.msgs[msgID]; ok {
		msg.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
		return nil
	}
	return ErrMessageNotFound
}

func (m *mockMessageRepo) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	return 0, nil
}

func (m *mockMessageRepo) MarkRead(ctx context.Context, userID, convID int64) error {
	return nil
}

func TestMessageRepo_Create(t *testing.T) {
	repo := newMockMessageRepo()
	msg := &model.Message{ConvID: 1, SenderID: 100, Type: 1, Content: "hello"}
	err := repo.Create(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), msg.ID)
}

func TestMessageRepo_GetByID(t *testing.T) {
	repo := newMockMessageRepo()
	now := time.Now()
	repo.msgs[1] = &model.Message{ID: 1, ConvID: 1, SenderID: 100, Type: 1, Content: "test", CreatedAt: now}

	t.Run("成功查询", func(t *testing.T) {
		msg, err := repo.GetByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, "test", msg.Content)
	})

	t.Run("消息不存在", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 999)
		assert.ErrorIs(t, err, ErrMessageNotFound)
	})
}

func TestMessageRepo_GetByConvID(t *testing.T) {
	repo := newMockMessageRepo()
	now := time.Now()
	for i := int64(1); i <= 5; i++ {
		repo.msgs[i] = &model.Message{ID: i, ConvID: 1, SenderID: 100, Type: 1, Content: "msg", CreatedAt: now}
	}

	t.Run("成功获取", func(t *testing.T) {
		msgs, err := repo.GetByConvID(context.Background(), 1, 0, 20)
		assert.NoError(t, err)
		assert.Len(t, msgs, 5)
	})

	t.Run("limit 限制", func(t *testing.T) {
		msgs, err := repo.GetByConvID(context.Background(), 1, 0, 3)
		assert.NoError(t, err)
		assert.Len(t, msgs, 3)
	})

	t.Run("beforeID 分页", func(t *testing.T) {
		msgs, err := repo.GetByConvID(context.Background(), 1, 4, 10)
		assert.NoError(t, err)
		assert.Len(t, msgs, 3)
	})
}

func TestMessageRepo_Update(t *testing.T) {
	repo := newMockMessageRepo()
	now := time.Now()
	repo.msgs[1] = &model.Message{ID: 1, ConvID: 1, SenderID: 100, Type: 1, Content: "original", CreatedAt: now}

	t.Run("成功更新", func(t *testing.T) {
		msg := &model.Message{ID: 1, Content: "updated"}
		err := repo.Update(context.Background(), msg)
		assert.NoError(t, err)
		assert.Equal(t, "updated", repo.msgs[1].Content)
	})

	t.Run("消息不存在", func(t *testing.T) {
		msg := &model.Message{ID: 999, Content: "ghost"}
		err := repo.Update(context.Background(), msg)
		assert.ErrorIs(t, err, ErrMessageNotFound)
	})
}

func TestMessageRepo_SoftDelete(t *testing.T) {
	repo := newMockMessageRepo()
	now := time.Now()
	repo.msgs[1] = &model.Message{ID: 1, ConvID: 1, SenderID: 100, Type: 1, Content: "to delete", CreatedAt: now}

	t.Run("成功软删除", func(t *testing.T) {
		err := repo.SoftDelete(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, repo.msgs[1].DeletedAt.Valid)
	})

	t.Run("消息不存在", func(t *testing.T) {
		err := repo.SoftDelete(context.Background(), 999)
		assert.ErrorIs(t, err, ErrMessageNotFound)
	})
}

func TestMessageRepo_CountUnread(t *testing.T) {
	repo := newMockMessageRepo()
	count, err := repo.CountUnread(context.Background(), 1, 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMessageRepo_MarkRead(t *testing.T) {
	repo := newMockMessageRepo()
	err := repo.MarkRead(context.Background(), 100, 1)
	assert.NoError(t, err)
}