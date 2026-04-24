package repo

import (
	"context"
	"errors"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"gorm.io/gorm"
)

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// Create 创建消息
func (r *MessageRepo) Create(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// CreateWithTx 创建消息（事务）
func (r *MessageRepo) CreateWithTx(tx *gorm.DB, msg *model.Message) error {
	return tx.Create(msg).Error
}

// GetByID 获取消息
func (r *MessageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	var msg model.Message
	if err := r.db.WithContext(ctx).First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &msg, nil
}

// GetByConvID 获取会话消息列表（分页，按时间倒序）
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

// CountUnread 统计未读消息数
func (r *MessageRepo) CountUnread(ctx context.Context, convID, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("messages m").
		Joins("LEFT JOIN message_reads mr ON m.id = mr.msg_id AND mr.user_id = ?", userID).
		Where("m.conv_id = ? AND m.sender_id != ? AND m.deleted_at IS NULL AND mr.msg_id IS NULL", convID, userID).
		Count(&count).Error
	return count, err
}

// MarkRead 标记已读
func (r *MessageRepo) MarkRead(ctx context.Context, userID, convID int64) error {
	// 标记该会话所有消息为已读（删除已读记录再插入，或直接 UPSERT）
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO message_reads (user_id, conv_id, read_at)
		SELECT ? , ?, NOW()
		WHERE NOT EXISTS (
			SELECT 1 FROM message_reads WHERE user_id = ? AND conv_id = ?
		)
	`, userID, convID, userID, convID).Error
}

// Update 更新消息
func (r *MessageRepo) Update(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Save(msg).Error
}

// SoftDelete 软删除（撤回）
func (r *MessageRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Delete(&model.Message{}, id).Error
}

// GetBySender 获取用户发送的消息
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
