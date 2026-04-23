package service

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/Tangyd893/WorkPal/backend/internal/user/repo"
)

type UserService struct {
	userRepo *repo.UserRepo
}

func NewUserService(userRepo *repo.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateProfileReq struct {
	Nickname  string `json:"nickname" binding:"max=128"`
	AvatarURL string `json:"avatar_url" binding:"max=512"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone" binding:"max=32"`
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *UpdateProfileReq) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, offset, pageSize)
}

func (s *UserService) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error) {
	// TODO: 实现模糊搜索
	return s.ListUsers(ctx, page, pageSize)
}

// GetCurrentUser 获取当前登录用户信息（从 context 的 userID）
func (s *UserService) GetCurrentUser(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
