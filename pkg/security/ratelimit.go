package security

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"fmt"
	"strconv"
	"time"
)

const (
	// IPRateLimitKeyPrefix Redis中IP限流的key前缀
	IPRateLimitKeyPrefix = "cmdb:ratelimit:ip:"
)

// CheckIPRateLimit 检查IP是否可以访问接口（限流）
func CheckIPRateLimit(ctx context.Context, ipAddress string) error {
	// 获取安全配置
	config, err := GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取安全配置失败: %v", err)
	}

	key := middleware.GetIPRateLimitKey(ipAddress)

	// 获取上次访问时间
	lastAccessStr, err := database.RDB.Get(ctx, key).Result()
	if err == nil {
		// 如果找到了记录，检查时间间隔
		lastAccess, err := strconv.ParseInt(lastAccessStr, 10, 64)
		if err != nil {
			return nil // 如果解析失败，允许访问
		}

		now := time.Now().Unix()
		elapsed := now - lastAccess

		// 如果没有达到最小间隔时间
		if elapsed < int64(config.RateLimitCD) {
			remaining := int64(config.RateLimitCD) - elapsed
			return fmt.Errorf("请求过于频繁，请等待 %d 秒后重试", remaining)
		}
	}

	// 更新访问时间
	now := time.Now().Unix()
	err = database.RDB.Set(ctx, key, strconv.FormatInt(now, 10), time.Duration(config.RateLimitTTL)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("更新访问时间失败: %v", err)
	}

	return nil
}
