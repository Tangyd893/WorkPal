package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// Init 初始化 Redis 客户端
func Init(addr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}

// Client 返回 Redis 客户端
func Client() *redis.Client {
	return rdb
}

// Close 关闭连接
func Close() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}
