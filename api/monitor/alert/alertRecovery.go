package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"strings"
	"time"
)

// checkRecoveredAlerts 检查并处理已恢复的告警
func checkRecoveredAlerts(activeAlertsMap map[string]bool) {
	redisAlertKeys, err := GetAllAlertKeysFromRedis()

	if err != nil {
		middleware.Logger.Errorf("从Redis获取告警键失败: %v", err)
		return
	}

	// 遍历Redis中的告警键
	for _, key := range redisAlertKeys {
		// 从键中提取指纹
		fingerprint := strings.TrimPrefix(key, monitor.AlertRedisKeyPrefix)

		// 获取告警信息
		cache, err := GetAlertFromRedis(fingerprint)
		if err != nil {
			middleware.Logger.Errorf("从Redis获取告警信息失败: %v", err)
			continue
		}

		// 如果告警状态是firing，但不再出现在活跃列表中，则认为已恢复
		if cache.Status == "firing" && !activeAlertsMap[fingerprint] {
			middleware.Logger.Infof("检测到告警 [%s] 已恢复", fingerprint)

			// 从Redis获取告警详情
			alertDetail, err := GetAlertDetailFromRedis(fingerprint)
			if err != nil {
				middleware.Logger.Warnf("从Redis获取告警详情失败: %v，使用简化信息发送恢复通知", err)
				sendSimpleRecoveryNotification(fingerprint)
			} else {
				sendRecoveryNotification(*alertDetail, fingerprint)
			}
		}
	}
}

// sendSimpleRecoveryNotification 发送简化版恢复通知
func sendSimpleRecoveryNotification(fingerprint string) {
	middleware.Logger.Warnf("未找到告警 [%s] 的详情，使用简化信息发送恢复通知", fingerprint)

	// 构建一个简化的告警详情
	simpleAlert := monitor.AlertDetail{
		Labels: map[string]string{
			"alertName": "未知告警-" + fingerprint[:8],
		},
		Annotations: map[string]string{
			"summary": "告警已恢复，但无法获取原始告警详情",
		},
		State:    "resolved",
		ActiveAt: time.Now().Format(time.RFC3339),
	}

	sendRecoveryNotification(simpleAlert, fingerprint)
}
