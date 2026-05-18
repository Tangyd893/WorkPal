package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"gorm.io/gorm"
)

type ChannelRepo struct {
	db *gorm.DB
}

func NewChannelRepo(db *gorm.DB) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (r *ChannelRepo) List(ctx context.Context) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&channels).Error
	return channels, err
}

func (r *ChannelRepo) Create(ctx context.Context, ch *model.Channel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *ChannelRepo) GetByID(ctx context.Context, id int64) (*model.Channel, error) {
	var ch model.Channel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Channel{}).Error
}

func (r *ChannelRepo) AddMember(ctx context.Context, m *model.ChannelMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ChannelRepo) RemoveMember(ctx context.Context, channelID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&model.ChannelMember{}).Error
}

func (r *ChannelRepo) ListMembers(ctx context.Context, channelID int64) ([]*model.ChannelMember, error) {
	var members []*model.ChannelMember
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Find(&members).Error
	return members, err
}

func (r *ChannelRepo) IsMember(ctx context.Context, channelID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ChannelRepo) ListThreads(ctx context.Context, channelID int64) ([]*model.Thread, error) {
	var threads []*model.Thread
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Find(&threads).Error
	return threads, err
}

func (r *ChannelRepo) CreateThread(ctx context.Context, t *model.Thread) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *ChannelRepo) GetThreadByID(ctx context.Context, id int64) (*model.Thread, error) {
	var thread model.Thread
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&thread).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *ChannelRepo) IncrementReplyCount(ctx context.Context, threadID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Thread{}).
		Where("id = ?", threadID).
		UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).Error
}
