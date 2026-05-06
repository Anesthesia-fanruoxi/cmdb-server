package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"cmdb/pkg/notice"
	"fmt"
)

// getAlertName 获取告警名称，用于日志输出
func getAlertName(alert monitor.AlertDetail) string {
	return GetAlertName(alert)
}

// 构建飞书告警卡片
func buildFeiShuCard(alert monitor.AlertDetail) map[string]interface{} {
	// 获取告警名称
	alertName := "未知告警"
	if name, ok := alert.Labels["alertname"]; ok {
		alertName = name
	}

	// 获取告警摘要和描述
	summary := "无摘要信息"
	if sum, ok := alert.Annotations["summary"]; ok {
		summary = sum
	}

	description := "无详细描述"
	if desc, ok := alert.Annotations["description"]; ok {
		description = desc
	}

	// 获取告警级别
	severityColor := "orange"
	severityText := "警告"
	if severity, ok := alert.Labels["severity"]; ok {
		switch severity {
		case "critical":
			severityColor = "red"
			severityText = "严重"
		case "warning":
			severityColor = "orange"
			severityText = "警告"
		case "info":
			severityColor = "blue"
			severityText = "信息"
		}
	}

	// 提取重要标签
	hostName := ""
	if h, ok := alert.Labels["hostName"]; ok {
		hostName = h
	}

	instance := ""
	if i, ok := alert.Labels["instance"]; ok {
		instance = i
	}

	job := ""
	if j, ok := alert.Labels["job"]; ok {
		job = j
	}

	project := ""
	if p, ok := alert.Labels["project"]; ok {
		project = p
	}

	// 格式化触发时间（转换为中国时区）
	var triggerTime string
	if t, err := ParseUTCTime(alert.ActiveAt); err == nil {
		// 使用middleware包中的FormatTimeToChina函数
		triggerTime = middleware.FormatTimeToChina(t)
	} else {
		triggerTime = "未知时间"
		middleware.Logger.Errorf("解析告警时间失败: %v, 原始时间字符串: %s", err, alert.ActiveAt)
	}

	// 构建额外标签显示
	var extraLabels []map[string]interface{}
	for k, v := range alert.Labels {
		// 跳过已经单独显示的标签
		if k != "alertName" && k != "severity" &&
			k != "hostName" && k != "instance" &&
			k != "job" && k != "project" {
			extraLabels = append(extraLabels, map[string]interface{}{
				"key":   k,
				"value": v,
			})
		}
	}

	// 添加当前值（如果存在）
	if alert.Value != "" {
		extraLabels = append(extraLabels, map[string]interface{}{
			"key":   "当前值",
			"value": alert.Value,
		})
	}

	// 构建卡片内容
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("【%s】%s", severityText, alertName),
			},
			"template": severityColor,
		},
		"elements": []interface{}{
			// 摘要部分
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": summary,
				},
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**告警时间**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": triggerTime,
						},
					},
				},
			},
			// 分割线
			map[string]interface{}{
				"tag": "hr",
			},
			// 关键信息部分
			map[string]interface{}{
				"tag": "div",
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**主机名**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": hostName,
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**项目**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": project,
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**实例**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": instance,
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**任务**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": job,
						},
					},
				},
			},
			// 描述部分
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": "**详细描述**",
				},
			},
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": description,
				},
			},
		},
	}

	// 添加额外标签（如果有）
	if len(extraLabels) > 0 {
		var fields []map[string]interface{}
		for _, label := range extraLabels {
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": "**" + label["key"].(string) + "**",
				},
			})
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": label["value"].(string),
				},
			})
		}

		if len(fields) > 0 {
			card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
				"tag": "hr",
			})
			card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": "**其他信息**",
				},
			})
			card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
				"tag":    "div",
				"fields": fields,
			})
		}
	}

	// 添加底部提示
	card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
		"tag": "hr",
	})
	card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
		"tag": "note",
		"elements": []map[string]interface{}{
			{
				"tag":     "plain_text",
				"content": "此告警通知在24小时内不会重复发送",
			},
		},
	})

	return card
}

// buildFeiShuRecoveryCard 构建告警恢复的飞书卡片
func buildFeiShuRecoveryCard(alert monitor.AlertDetail) map[string]interface{} {
	// 获取告警名称
	alertName := "未知告警"
	if name, ok := alert.Labels["alertname"]; ok {
		alertName = name
	}

	// 获取告警摘要和描述
	summary := "无摘要信息"
	if sum, ok := alert.Annotations["summary"]; ok {
		summary = sum
	}

	// 提取重要标签
	hostName := ""
	if h, ok := alert.Labels["hostName"]; ok {
		hostName = h
	}

	project := ""
	if p, ok := alert.Labels["project"]; ok {
		project = p
	}

	// 格式化恢复时间
	var recoveryTime string
	if t, err := ParseUTCTime(alert.ActiveAt); err == nil {
		// 使用middleware包中的FormatTimeToChina函数
		recoveryTime = middleware.FormatTimeToChina(t)
	} else {
		recoveryTime = "未知时间"
		middleware.Logger.Errorf("解析告警时间失败: %v, 原始时间字符串: %s", err, alert.ActiveAt)
	}

	// 构建恢复卡片内容
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("【已恢复】%s", alertName),
			},
			"template": "green", // 恢复使用绿色
		},
		"elements": []interface{}{
			// 摘要部分
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": summary + " (已恢复)",
				},
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**恢复时间**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": recoveryTime,
						},
					},
				},
			},
			// 分割线
			map[string]interface{}{
				"tag": "hr",
			},
			// 关键信息部分
			map[string]interface{}{
				"tag": "div",
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**主机名**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": hostName,
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "lark_md",
							"content": "**项目**",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": project,
						},
					},
				},
			},
			map[string]interface{}{
				"tag": "note",
				"elements": []map[string]interface{}{
					{
						"tag":     "plain_text",
						"content": "此告警已恢复正常",
					},
				},
			},
		},
	}

	return card
}

// sendAlertNotification 发送告警通知
func sendAlertNotification(alert monitor.AlertDetail, fingerprint string, status string) {
	alertName := getAlertName(alert)
	middleware.Logger.Infof("发送告警 [%s] 的通知", alertName)

	// 构建飞书卡片
	card := buildFeiShuCard(alert)

	// 获取项目名并发送飞书通知
	projectName := ""
	if project, ok := alert.Labels["project"]; ok {
		projectName = project
	}

	var err error
	if projectName != "" {
		err = notice.SendFeiShuCardWithProject(card, projectName)
	} else {
		err = notice.SendFeiShuCard(card)
	}
	if err != nil {
		middleware.Logger.Errorf("发送飞书通知失败: %v", err)
		return
	}

	// 标记告警已发送
	markSent(fingerprint, status)
}

// sendRecoveryNotification 发送告警恢复通知
func sendRecoveryNotification(alert monitor.AlertDetail, fingerprint string) {
	alertName := getAlertName(alert)
	middleware.Logger.Infof("发送告警 [%s] 的恢复通知", alertName)

	// 构建恢复通知卡片
	card := buildFeiShuRecoveryCard(alert)

	// 获取项目名并发送飞书恢复通知
	projectName := ""
	if project, ok := alert.Labels["project"]; ok {
		projectName = project
	}

	var err error
	if projectName != "" {
		err = notice.SendFeiShuCardWithProject(card, projectName)
	} else {
		err = notice.SendFeiShuCard(card)
	}
	if err != nil {
		middleware.Logger.Errorf("发送告警恢复通知失败: %v", err)
		return
	}

	// 标记告警状态为已恢复
	markSent(fingerprint, "resolved")
}
