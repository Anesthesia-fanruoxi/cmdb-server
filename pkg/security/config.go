package security

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	ID             int64 `gorm:"primarykey"`
	BlacklistCount int   `gorm:"column:blacklist_count;not null;default:5" json:"blacklist_count"` // 黑名单最大错误次数
	BlacklistLock  int   `gorm:"column:blacklist_lock;not null;default:30" json:"blacklist_lock"`  // 黑名单锁定时间(分钟)
	RateLimitCD    int   `gorm:"column:ratelimit_cd;not null;default:60" json:"ratelimit_cd"`      // 限流冷却时间(秒)
	RateLimitTTL   int   `gorm:"column:ratelimit_ttl;not null;default:3600" json:"ratelimit_ttl"`  // Redis存储时间(秒)
}

// TableName 设置表名
func (SecurityConfig) TableName() string {
	return middleware.TableSecurity
}

// GetConfig 获取配置
func GetConfig(ctx context.Context) (*SecurityConfig, error) {
	// 先从Redis获取
	key := middleware.KeyPathSecurityConfig
	data, err := database.RDB.Get(ctx, key).Bytes()
	if err == nil {
		var config SecurityConfig
		if err := json.Unmarshal(data, &config); err == nil {
			return &config, nil
		}
	}

	// Redis获取失败，从数据库获取
	var config SecurityConfig
	if err := database.DB.First(&config).Error; err != nil {
		return nil, err
	}

	// 缓存到Redis
	if err := cacheConfig(ctx, &config); err != nil {
		// 缓存失败只记录日志，不影响返回
		fmt.Printf("缓存配置失败: %v\n", err)
	}

	return &config, nil
}

// UpdateConfig 更新安全配置
func UpdateConfig(ctx context.Context, config *SecurityConfig) error {
	// 更新数据库
	result := database.DB.Model(&SecurityConfig{}).Where("id = ?", 1).UpdateColumns(map[string]interface{}{
		"blacklist_count": config.BlacklistCount,
		"blacklist_lock":  config.BlacklistLock,
		"ratelimit_cd":    config.RateLimitCD,
		"ratelimit_ttl":   config.RateLimitTTL,
	})
	if result.Error != nil {
		return fmt.Errorf("更新安全配置失败: %v", result.Error)
	}

	// 更新Redis缓存
	key := middleware.KeyPathSecurityConfig
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := database.RDB.Set(ctx, key, jsonData, time.Duration(config.RateLimitTTL)*time.Second).Err(); err != nil {
		return fmt.Errorf("更新Redis缓存失败: %v", err)
	}

	return nil
}

// cacheConfig 缓存配置到Redis
func cacheConfig(ctx context.Context, config *SecurityConfig) error {
	key := middleware.KeyPathSecurityConfig
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return database.RDB.Set(ctx, key, data, time.Duration(config.RateLimitTTL)*time.Second).Err()
}
