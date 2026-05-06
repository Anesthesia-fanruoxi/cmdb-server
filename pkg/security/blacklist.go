package security

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"fmt"
	"time"
)

// CheckIPBlacklist 检查IP是否在黑名单中
func CheckIPBlacklist(ipAddress string) error {
	ctx := context.Background()
	key := middleware.GetIPBlacklistKey(ipAddress)

	// 检查IP是否在黑名单中
	exists, err := database.RDB.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("检查IP黑名单失败: %v", err)
	}

	if exists == 1 {
		// 获取封禁剩余时间
		ttl, err := database.RDB.TTL(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("获取IP封禁时间失败: %v", err)
		}

		remainingMinutes := int(ttl.Minutes())
		if remainingMinutes > 0 {
			return fmt.Errorf("您的IP已被临时封禁，剩余%d分钟", remainingMinutes)
		} else {
			return fmt.Errorf("您的IP已被临时封禁")
		}
	}

	return nil
}

// IncrementIPLoginFailed 增加IP登录失败次数
func IncrementIPLoginFailed(ipAddress string) error {
	// 获取安全配置
	config, err := GetConfig(context.Background())
	if err != nil {
		return fmt.Errorf("获取安全配置失败: %v", err)
	}

	ctx := context.Background()
	key := middleware.GetIPLoginFailedKey(ipAddress)

	// 检查是否存在失败记录
	exists, err := database.RDB.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("检查失败记录失败: %v", err)
	}

	var count int64
	if exists == 0 {
		// 如果不存在，设置为1
		count = 1
		err = database.RDB.Set(ctx, key, count, time.Duration(config.RateLimitTTL)*time.Second).Err()
	} else {
		// 如果存在，增加1
		count, err = database.RDB.Incr(ctx, key).Result()
	}
	if err != nil {
		return fmt.Errorf("更新失败次数失败: %v", err)
	}

	// 如果达到最大失败次数，加入黑名单
	if count >= int64(config.BlacklistCount) {
		blacklistKey := middleware.GetIPBlacklistKey(ipAddress)
		// 将锁定时间从分钟转换为小时
		if err := database.RDB.Set(ctx, blacklistKey, "blocked", time.Duration(config.BlacklistLock)*time.Hour).Err(); err != nil {
			return fmt.Errorf("加入黑名单失败: %v", err)
		}
	}

	return nil
}

// ResetIPLoginFailed 重置IP登录失败次数
func ResetIPLoginFailed(ipAddress string) error {
	ctx := context.Background()
	key := middleware.GetIPLoginFailedKey(ipAddress)
	return database.RDB.Del(ctx, key).Err()
}

// RemoveIPFromBlacklist 从黑名单中移除IP（用于管理接口）
func RemoveIPFromBlacklist(ipAddress string) error {
	ctx := context.Background()
	key := middleware.GetIPBlacklistKey(ipAddress)
	return database.RDB.Del(ctx, key).Err()
}
