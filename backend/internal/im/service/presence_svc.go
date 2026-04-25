package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Tangyd893/WorkPal/backend/pkg/cache"
	"github.com/redis/go-redis/v9"
)

const (
	presenceOnlineKey   = "presence:online"
	presenceUserKeyFmt  = "presence:user:%d"
	presenceTimeoutSec  = 60 // 60秒无心跳视为离线
)

type PresenceService struct{}

func NewPresenceService() *PresenceService {
	return &PresenceService{}
}

// SetOnline 设置用户在线（心跳）
func (s *PresenceService) SetOnline(ctx context.Context, userID int64) error {
	rdb := cache.Client()
	now := float64(time.Now().Unix())

	pipe := rdb.Pipeline()
	pipe.ZAdd(ctx, presenceOnlineKey, redis.Z{Score: now, Member: userID})
	pipe.Set(ctx, fmt.Sprintf(presenceUserKeyFmt, userID), "online", time.Duration(presenceTimeoutSec)*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}

// SetOffline 设置用户离线
func (s *PresenceService) SetOffline(ctx context.Context, userID int64) error {
	rdb := cache.Client()
	key := fmt.Sprintf(presenceUserKeyFmt, userID)
	pipe := rdb.Pipeline()
	pipe.ZRem(ctx, presenceOnlineKey, userID)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	return err
}

// IsOnline 检查用户是否在线
func (s *PresenceService) IsOnline(ctx context.Context, userID int64) (bool, error) {
	rdb := cache.Client()
	key := fmt.Sprintf(presenceUserKeyFmt, userID)
	exists, err := rdb.Exists(ctx, key).Result()
	return exists > 0, err
}

// GetOnlineUsers 获取所有在线用户ID
func (s *PresenceService) GetOnlineUsers(ctx context.Context) ([]int64, error) {
	rdb := cache.Client()
	now := time.Now().Unix()
	cutoff := now - int64(presenceTimeoutSec)

	// 清理过期成员
	rdb.ZRemRangeByScore(ctx, presenceOnlineKey, "0", fmt.Sprintf("%d", cutoff))

	// 获取在线用户
	result, err := rdb.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     presenceOnlineKey,
		ByScore: true,
		Start:   fmt.Sprintf("%d", cutoff),
		Stop:    fmt.Sprintf("%d", now),
	}).Result()
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(result))
	for _, m := range result {
		var uid int64
		if _, err := fmt.Sscanf(m, "%d", &uid); err == nil {
			userIDs = append(userIDs, uid)
		}
	}
	return userIDs, nil
}

// CleanupInactive 清理不活跃用户（定期调用）
func (s *PresenceService) CleanupInactive(ctx context.Context) error {
	rdb := cache.Client()
	now := time.Now().Unix()
	cutoff := now - int64(presenceTimeoutSec)
	return rdb.ZRemRangeByScore(ctx, presenceOnlineKey, "0", fmt.Sprintf("%d", cutoff)).Err()
}
