package repo

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"gorm.io/gorm"
)

type ConversationRepo struct {
	db *gorm.DB
}

func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

func (r *ConversationRepo) Create(ctx context.Context, conv *model.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *ConversationRepo) CreateWithTx(tx *gorm.DB, conv *model.Conversation) error {
	return tx.Create(conv).Error
}

func (r *ConversationRepo) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	var conv model.Conversation
	if err := r.db.WithContext(ctx).First(&conv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrConversationNotFound
		}
		return nil, err
	}
	return &conv, nil
}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

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

func (r *ConversationRepo) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("conversations c").
		Joins("JOIN conversation_members cm ON c.id = cm.conv_id").
		Where("cm.user_id = ? AND c.deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *ConversationRepo) Update(ctx context.Context, conv *model.Conversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

func (r *ConversationRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Conversation{}, id).Error
}

func (r *ConversationRepo) AddMemberTx(ctx context.Context, tx *gorm.DB, convID, userID int64) error {
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}
	member := &model.ConversationMember{
		ConvID:   convID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return db.Create(member).Error
}

func (r *ConversationRepo) RemoveMember(ctx context.Context, convID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("conv_id = ? AND user_id = ?", convID, userID).
		Delete(&model.ConversationMember{}).Error
}

func (r *ConversationRepo) AddMember(ctx context.Context, member *model.ConversationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *ConversationRepo) GetMembers(ctx context.Context, convID int64) ([]int64, error) {
	var userIDs []int64
	err := r.db.WithContext(ctx).
		Table("conversation_members").
		Where("conv_id = ?", convID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

func (r *ConversationRepo) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("conversation_members").
		Where("conv_id = ? AND user_id = ?", convID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ConversationRepo) GetLastMessage(ctx context.Context, convID int64) (*model.Message, error) {
	var msg model.Message
	err := r.db.WithContext(ctx).
		Where("conv_id = ? AND deleted_at IS NULL", convID).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

var ErrConvNotFound = apperrors.ErrConversationNotFound
