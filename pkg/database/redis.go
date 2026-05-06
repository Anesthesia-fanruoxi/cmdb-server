package database

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	// RDB Redis客户端
	RDB *redis.Client
	// RedisNil Redis key 不存在错误
	RedisNil = redis.Nil
)

// InitRedis 初始化Redis连接
func InitRedis(addr, password string, db int) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		return err
	}

	return nil
}
