package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"net/http"
	"strings"
	"time"
)

// AlertInfo 告警信息展示结构
type AlertInfo struct {
	Fingerprint string               `json:"fingerprint"`      // 告警唯一标识
	AlertName   string               `json:"alertName"`        // 告警名称
	Summary     string               `json:"summary"`          // 告警摘要
	Status      string               `json:"status"`           // 告警状态
	Severity    string               `json:"severity"`         // 告警级别
	Project     string               `json:"project"`          // 项目
	HostName    string               `json:"hostName"`         // 主机名
	Instance    string               `json:"instance"`         // 实例
	StartTime   string               `json:"startTime"`        // 开始时间
	EndTime     string               `json:"endTime"`          // 恢复时间（如有）
	Detail      *monitor.AlertDetail `json:"detail,omitempty"` // 告警详情（可选）
}

// AlertListResponse 告警列表响应结构
type AlertListResponse struct {
	Total  int         `json:"total"`  // 总告警数
	Alerts []AlertInfo `json:"alerts"` // 告警列表
}

// AlertList 获取告警列表API
func AlertList(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()

	// 获取所有告警键(包括告警状态和告警详情)
	alertKeys, err := database.RDB.Keys(ctx, monitor.AlertRedisKeyPrefix+"*").Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取告警键失败: "+err.Error())
		return
	}

	detailKeys, err := database.RDB.Keys(ctx, monitor.AlertDetailRedisKeyPrefix+"*").Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取告警详情键失败: "+err.Error())
		return
	}

	// 将所有指纹收集到一个集合中
	fingerprintMap := make(map[string]bool)

	for _, key := range alertKeys {
		// 从普通告警键中提取指纹
		fingerprint := strings.TrimPrefix(key, monitor.AlertRedisKeyPrefix)
		fingerprintMap[fingerprint] = true
	}

	for _, key := range detailKeys {
		// 从告警详情键中提取指纹
		fingerprint := strings.TrimPrefix(key, monitor.AlertDetailRedisKeyPrefix)
		fingerprintMap[fingerprint] = true
	}

	// 遍历所有指纹并收集告警信息
	var alerts []AlertInfo

	for fingerprint := range fingerprintMap {
		// 获取告警状态
		alertCache, err := GetAlertFromRedis(fingerprint)
		// 只处理firing状态的告警
		if err == nil && alertCache != nil && alertCache.Status == "firing" {
			// 获取告警详情
			alertDetail, err := GetAlertDetailFromRedis(fingerprint)
			if err != nil {
				middleware.Logger.Warnf("获取告警详情失败: %v, fingerprint: %s", err, fingerprint)
				continue
			}

			// 转换为展示结构
			alertInfo := convertToAlertInfo(*alertDetail, fingerprint)

			// 设置告警状态和时间
			alertInfo.Status = alertCache.Status
			alertInfo.StartTime = alertCache.LastSent.Format(time.RFC3339)

			alerts = append(alerts, alertInfo)
		}
	}

	// 构建响应
	response := &AlertListResponse{
		Total:  len(alerts),
		Alerts: alerts,
	}

	// 返回响应
	middleware.ResponseSuccess(w, "获取告警列表成功", response)
}

// convertToAlertInfo 将AlertDetail转换为AlertInfo
func convertToAlertInfo(alert monitor.AlertDetail, fingerprint string) AlertInfo {
	info := AlertInfo{
		Fingerprint: fingerprint,
		AlertName:   GetAlertName(alert),
		Status:      alert.State,
		StartTime:   alert.ActiveAt,
	}

	// 提取关键标签
	if summary, ok := alert.Annotations["summary"]; ok {
		info.Summary = summary
	}

	if severity, ok := alert.Labels["severity"]; ok {
		info.Severity = severity
	}

	if project, ok := alert.Labels["project"]; ok {
		info.Project = project
	}

	if hostName, ok := alert.Labels["hostName"]; ok {
		info.HostName = hostName
	}

	if instance, ok := alert.Labels["instance"]; ok {
		info.Instance = instance
	}

	return info
}

// GetAlertName 从告警详情中获取告警名称
func GetAlertName(alert monitor.AlertDetail) string {
	// 优先使用标准化后的alertName字段
	if name, ok := alert.Labels["alertName"]; ok {
		return name
	}
	// 其次使用原始的alertname字段
	if name, ok := alert.Labels["alertname"]; ok {
		return name
	}
	return "未知告警"
}
