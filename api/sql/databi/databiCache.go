package databi

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DataBITablesCache 表列表缓存结构
type DataBITablesCache struct {
	Project    string              `json:"project"`
	Tables     map[string][]string `json:"tables"`      // {"数据库名": ["表1", "表2"]}
	UpdateTime time.Time           `json:"update_time"` // 最后更新时间
}

// RefreshStatus 刷新状态
type RefreshStatus struct {
	Status     string              `json:"status"`      // refreshing, loading, success, error
	Message    string              `json:"message"`     // 状态消息
	Progress   int                 `json:"progress"`    // 进度 0-100
	Tables     map[string][]string `json:"tables"`      // 表数据（仅在 success 时有值）
	UpdateTime time.Time           `json:"update_time"` // 更新时间
	Error      string              `json:"error"`       // 错误信息（仅在 error 时有值）
}

const (
	// 缓存过期时间：24小时
	DataBITablesCacheExpire = 24 * time.Hour
	// 刷新状态过期时间：10分钟
	RefreshStatusExpire = 10 * time.Minute
	// 刷新锁过期时间：5分钟（防止死锁）
	RefreshLockExpire = 5 * time.Minute
)

// SaveTablesToCache 保存表列表到 Redis 缓存
func SaveTablesToCache(ctx context.Context, project string, tables map[string][]string) error {
	cache := DataBITablesCache{
		Project:    project,
		Tables:     tables,
		UpdateTime: time.Now(),
	}

	cacheJSON, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化表列表失败: %v", err)
	}

	key := middleware.GetDataBITablesKey(project)
	return database.RDB.Set(ctx, key, cacheJSON, DataBITablesCacheExpire).Err()
}

// GetTablesFromCache 从 Redis 缓存获取表列表
func GetTablesFromCache(ctx context.Context, project string) (*DataBITablesCache, error) {
	key := middleware.GetDataBITablesKey(project)
	cacheJSON, err := database.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var cache DataBITablesCache
	if err := json.Unmarshal(cacheJSON, &cache); err != nil {
		return nil, fmt.Errorf("解析表列表缓存失败: %v", err)
	}

	return &cache, nil
}

// SaveRefreshStatus 保存刷新状态到 Redis
func SaveRefreshStatus(ctx context.Context, project string, status RefreshStatus) error {
	status.UpdateTime = time.Now()
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("序列化刷新状态失败: %v", err)
	}

	key := middleware.GetDataBIRefreshStatusKey(project)
	return database.RDB.Set(ctx, key, statusJSON, RefreshStatusExpire).Err()
}

// GetRefreshStatus 从 Redis 获取刷新状态
func GetRefreshStatus(ctx context.Context, project string) (*RefreshStatus, error) {
	key := middleware.GetDataBIRefreshStatusKey(project)
	statusJSON, err := database.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var status RefreshStatus
	if err := json.Unmarshal(statusJSON, &status); err != nil {
		return nil, fmt.Errorf("解析刷新状态失败: %v", err)
	}

	return &status, nil
}

// TryAcquireRefreshLock 尝试获取刷新锁（防止并发刷新）
func TryAcquireRefreshLock(ctx context.Context, project string) (bool, error) {
	key := middleware.GetDataBIRefreshLockKey(project)
	// 使用 SetNX：只有键不存在时才设置成功
	success, err := database.RDB.SetNX(ctx, key, "locked", RefreshLockExpire).Result()
	return success, err
}

// ReleaseRefreshLock 释放刷新锁
func ReleaseRefreshLock(ctx context.Context, project string) error {
	key := middleware.GetDataBIRefreshLockKey(project)
	return database.RDB.Del(ctx, key).Err()
}
