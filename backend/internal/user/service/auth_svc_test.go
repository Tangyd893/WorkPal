package service

import (
	"context"
	"testing"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepo implements a simple in-memory user repo for testing
type mockUserRepo struct {
	users    map[string]*model.User // key: username
	byID     map[int64]*model.User // key: userID
	nextID   int64
	createErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]*model.User),
		byID:   make(map[int64]*model.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, exists := m.users[user.Username]; exists {
		return errUserAlreadyExists
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Username] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if user, ok := m.users[username]; ok {
		return user, nil
	}
	return nil, errUserNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if user, ok := m.byID[id]; ok {
		return user, nil
	}
	return nil, errUserNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	if _, ok := m.byID[user.ID]; !ok {
		return errUserNotFound
	}
	m.byID[user.ID] = user
	m.users[user.Username] = user
	return nil
}

var errUserAlreadyExists = apperrors.ErrUserAlreadyExists
var errUserNotFound  = apperrors.ErrUserNotFound

func TestRegister(t *testing.T) {
	auth.SetSecret("test-secret-for-unit-test")
	defer auth.SetSecret("")

	svc := NewAuthService(newMockUserRepo(), 24)

	tests := []struct {
		name    string
		req     *RegisterReq
		wantErr bool
		errType error
	}{
		{
			name: "正常注册",
			req: &RegisterReq{
				Username: "alice",
				Password: "password123",
				Nickname: "Alice",
				Email:    "alice@example.com",
			},
			wantErr: false,
		},
		{
			name: "用户名已存在",
			req: &RegisterReq{
				Username: "alice", // same as first test
				Password: "password123",
			},
			wantErr: true,
			errType: apperrors.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.Register(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Username, user.Username)
				assert.NotEmpty(t, user.PasswordHash)
				// Password should be hashed (not plain text)
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.req.Password))
				assert.NoError(t, err, "password should match hash")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	auth.SetSecret("test-secret-for-login-test")
	defer auth.SetSecret("")

	repo := newMockUserRepo()
	// Pre-create a user for login tests
	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	repo.users["admin"] = &model.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: string(hashedPw),
		Nickname:     "Admin",
		Status:       1,
	}
	repo.byID[1] = repo.users["admin"]

	svc := NewAuthService(repo, 24)

	tests := []struct {
		name    string
		req     *LoginReq
		wantErr bool
		errType error
	}{
		{
			name: "正常登录",
			req: &LoginReq{
				Username: "admin",
				Password: "correct-password",
			},
			wantErr: false,
		},
		{
			name: "用户不存在",
			req: &LoginReq{
				Username: "nonexistent",
				Password: "password",
			},
			wantErr: true,
			errType: apperrors.ErrInvalidPassword, // spec: user not found → invalid password error
		},
		{
			name: "密码错误",
			req: &LoginReq{
				Username: "admin",
				Password: "wrong-password",
			},
			wantErr: true,
			errType: apperrors.ErrInvalidPassword,
		},
		{
			name: "缺少用户名",
			req: &LoginReq{
				Password: "password",
			},
			wantErr: true,
		},
		{
			name: "缺少密码",
			req: &LoginReq{
				Username: "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Login(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Greater(t, resp.ExpiresAt, int64(0))
				assert.Equal(t, "admin", resp.User.Username)
			}
		})
	}
}

func TestLogin_TokenIsValid(t *testing.T) {
	auth.SetSecret("test-secret-for-token-test")
	defer auth.SetSecret("")

	repo := newMockUserRepo()
	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	repo.users["testuser"] = &model.User{
		ID:           99,
		Username:     "testuser",
		PasswordHash: string(hashedPw),
		Status:       1,
	}
	repo.byID[99] = repo.users["testuser"]

	svc := NewAuthService(repo, 72) // 72 hours expiry

	resp, err := svc.Login(context.Background(), &LoginReq{
		Username: "testuser",
		Password: "pass123",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)

	// Verify the token is parseable and contains correct claims
	claims, err := auth.ParseToken(resp.Token)
	assert.NoError(t, err)
	assert.Equal(t, int64(99), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestRegister_PasswordHashed(t *testing.T) {
	auth.SetSecret("test-secret-hash-check")
	defer auth.SetSecret("")

	svc := NewAuthService(newMockUserRepo(), 24)

	user, err := svc.Register(context.Background(), &RegisterReq{
		Username: "hashtest",
		Password: "my-secret-password",
	})
	assert.NoError(t, err)

	// Plain text password should NOT be stored
	assert.NotEqual(t, "my-secret-password", user.PasswordHash)

	// Hash should be verifiable
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("my-secret-password"))
	assert.NoError(t, err)
}
