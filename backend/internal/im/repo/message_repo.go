package repo

import (
	"context"
	"errors"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"gorm.io/gorm"
)

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *MessageRepo) CreateWithTx(tx *gorm.DB, msg *model.Message) error {
	return tx.Create(msg).Error
}

func (r *MessageRepo) GetByIDWithTx(tx *gorm.DB, id int64) (*model.Message, error) {
	var msg model.Message
	if err := tx.First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMessageNotFound
		}
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	var msg model.Message
	if err := r.db.WithContext(ctx).First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMessageNotFound
		}
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepo) GetByConvID(ctx context.Context, convID int64, beforeID int64, limit int) ([]*model.Message, error) {
	var msgs []*model.Message
	query := r.db.WithContext(ctx).
		Where("conv_id = ? AND deleted_at IS NULL", convID)

	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *MessageRepo) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("messages m").
		Joins("LEFT JOIN message_reads mr ON mr.conv_id = m.conv_id AND mr.user_id = ?", userID).
		Where("m.conv_id = ? AND m.sender_id != ? AND m.deleted_at IS NULL AND (mr.read_at IS NULL OR m.created_at > mr.read_at)", convID, userID).
		Count(&count).Error
	return count, err
}

func (r *MessageRepo) MarkRead(ctx context.Context, userID, convID int64) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO message_reads (user_id, conv_id, read_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (user_id, conv_id)
		DO UPDATE SET read_at = EXCLUDED.read_at
	`, userID, convID).Error
}

func (r *MessageRepo) Update(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Save(msg).Error
}

func (r *MessageRepo) UpdateWithTx(tx *gorm.DB, msg *model.Message) error {
	return tx.Save(msg).Error
}

func (r *MessageRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Message{}, id).Error
}

func (r *MessageRepo) SoftDeleteWithTx(tx *gorm.DB, id int64) error {
	return tx.Delete(&model.Message{}, id).Error
}

func (r *MessageRepo) GetBySender(ctx context.Context, senderID int64, offset, limit int) ([]*model.Message, error) {
	var msgs []*model.Message
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND deleted_at IS NULL", senderID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

var ErrMessageNotFound = apperrors.ErrMessageNotFound
