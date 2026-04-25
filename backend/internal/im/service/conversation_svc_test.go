package service

import (
	"context"
	"testing"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
)

// mockConversationRepo implements ConversationRepository in-memory for testing
type mockConversationRepo struct {
	convs   map[int64]*model.Conversation
	members map[int64]map[int64]bool // convID -> userID -> true
	nextID  int64
}

func newMockConversationRepo() *mockConversationRepo {
	return &mockConversationRepo{
		convs:   make(map[int64]*model.Conversation),
		members: make(map[int64]map[int64]bool),
		nextID:  1,
	}
}

func (r *mockConversationRepo) Create(ctx context.Context, conv *model.Conversation) error {
	conv.ID = r.nextID
	r.nextID++
	r.convs[conv.ID] = conv
	return nil
}

func (r *mockConversationRepo) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	if c, ok := r.convs[id]; ok {
		return c, nil
	}
	return nil, errConvNotFound
}

func (r *mockConversationRepo) FindPrivateConv(ctx context.Context, uid1, uid2 int64) (*model.Conversation, error) {
	for _, c := range r.convs {
		if c.Type == model.ConversationTypePrivate {
			m1 := r.members[c.ID][uid1]
			m2 := r.members[c.ID][uid2]
			if m1 && m2 {
				return c, nil
			}
		}
	}
	return nil, nil
}

func (r *mockConversationRepo) AddMember(ctx context.Context, m *model.ConversationMember) error {
	if r.members[m.ConvID] == nil {
		r.members[m.ConvID] = make(map[int64]bool)
	}
	r.members[m.ConvID][m.UserID] = true
	return nil
}

func (r *mockConversationRepo) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	return r.members[convID][userID], nil
}

func (r *mockConversationRepo) RemoveMember(ctx context.Context, convID, userID int64) error {
	delete(r.members[convID], userID)
	return nil
}

func (r *mockConversationRepo) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	var ids []int64
	for uid := range r.members[convID] {
		ids = append(ids, uid)
	}
	return ids, nil
}

func (r *mockConversationRepo) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, error) {
	var result []*model.Conversation
	for _, c := range r.convs {
		if r.members[c.ID][userID] {
			result = append(result, c)
		}
	}
	if offset > len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (r *mockConversationRepo) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	for _, c := range r.convs {
		if r.members[c.ID][userID] {
			count++
		}
	}
	return count, nil
}

func (r *mockConversationRepo) Update(ctx context.Context, conv *model.Conversation) error {
	r.convs[conv.ID] = conv
	return nil
}

func (r *mockConversationRepo) Delete(ctx context.Context, convID int64) error {
	delete(r.convs, convID)
	delete(r.members, convID)
	return nil
}

var errConvNotFound = &testAppError{Code: 40400, Message: "资源不存在"}

func TestConversationService_CreatePrivateConv(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	t.Run("创建新私聊", func(t *testing.T) {
		conv, err := svc.CreatePrivateConv(ctx, 1, 2)
		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, int8(1), conv.Type)
	})

	t.Run("重复创建返回已有会话", func(t *testing.T) {
		conv2, err := svc.CreatePrivateConv(ctx, 1, 2)
		assert.NoError(t, err)
		assert.NotNil(t, conv2)
	})

	t.Run("不能和自己聊天", func(t *testing.T) {
		_, err := svc.CreatePrivateConv(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不能和自己聊天")
	})
}

func TestConversationService_CreateGroup(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	t.Run("创建群聊", func(t *testing.T) {
		conv, err := svc.CreateGroup(ctx, "测试群", 1, []int64{2, 3})
		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, "测试群", conv.Name)
		assert.Equal(t, int8(2), conv.Type)
		assert.Equal(t, int64(1), conv.OwnerID)
	})

	t.Run("空名称默认为群聊", func(t *testing.T) {
		conv, err := svc.CreateGroup(ctx, "", 1, []int64{2})
		assert.NoError(t, err)
		assert.Equal(t, "群聊", conv.Name)
	})
}

func TestConversationService_Delete(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	conv, _ := svc.CreateGroup(ctx, "待解散", 1, []int64{2})

	t.Run("群主解散成功", func(t *testing.T) {
		err := svc.Delete(ctx, conv.ID, 1)
		assert.NoError(t, err)
	})

	t.Run("非群主解散被拒绝", func(t *testing.T) {
		conv2, _ := svc.CreateGroup(ctx, "另一个群", 1, []int64{2})
		err := svc.Delete(ctx, conv2.ID, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "只有群主可以解散群聊")
	})
}

func TestConversationService_AddMember(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	conv, _ := svc.CreateGroup(ctx, "群1", 1, []int64{})

	t.Run("添加成员成功", func(t *testing.T) {
		err := svc.AddMember(ctx, conv.ID, 2)
		assert.NoError(t, err)
	})

	t.Run("重复添加被拒绝", func(t *testing.T) {
		err := svc.AddMember(ctx, conv.ID, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已在群中")
	})
}

func TestConversationService_RemoveMember(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	conv, _ := svc.CreateGroup(ctx, "待移出", 1, []int64{2, 3})

	t.Run("移除成员成功", func(t *testing.T) {
		err := svc.RemoveMember(ctx, conv.ID, 3)
		assert.NoError(t, err)
	})

	t.Run("私聊无法移除成员", func(t *testing.T) {
		priv, _ := svc.CreatePrivateConv(ctx, 1, 2)
		err := svc.RemoveMember(ctx, priv.ID, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "私聊无法移除成员")
	})
}

func TestConversationService_ListByUser(t *testing.T) {
	svc := newConversationService(newMockConversationRepo())
	ctx := context.Background()

	// 创建两个会话都属于 userID=1
	_, _ = svc.CreatePrivateConv(ctx, 1, 2)
	_, _ = svc.CreatePrivateConv(ctx, 1, 3)

	convs, total, err := svc.ListByUser(ctx, 1, 0, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(convs))
	assert.Equal(t, int64(2), total)
}
