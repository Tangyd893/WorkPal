package repo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
)

// mockConversationRepo 内存实现，用于测试
type mockConversationRepo struct {
	mu      sync.RWMutex
	convs   map[int64]*model.Conversation
	members map[int64]map[int64]bool // convID -> userID -> true
	next    int64
}

func newMockConversationRepo() *mockConversationRepo {
	return &mockConversationRepo{
		convs:   make(map[int64]*model.Conversation),
		members: make(map[int64]map[int64]bool),
		next:    1,
	}
}

func (m *mockConversationRepo) Create(ctx context.Context, conv *model.Conversation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	conv.ID = m.next
	m.next++
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}
	conv.UpdatedAt = conv.CreatedAt
	m.convs[conv.ID] = conv
	return nil
}

func (m *mockConversationRepo) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.convs[id]; ok {
		return c, nil
	}
	return nil, ErrConvNotFound
}

func (m *mockConversationRepo) FindPrivateConv(ctx context.Context, uid1, uid2 int64) (*model.Conversation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.convs {
		if c.Type == model.ConversationTypePrivate {
			if m.members[c.ID] != nil && m.members[c.ID][uid1] && m.members[c.ID][uid2] {
				return c, nil
			}
		}
	}
	return nil, nil
}

func (m *mockConversationRepo) AddMember(ctx context.Context, member *model.ConversationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.members[member.ConvID] == nil {
		m.members[member.ConvID] = make(map[int64]bool)
	}
	m.members[member.ConvID][member.UserID] = true
	return nil
}

func (m *mockConversationRepo) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.members[convID] == nil {
		return false, nil
	}
	return m.members[convID][userID], nil
}

func (m *mockConversationRepo) RemoveMember(ctx context.Context, convID, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.members[convID] != nil {
		delete(m.members[convID], userID)
	}
	return nil
}

func (m *mockConversationRepo) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []int64
	if m.members[convID] != nil {
		for uid := range m.members[convID] {
			result = append(result, uid)
		}
	}
	return result, nil
}

func (m *mockConversationRepo) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*model.Conversation
	for _, c := range m.convs {
		if m.members[c.ID] != nil && m.members[c.ID][userID] {
			result = append(result, c)
		}
	}
	// 分页
	if offset >= len(result) {
		return []*model.Conversation{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockConversationRepo) CountByUser(ctx context.Context, userID int64) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, c := range m.convs {
		if m.members[c.ID] != nil && m.members[c.ID][userID] {
			count++
		}
	}
	return count, nil
}

func (m *mockConversationRepo) Update(ctx context.Context, conv *model.Conversation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.convs[conv.ID]; !ok {
		return ErrConvNotFound
	}
	conv.UpdatedAt = time.Now()
	m.convs[conv.ID] = conv
	return nil
}

func (m *mockConversationRepo) Delete(ctx context.Context, convID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.convs[convID]; !ok {
		return ErrConvNotFound
	}
	delete(m.convs, convID)
	delete(m.members, convID)
	return nil
}


// === 单元测试 ===

func TestConversationRepo_Create(t *testing.T) {
	repo := newMockConversationRepo()

	conv := &model.Conversation{Type: model.ConversationTypePrivate, OwnerID: 1}
	err := repo.Create(context.Background(), conv)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), conv.ID)
}

func TestConversationRepo_GetByID(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1, Type: model.ConversationTypePrivate, OwnerID: 1}

	t.Run("成功查询", func(t *testing.T) {
		c, err := repo.GetByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, int8(1), c.Type)
	})

	t.Run("会话不存在", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 999)
		assert.ErrorIs(t, err, ErrConvNotFound)
	})
}

func TestConversationRepo_FindPrivateConv(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1, Type: model.ConversationTypePrivate, OwnerID: 100}
	repo.members[1] = map[int64]bool{100: true, 200: true}

	t.Run("找到私聊", func(t *testing.T) {
		c, err := repo.FindPrivateConv(context.Background(), 100, 200)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, int64(1), c.ID)
	})

	t.Run("不存在私聊", func(t *testing.T) {
		c, err := repo.FindPrivateConv(context.Background(), 100, 999)
		assert.NoError(t, err)
		assert.Nil(t, c)
	})
}

func TestConversationRepo_AddMember(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1, Type: model.ConversationTypeGroup, OwnerID: 1}

	err := repo.AddMember(context.Background(), &model.ConversationMember{ConvID: 1, UserID: 100})
	assert.NoError(t, err)

	isMember, _ := repo.IsMember(context.Background(), 1, 100)
	assert.True(t, isMember)
}

func TestConversationRepo_IsMember(t *testing.T) {
	repo := newMockConversationRepo()
	repo.members[1] = map[int64]bool{100: true}

	t.Run("是成员", func(t *testing.T) {
		ok, err := repo.IsMember(context.Background(), 1, 100)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("非成员", func(t *testing.T) {
		ok, err := repo.IsMember(context.Background(), 1, 999)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestConversationRepo_RemoveMember(t *testing.T) {
	repo := newMockConversationRepo()
	repo.members[1] = map[int64]bool{100: true}

	err := repo.RemoveMember(context.Background(), 1, 100)
	assert.NoError(t, err)

	ok, _ := repo.IsMember(context.Background(), 1, 100)
	assert.False(t, ok)
}

func TestConversationRepo_GetMembers(t *testing.T) {
	repo := newMockConversationRepo()
	repo.members[1] = map[int64]bool{100: true, 200: true, 300: true}

	members, err := repo.GetMembers(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, members, 3)
}

func TestConversationRepo_ListByUser(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1, Type: model.ConversationTypePrivate, OwnerID: 1}
	repo.convs[2] = &model.Conversation{ID: 2, Type: model.ConversationTypeGroup, OwnerID: 1}
	repo.convs[3] = &model.Conversation{ID: 3, Type: model.ConversationTypeGroup, OwnerID: 2}
	repo.members[1] = map[int64]bool{100: true}
	repo.members[2] = map[int64]bool{100: true, 200: true}

	t.Run("用户100的会话列表", func(t *testing.T) {
		convs, err := repo.ListByUser(context.Background(), 100, 0, 20)
		assert.NoError(t, err)
		assert.Len(t, convs, 2) // conv 1 和 2
	})

	t.Run("分页", func(t *testing.T) {
		convs, err := repo.ListByUser(context.Background(), 100, 0, 1)
		assert.NoError(t, err)
		assert.Len(t, convs, 1)
	})
}

func TestConversationRepo_CountByUser(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1}
	repo.convs[2] = &model.Conversation{ID: 2}
	repo.members[1] = map[int64]bool{100: true}
	repo.members[2] = map[int64]bool{100: true, 200: true}

	count, err := repo.CountByUser(context.Background(), 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestConversationRepo_Update(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1, Name: "Old Name", Type: model.ConversationTypeGroup, OwnerID: 1}

	t.Run("成功更新", func(t *testing.T) {
		conv := &model.Conversation{ID: 1, Name: "New Name", Type: model.ConversationTypeGroup, OwnerID: 1}
		err := repo.Update(context.Background(), conv)
		assert.NoError(t, err)
		assert.Equal(t, "New Name", repo.convs[1].Name)
	})

	t.Run("会话不存在", func(t *testing.T) {
		conv := &model.Conversation{ID: 999, Name: "Ghost"}
		err := repo.Update(context.Background(), conv)
		assert.ErrorIs(t, err, ErrConvNotFound)
	})
}

func TestConversationRepo_Delete(t *testing.T) {
	repo := newMockConversationRepo()
	repo.convs[1] = &model.Conversation{ID: 1}
	repo.members[1] = map[int64]bool{100: true}

	t.Run("成功删除", func(t *testing.T) {
		err := repo.Delete(context.Background(), 1)
		assert.NoError(t, err)
		_, err = repo.GetByID(context.Background(), 1)
		assert.ErrorIs(t, err, ErrConvNotFound)
	})

	t.Run("会话不存在", func(t *testing.T) {
		err := repo.Delete(context.Background(), 999)
		assert.ErrorIs(t, err, ErrConvNotFound)
	})
}