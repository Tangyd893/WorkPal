package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if isDuplicateKey(result.Error) {
			return apperrors.ErrUserAlreadyExists
		}
		return result.Error
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	r.db.WithContext(ctx).Model(&model.User{}).Count(&total)
	result := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id desc").Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return users, total, nil
}

// Search 模糊搜索用户（用户名或昵称）
func (r *UserRepo) Search(ctx context.Context, keyword string, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{}).
		Where("deleted_at IS NULL AND (username ILIKE ? OR nickname ILIKE ? OR email ILIKE ?)",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	query.Count(&total)
	result := query.Offset(offset).Limit(limit).Order("id desc").Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return users, total, nil
}

func isDuplicateKey(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "UNIQUE constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 包内哨兵错误（用于测试和跨层判断）
var (
	ErrUserAlreadyExists = apperrors.ErrUserAlreadyExists
	ErrUserNotFound      = apperrors.ErrUserNotFound
	ErrDuplicateKey      = apperrors.ErrUserAlreadyExists
)
