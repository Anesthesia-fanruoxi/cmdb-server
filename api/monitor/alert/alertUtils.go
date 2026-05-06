package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/database"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ============================ 规范化函数 ============================

// StandardizeAlertName 标准化告警名称（从alertname转换为alertName）
func StandardizeAlertName(alert *monitor.AlertDetail) {
	// 检查是否存在小写的alertname
	if name, ok := alert.Labels["alertname"]; ok {
		// 添加驼峰式命名的alertName
		alert.Labels["alertName"] = name
		// 可选：删除原始的alertname（如果想保持纯净）
		// delete(alert.Labels, "alertname")
	}
}

// StandardizeAlertValue 标准化告警值（从科学计数法转为十进制）
func StandardizeAlertValue(alert *monitor.AlertDetail) {
	if alert.Value != "" {
		valueFloat, err := strconv.ParseFloat(alert.Value, 64)
		if err == nil {
			// 使用标准十进制格式（保留6位小数）
			alert.Value = strconv.FormatFloat(valueFloat, 'f', 6, 64)
		}
	}
}

// StandardizeAlert 综合标准化告警（处理名称和值）
func StandardizeAlert(alert *monitor.AlertDetail) {
	StandardizeAlertName(alert)
	StandardizeAlertValue(alert)
}

// ============================ Redis键管理 ============================

// GetRedisAlertKey 获取告警在Redis中的键
func GetRedisAlertKey(fingerprint string) string {
	return monitor.AlertRedisKeyPrefix + fingerprint
}

// GetRedisAlertDetailKey 获取告警详情在Redis中的键
func GetRedisAlertDetailKey(fingerprint string) string {
	return monitor.AlertDetailRedisKeyPrefix + fingerprint
}

// ============================ 指纹生成 ============================

// GenerateAlertFingerprint 生成告警的唯一指纹
func GenerateAlertFingerprint(alert monitor.AlertDetail) string {
	// 收集所有标签的键，并排序
	keys := make([]string, 0, len(alert.Labels))
	for k := range alert.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建指纹字符串
	var sb strings.Builder

	// 添加告警名称 (优先使用alertName)
	if name, ok := alert.Labels["alertName"]; ok {
		sb.WriteString(name)
	} else if name, ok := alert.Labels["alertname"]; ok {
		sb.WriteString(name)
	}

	// 添加排序后的标签
	for _, k := range keys {
		if k != "alertName" && k != "alertname" { // 避免重复添加alertname
			sb.WriteString(fmt.Sprintf("%s=%s;", k, alert.Labels[k]))
		}
	}

	// 计算MD5哈希值作为指纹
	hash := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// ============================ Redis存取操作 ============================

// SaveAlertToRedis 将告警信息保存到Redis
func SaveAlertToRedis(fingerprint string, status string) error {
	ctx := context.Background()
	key := GetRedisAlertKey(fingerprint)

	// 创建告警缓存对象
	cache := &monitor.AlertCache{
		Fingerprint: fingerprint,
		LastSent:    time.Now(),
		Status:      status,
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化告警缓存失败: %v", err)
	}

	// 保存到Redis
	return database.RDB.Set(ctx, key, jsonData, monitor.AlertRedisExpireTime).Err()
}

// SaveAlertDetailToRedis 将告警详情保存到Redis
func SaveAlertDetailToRedis(fingerprint string, alert monitor.AlertDetail) error {
	ctx := context.Background()
	key := GetRedisAlertDetailKey(fingerprint)

	// 序列化为JSON
	jsonData, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("序列化告警详情失败: %v", err)
	}

	// 保存到Redis
	return database.RDB.Set(ctx, key, jsonData, monitor.AlertRedisExpireTime).Err()
}

// GetAlertFromRedis 从Redis获取告警信息
func GetAlertFromRedis(fingerprint string) (*monitor.AlertCache, error) {
	ctx := context.Background()
	key := GetRedisAlertKey(fingerprint)

	// 从Redis获取告警信息
	data, err := database.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化
	var cache monitor.AlertCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("反序列化告警缓存失败: %v", err)
	}

	return &cache, nil
}

// GetAlertDetailFromRedis 从Redis获取告警详情
func GetAlertDetailFromRedis(fingerprint string) (*monitor.AlertDetail, error) {
	ctx := context.Background()
	key := GetRedisAlertDetailKey(fingerprint)

	// 从Redis获取告警详情
	data, err := database.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	// 反序列化
	var alert monitor.AlertDetail
	if err := json.Unmarshal(data, &alert); err != nil {
		return nil, fmt.Errorf("反序列化告警详情失败: %v", err)
	}

	return &alert, nil
}

// GetAllAlertKeysFromRedis 获取所有告警键
func GetAllAlertKeysFromRedis() ([]string, error) {
	ctx := context.Background()
	pattern := monitor.AlertRedisKeyPrefix + "*"

	// 获取所有匹配的键
	keys, err := database.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// ============================ 时间处理 ============================

// ParseUTCTime 将UTC格式时间字符串转换为time.Time
func ParseUTCTime(utcTimeStr string) (time.Time, error) {
	// 尝试多种格式解析UTC时间
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, utcTimeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", utcTimeStr)
}

// ============================ 告警状态管理 ============================

// IsSuppressed 检查告警是否在抑制期内
func IsSuppressed(fingerprint string, status string) bool {
	// 从Redis获取告警信息
	cache, err := GetAlertFromRedis(fingerprint)
	if err != nil {
		// Redis中没有找到，表示首次出现，不受抑制
		return false
	}

	// 如果状态不同，不受抑制
	if cache.Status != status {
		return false
	}

	return time.Since(cache.LastSent) < monitor.SuppressionDuration
}

// IsStatusChanged 检查告警状态是否已变更
func IsStatusChanged(fingerprint string, newStatus string) bool {
	// 从Redis获取告警信息
	cache, err := GetAlertFromRedis(fingerprint)
	if err != nil {
		return true // 首次出现视为状态变更
	}

	return cache.Status != newStatus
}
