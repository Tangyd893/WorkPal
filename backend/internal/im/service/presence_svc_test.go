package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestPresenceService(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis not available: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := NewPresenceServiceWithClient(rdb)
	ctx := context.Background()

	t.Run("SetOnline 设置在线", func(t *testing.T) {
		err := svc.SetOnline(ctx, 1)
		assert.NoError(t, err)

		online, err := svc.IsOnline(ctx, 1)
		assert.NoError(t, err)
		assert.True(t, online)
	})

	t.Run("SetOffline 设置离线", func(t *testing.T) {
			err := svc.SetOnline(ctx, 2)
		assert.NoError(t, err)
		err = svc.SetOffline(ctx, 2)
		assert.NoError(t, err)

		online, err := svc.IsOnline(ctx, 2)
		assert.NoError(t, err)
		assert.False(t, online)
	})

	t.Run("GetOnlineUsers 批量获取在线用户", func(t *testing.T) {
		_ = svc.SetOnline(ctx, 10)
		_ = svc.SetOnline(ctx, 11)
		_ = svc.SetOnline(ctx, 12)

		users, err := svc.GetOnlineUsers(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 3)
	})

	t.Run("CleanupInactive 清理过期成员", func(t *testing.T) {
		_ = svc.SetOnline(ctx, 20)
		// CleanupInactive 只清理 sorted set 中的过期 score
		// （实际 TTL 过期由 Redis 独立处理）
		err := svc.CleanupInactive(ctx)
		assert.NoError(t, err)
	})

	_ = rdb.Close()
}
