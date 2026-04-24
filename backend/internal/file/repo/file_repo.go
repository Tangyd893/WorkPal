package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/file/model"
	"gorm.io/gorm"
)

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

// Create 创建文件记录
func (r *FileRepo) Create(ctx context.Context, f *model.File) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// GetByID 获取文件
func (r *FileRepo) GetByID(ctx context.Context, id int64) (*model.File, error) {
	var f model.File
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// GetByKey 通过 key 获取
func (r *FileRepo) GetByKey(ctx context.Context, key string) (*model.File, error) {
	var f model.File
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// ListByUser 获取用户文件列表
func (r *FileRepo) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.File, error) {
	var files []*model.File
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	return files, err
}

// ListByConv 获取会话文件列表
func (r *FileRepo) ListByConv(ctx context.Context, convID int64, offset, limit int) ([]*model.File, error) {
	var files []*model.File
	err := r.db.WithContext(ctx).
		Where("conv_id = ?", convID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&files).Error
	return files, err
}

// Delete 删除文件记录
func (r *FileRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.File{}, id).Error
}
