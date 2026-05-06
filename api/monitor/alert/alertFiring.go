package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
)

// processAlerts 处理获取到的告警
func processAlerts(alerts []monitor.AlertDetail) {
	// 创建活跃告警指纹映射，用于后续检测恢复的告警
	activeAlertsMap := make(map[string]bool)

	// 遍历所有告警
	for _, alert := range alerts {
		if alert.State == "firing" {
			// 生成告警指纹
			fingerprint := GenerateAlertFingerprint(alert)
			activeAlertsMap[fingerprint] = true

			// 检查是否在抑制期内
			if IsSuppressed(fingerprint, "firing") {
				middleware.Logger.Debugf("告警 [%s] 在抑制期内，跳过处理", fingerprint)
				continue
			}

			// 检查状态是否已更改
			if !IsStatusChanged(fingerprint, "firing") {
				middleware.Logger.Debugf("告警 [%s] 状态未变更，跳过处理", fingerprint)
				continue
			}

			// 发送告警通知并标记为已发送
			sendFiringNotification(alert, fingerprint)
		}
	}

	// 检查已恢复的告警
	checkRecoveredAlerts(activeAlertsMap)
}

// updateAlertDetailsCache 更新告警详情缓存
func updateAlertDetailsCache(alerts []monitor.AlertDetail) {
	// 过滤出firing状态的告警，只缓存这些告警
	for _, alert := range alerts {
		if alert.State == "firing" {
			// 为每个告警生成指纹并添加到Redis缓存中
			fingerprint := GenerateAlertFingerprint(alert)

			// 将告警详情存储到Redis
			if err := SaveAlertDetailToRedis(fingerprint, alert); err != nil {
				middleware.Logger.Errorf("保存告警详情到Redis失败: %v", err)
			}
		}
	}
}

// sendFiringNotification 发送告警通知
func sendFiringNotification(alert monitor.AlertDetail, fingerprint string) {
	// 获取告警名称
	alertName := GetAlertName(alert)

	// 记录日志
	middleware.Logger.Infof("发送告警通知: %s [%s]", alertName, fingerprint)

	// 发送飞书通知
	sendAlertNotification(alert, fingerprint, "firing")

	// 更新Redis中的告警状态
	markSent(fingerprint, "firing")
}

// markSent 标记告警已发送
func markSent(fingerprint string, status string) {
	// 保存到Redis
	if err := SaveAlertToRedis(fingerprint, status); err != nil {
		middleware.Logger.Errorf("保存告警状态到Redis失败: %v", err)
	}
}
