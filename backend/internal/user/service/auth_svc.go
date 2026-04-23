package service

import (
	"context"
	"errors"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	apierr "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

// 预定义业务错误
var (
	ErrUserNotFound      = apierr.New(40401, "用户不存在")
	ErrUserAlreadyExists = apierr.New(40901, "用户名已存在")
	ErrInvalidPassword   = apierr.New(40101, "用户名或密码错误")
)

type AuthService struct {
	userRepo      *repo.UserRepo
	jwtExpiryHours int
}

func NewAuthService(userRepo *repo.UserRepo, jwtExpiryHours int) *AuthService {
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
