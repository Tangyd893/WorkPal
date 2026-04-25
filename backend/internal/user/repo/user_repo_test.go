package repo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/stretchr/testify/assert"
)

// mockUserRepo 内存实现，用于测试
type mockUserRepo struct {
	mu   sync.RWMutex
	data map[int64]*model.User
	next int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{data: make(map[int64]*model.User), next: 1}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.data {
		if u.Username == user.Username {
			return ErrUserAlreadyExists
		}
	}
	user.ID = m.next
	m.next++
	m.data[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.data[id]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.data {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[user.ID]; !ok {
		return ErrUserNotFound
	}
	m.data[user.ID] = user
	return nil
}

func (m *mockUserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]*model.User, 0, len(m.data))
	for _, u := range m.data {
		all = append(all, u)
	}
	if offset >= len(all) {
		return []*model.User{}, int64(len(all)), nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], int64(len(all)), nil
}

func (m *mockUserRepo) Search(ctx context.Context, keyword string, offset, limit int) ([]*model.User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var matched []*model.User
	for _, u := range m.data {
		if containsIgnoreCase(u.Username, keyword) || containsIgnoreCase(u.Nickname, keyword) {
			matched = append(matched, u)
		}
	}
	total := int64(len(matched))
	if offset >= len(matched) {
		return []*model.User{}, total, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	return matched[offset:end], total, nil
}

func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	// simple case-insensitive contains
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func TestUserRepo_Create(t *testing.T) {
	repo := newMockUserRepo()

	t.Run("成功创建", func(t *testing.T) {
		user := &model.User{Username: "alice", PasswordHash: "hash", Nickname: "Alice", Email: "alice@example.com"}
		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
	})

	t.Run("用户名重复", func(t *testing.T) {
		user := &model.User{Username: "alice", PasswordHash: "hash"}
		err := repo.Create(context.Background(), user)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	repo := newMockUserRepo()
	repo.data[1] = &model.User{ID: 1, Username: "alice", PasswordHash: "hash", Nickname: "Alice", Status: 1, CreatedAt: time.Now()}

	t.Run("成功查询", func(t *testing.T) {
		user, err := repo.GetByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, "alice", user.Username)
	})

	t.Run("用户不存在", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), 999)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUserRepo_GetByUsername(t *testing.T) {
	repo := newMockUserRepo()
	repo.data[1] = &model.User{ID: 1, Username: "alice", PasswordHash: "hash", Nickname: "Alice", Status: 1, CreatedAt: time.Now()}

	t.Run("成功查询", func(t *testing.T) {
		user, err := repo.GetByUsername(context.Background(), "alice")
		assert.NoError(t, err)
		assert.Equal(t, "alice", user.Username)
	})

	t.Run("用户名不存在", func(t *testing.T) {
		_, err := repo.GetByUsername(context.Background(), "nobody")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUserRepo_Update(t *testing.T) {
	repo := newMockUserRepo()
	repo.data[1] = &model.User{ID: 1, Username: "alice", PasswordHash: "hash", Nickname: "Alice", Status: 1, CreatedAt: time.Now()}

	t.Run("成功更新", func(t *testing.T) {
		user := &model.User{ID: 1, Username: "alice", PasswordHash: "newhash", Nickname: "Alice Updated"}
		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, "newhash", repo.data[1].PasswordHash)
	})

	t.Run("用户不存在", func(t *testing.T) {
		user := &model.User{ID: 999, Username: "ghost", PasswordHash: "hash"}
		err := repo.Update(context.Background(), user)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUserRepo_List(t *testing.T) {
	repo := newMockUserRepo()
	for i := int64(1); i <= 5; i++ {
		repo.data[i] = &model.User{ID: i, Username: "user", PasswordHash: "hash", Nickname: "User", Status: 1, CreatedAt: time.Now()}
	}

	t.Run("成功列出（分页）", func(t *testing.T) {
		users, total, err := repo.List(context.Background(), 0, 3)
		assert.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, int64(5), total)
	})

	t.Run("offset 超出范围", func(t *testing.T) {
		users, total, err := repo.List(context.Background(), 100, 10)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
		assert.Equal(t, int64(5), total)
	})
}

func TestUserRepo_Search(t *testing.T) {
	repo := newMockUserRepo()
	repo.data[1] = &model.User{ID: 1, Username: "alice", Nickname: "Alice Smith", Status: 1, CreatedAt: time.Now()}
	repo.data[2] = &model.User{ID: 2, Username: "bob", Nickname: "Bob Jones", Status: 1, CreatedAt: time.Now()}
	repo.data[3] = &model.User{ID: 3, Username: "charlie", Nickname: "Charlie", Status: 1, CreatedAt: time.Now()}

	t.Run("按用户名搜索", func(t *testing.T) {
		users, total, err := repo.Search(context.Background(), "alice", 0, 10)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "alice", users[0].Username)
	})

	t.Run("按昵称搜索", func(t *testing.T) {
		users, total, err := repo.Search(context.Background(), "Smith", 0, 10)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Alice Smith", users[0].Nickname)
	})

	t.Run("空关键词返回全部", func(t *testing.T) {
		users, total, err := repo.Search(context.Background(), "", 0, 10)
		assert.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, int64(3), total)
	})

	t.Run("无匹配结果", func(t *testing.T) {
		users, total, err := repo.Search(context.Background(), "xyz", 0, 10)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
		assert.Equal(t, int64(0), total)
	})
}