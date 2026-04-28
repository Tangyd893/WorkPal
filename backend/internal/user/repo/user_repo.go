package repo

import (
	"context"
	"errors"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"gorm.io/gorm"
)

type DirectoryFilter struct {
	Query        string
	DepartmentID int64
	Offset       int
	Limit        int
}

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
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
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
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetDirectoryByID(ctx context.Context, id int64) (*model.DirectoryUser, error) {
	var user model.DirectoryUser
	err := r.directoryBaseQuery(ctx).
		Where("u.id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
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

func (r *UserRepo) ListDirectoryUsers(ctx context.Context, filter DirectoryFilter) ([]*model.DirectoryUser, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	query := r.applyDirectoryFilters(r.directoryBaseQuery(ctx), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*model.DirectoryUser
	err := query.
		Order("u.nickname ASC, u.username ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepo) ListDepartments(ctx context.Context) ([]*model.Department, error) {
	var departments []*model.Department
	err := r.db.WithContext(ctx).
		Model(&model.Department{}).
		Order("name ASC").
		Find(&departments).Error
	if err != nil {
		return nil, err
	}
	return departments, nil
}

// Search retains the lightweight user lookup used by existing tests and handlers.
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

func (r *UserRepo) directoryBaseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("users AS u").
		Select(`
			u.id,
			u.username,
			u.nickname,
			u.avatar_url,
			u.email,
			u.phone,
			u.status,
			COALESCE(e.department_id, u.department_id, 0) AS department_id,
			COALESCE(d.name, '') AS department_name,
			COALESCE(e.id, 0) AS employee_id,
			COALESCE(e.employee_no, '') AS employee_no,
			COALESCE(e.job_title, '') AS job_title,
			COALESCE(e.office_location, '') AS office_location,
			COALESCE(e.bio, '') AS bio,
			u.created_at,
			u.updated_at
		`).
		Joins("LEFT JOIN employees e ON e.user_id = u.id AND e.deleted_at IS NULL").
		Joins("LEFT JOIN departments d ON d.id = COALESCE(e.department_id, u.department_id)").
		Where("u.deleted_at IS NULL")
}

func (r *UserRepo) applyDirectoryFilters(query *gorm.DB, filter DirectoryFilter) *gorm.DB {
	if filter.DepartmentID > 0 {
		query = query.Where("COALESCE(e.department_id, u.department_id, 0) = ?", filter.DepartmentID)
	}

	if filter.Query != "" {
		likeValue := "%" + filter.Query + "%"
		query = query.Where(`
			u.username ILIKE ? OR
			u.nickname ILIKE ? OR
			u.email ILIKE ? OR
			u.phone ILIKE ? OR
			e.employee_no ILIKE ? OR
			e.job_title ILIKE ? OR
			e.office_location ILIKE ? OR
			d.name ILIKE ?
		`, likeValue, likeValue, likeValue, likeValue, likeValue, likeValue, likeValue, likeValue)
	}

	return query
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

var (
	ErrUserAlreadyExists = apperrors.ErrUserAlreadyExists
	ErrUserNotFound      = apperrors.ErrUserNotFound
	ErrDuplicateKey      = apperrors.ErrUserAlreadyExists
)
