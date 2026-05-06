package security

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// CheckLoginSecurity 检查登录安全
func CheckLoginSecurity(user_name string) error {
	// 1. 获取系统设置
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		return fmt.Errorf("获取系统设置失败: %v", err)
	}

	// 2. 检查是否开启登录保护
	if !setting.LoginProtection {
		return nil // 如果没有开启登录保护，直接返回
	}

	// 3. 获取Redis中的失败次数
	ctx := context.Background()
	key := middleware.GetLoginFailedKey(user_name)
	count, err := database.RDB.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("获取登录失败次数失败: %v", err)
	}

	// 4. 如果失败次数超过限制
	if count >= setting.LoginFailCount {
		// 获取锁定剩余时间
		ttl, err := database.RDB.TTL(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("获取锁定时间失败: %v", err)
		}

		remainingMinutes := int(ttl.Minutes())
		if remainingMinutes > 0 {
			return fmt.Errorf("账户已被锁定，请在%d分钟后重试", remainingMinutes)
		}

		// 如果锁定时间已过，重置计数器
		if err := database.RDB.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("重置登录失败次数失败: %v", err)
		}
	}

	return nil
}

// IncrementLoginFailed 增加登录失败次数
func IncrementLoginFailed(user_name string) error {
	// 1. 获取系统设置
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		return fmt.Errorf("获取系统设置失败: %v", err)
	}

	// 2. 检查是否开启登录保护
	if !setting.LoginProtection {
		return nil
	}

	ctx := context.Background()
	key := middleware.GetLoginFailedKey(user_name)

	// 3. 增加失败次数
	count, err := database.RDB.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("增加登录失败次数失败: %v", err)
	}

	// 4. 如果是第一次失败，设置过期时间
	if count == 1 {
		err = database.RDB.Expire(ctx, key, time.Duration(setting.LoginLockTime)*time.Minute).Err()
		if err != nil {
			return fmt.Errorf("设置过期时间失败: %v", err)
		}
	}

	return nil
}

// ResetLoginFailed 重置登录失败次数
func ResetLoginFailed(user_name string) error {
	ctx := context.Background()
	key := middleware.GetLoginFailedKey(user_name)
	return database.RDB.Del(ctx, key).Err()
}
