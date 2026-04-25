package service

import (
	"context"
	"errors"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	apierr "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	auth "github.com/Tangyd893/WorkPal/backend/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

// 预定义业务错误
var (
	ErrUserNotFound      = apierr.New(40401, "用户不存在")
	ErrUserAlreadyExists = apierr.New(40901, "用户名已存在")
	ErrInvalidPassword   = apierr.New(40101, "用户名或密码错误")
)

// UserRepository 接口，便于测试时注入 mock
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type AuthService struct {
	userRepo      UserRepository
	jwtExpiryHours int
}

// NewAuthService 注入 *repo.UserRepo（生产）或 mock（测试）
func NewAuthService(userRepo UserRepository, jwtExpiryHours int) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		jwtExpiryHours: jwtExpiryHours,
	}
}

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname" binding:"max=128"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expires_at"`
	User      *model.User  `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req *RegisterReq) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierr.Wrap(50001, "密码加密失败", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Email:        req.Email,
		Status:       1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, apierr.ErrUserAlreadyExists) {
			return nil, err
		}
		if appErr, ok := err.(*apierr.AppError); ok {
			return nil, appErr
		}
		return nil, apierr.Wrap(50002, "创建用户失败", err)
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, apierr.ErrUserNotFound) {
			return nil, ErrInvalidPassword
		}
		if appErr, ok := err.(*apierr.AppError); ok {
			return nil, appErr
		}
		return nil, apierr.Wrap(50003, "登录失败", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrInvalidPassword
	}

	token, err := auth.GenerateToken(user.ID, user.Username, s.jwtExpiryHours)
	if err != nil {
		return nil, apierr.Wrap(50004, "生成 Token 失败", err)
	}

	return &LoginResp{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(s.jwtExpiryHours) * time.Hour).Unix(),
		User:     user,
	}, nil
}
