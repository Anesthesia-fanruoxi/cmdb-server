package monitor

import (
	"encoding/json"
	"time"
)

// AlertRule 表示一条告警规则
type AlertRule struct {
	Name           string            `json:"name"`
	Query          string            `json:"query"`
	Duration       float64           `json:"duration"`
	KeepFiringFor  float64           `json:"keepFiringFor"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	State          string            `json:"state"`
	Health         string            `json:"health"`
	Type           string            `json:"type"`
	LastError      string            `json:"last_error,omitempty"`
	Alerts         json.RawMessage   `json:"alerts,omitempty"`
	EvaluationTime float64           `json:"evaluationTime,omitempty"`
	LastEvaluation string            `json:"lastEvaluation,omitempty"`
}

// AlertDetail 表示一条活跃告警的详情
type AlertDetail struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"` // 字符串格式的时间，后续手动解析
	Value       string            `json:"value"`
}

// AlertResponse 表示从Prometheus获取的告警数据
type AlertResponse struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []AlertDetail `json:"alerts"`
	} `json:"data"`
}

// AlertCache 用于缓存已发送的告警，避免重复发送
type AlertCache struct {
	Fingerprint string    // 告警指纹
	LastSent    time.Time // 最后发送时间
	Status      string    // 告警状态
}

// 告警抑制时间 (24小时)
const SuppressionDuration = 24 * time.Hour

// 轮询间隔
const PollingInterval = 5 * time.Second

// Redis相关常量
const (
	// AlertRedisKeyPrefix 告警在Redis中的键前缀
	AlertRedisKeyPrefix = "cmdb:alert:"

	// AlertDetailRedisKeyPrefix 告警详情在Redis中的键前缀
	AlertDetailRedisKeyPrefix = "cmdb:alert:detail:"

	// AlertRedisExpireTime 告警信息在Redis中的过期时间 (36小时，大于抑制期)
	AlertRedisExpireTime = 36 * time.Hour
)
