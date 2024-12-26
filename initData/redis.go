package initData

import (
	"cmdb/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(Ctx).Result()
	return err
}

// SetToken 将token存入redis
func SetToken(userID uint, token string, expiration time.Duration) error {
	key := fmt.Sprintf("token:%d", userID)
	return RedisClient.Set(Ctx, key, token, expiration).Err()
}

// GetToken 从redis获取token
func GetToken(userID uint) (string, error) {
	key := fmt.Sprintf("token:%d", userID)
	return RedisClient.Get(Ctx, key).Result()
}

// DeleteToken 删除token
func DeleteToken(userID uint) error {
	key := fmt.Sprintf("token:%d", userID)
	return RedisClient.Del(Ctx, key).Err()
}
