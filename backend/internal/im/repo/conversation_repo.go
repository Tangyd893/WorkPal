package repo

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"gorm.io/gorm"
)

type ConversationRepo struct {
	db *gorm.DB
}

func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// Create 创建会话
func (r *ConversationRepo) Create(ctx context.Context, conv *model.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

// CreateWithTx 创建会话（事务）
func (r *ConversationRepo) CreateWithTx(tx *gorm.DB, conv *model.Conversation) error {
	return tx.Create(conv).Error
}

// GetByID 获取会话
func (r *ConversationRepo) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	var conv model.Conversation
	if err := r.db.WithContext(ctx).First(&conv, id).Error; err != nil {
		if apperrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &conv, nil
}

// FindPrivateConv 查找私聊会话（两人之间的）
func (r *ConversationRepo) FindPrivateConv(ctx context.Context, uid1, uid2 int64) (*model.Conversation, error) {
	var conv model.Conversation
	subQuery := r.db.Table("conversation_members").
		Select("conv_id").
		Where("user_id IN ?", []int64{uid1, uid2}).
		Group("conv_id").
		Having("COUNT(DISTINCT user_id) = 2")

	err := r.db.WithContext(ctx).
		Where("type = ? AND id IN (?)", model.ConversationTypePrivate, subQuery).
		First(&conv).Error
	if err != nil {
		if apperrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

// ListByUser 获取用户的会话列表
func (r *ConversationRepo) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.Conversation, error) {
	var convs []*model.Conversation
	err := r.db.WithContext(ctx).
		Table("conversations c").
		Select("c.*").
		Joins("JOIN conversation_members cm ON c.id = cm.conv_id").
		Where("cm.user_id = ? AND c.deleted_at IS NULL", userID).
		Order("c.updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&convs).Error
	return convs, err
}

// CountByUser 统计用户会话数
func (r *ConversationRepo) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("conversations c").
		Joins("JOIN conversation_members cm ON c.id = cm.conv_id").
		Where("cm.user_id = ? AND c.deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// Update 更新会话
func (r *ConversationRepo) Update(ctx context.Context, conv *model.Conversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

// Delete 删除会话（软删除）
func (r *ConversationRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Conversation{}, id).Error
}

// AddMemberTx 添加成员（可选事务）
func (r *ConversationRepo) AddMemberTx(ctx context.Context, tx *gorm.DB, convID, userID int64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	member := &model.ConversationMember{
		ConvID:   convID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return db.Create(member).Error
}

// RemoveMember 移除成员
func (r *ConversationRepo) RemoveMember(ctx context.Context, convID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("conv_id = ? AND user_id = ?", convID, userID).
		Delete(&model.ConversationMember{}).Error
}

// AddMember 添加成员
func (r *ConversationRepo) AddMember(ctx context.Context, member *model.ConversationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMembers 获取会话成员
func (r *ConversationRepo) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	var userIDs []int64
	err := r.db.WithContext(ctx).
		Table("conversation_members").
		Select("user_id").
		Where("conv_id = ?", convID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// IsMember 是否成员
func (r *ConversationRepo) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("conversation_members").
		Where("conv_id = ? AND user_id = ?", convID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetLastMessage 获取会话最后一条消息
func (r *ConversationRepo) GetLastMessage(ctx context.Context, convID int64) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).
		Where("conv_id = ? AND deleted_at IS NULL", convID).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		if apperrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}
