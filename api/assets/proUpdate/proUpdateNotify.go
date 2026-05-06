package proUpdate

import (
	"bytes"
	"cmdb/models/assets"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// joinLines 用换行符连接字符串
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// StepTemplate 步骤模板
type StepTemplate struct {
	Step     int    `json:"step"`
	StepType string `json:"step_type"`
	StepName string `json:"step_name"`
	Remote   string `json:"remote"` // server 或 agent
}

// 双副本16步模板
var DualReplicaSteps = []StepTemplate{
	{Step: 1, StepType: "init", StepName: "创建工作目录", Remote: "server"},
	{Step: 2, StepType: "clear", StepName: "清理工作目录和旧镜像", Remote: "server"},
	{Step: 3, StepType: "git", StepName: "Git代码克隆", Remote: "server"},
	{Step: 4, StepType: "mvn", StepName: "编译代码", Remote: "server"},
	{Step: 5, StepType: "product", StepName: "获取项目产物", Remote: "server"},
	{Step: 6, StepType: "dockerfile", StepName: "创建Dockerfile", Remote: "server"},
	{Step: 7, StepType: "build", StepName: "构建Docker镜像", Remote: "server"},
	{Step: 8, StepType: "push", StepName: "推送Docker镜像", Remote: "server"},
	{Step: 9, StepType: "pullOnline", StepName: "拉取在线镜像", Remote: "agent"},
	{Step: 10, StepType: "tagImages", StepName: "标记镜像", Remote: "agent"},
	{Step: 11, StepType: "pushLocal", StepName: "推送本地镜像", Remote: "agent"},
	{Step: 12, StepType: "checkImage", StepName: "检查镜像", Remote: "agent"},
	{Step: 13, StepType: "deployService", StepName: "应用服务部署", Remote: "agent"},
	{Step: 14, StepType: "checkService", StepName: "检查服务就绪", Remote: "agent"},
	{Step: 15, StepType: "trafficSwitching", StepName: "流量切换", Remote: "agent"},
	{Step: 16, StepType: "cleanupOldVersion", StepName: "清理旧版本", Remote: "agent"},
}

// 单副本13步模板
var SingleReplicaSteps = []StepTemplate{
	{Step: 1, StepType: "init", StepName: "创建工作目录", Remote: "server"},
	{Step: 2, StepType: "clear", StepName: "清理工作目录和旧镜像", Remote: "server"},
	{Step: 3, StepType: "git", StepName: "Git代码克隆", Remote: "server"},
	{Step: 4, StepType: "mvn", StepName: "编译代码", Remote: "server"},
	{Step: 5, StepType: "product", StepName: "获取项目产物", Remote: "server"},
	{Step: 6, StepType: "dockerfile", StepName: "创建Dockerfile", Remote: "server"},
	{Step: 7, StepType: "build", StepName: "构建Docker镜像", Remote: "server"},
	{Step: 8, StepType: "push", StepName: "推送Docker镜像", Remote: "server"},
	{Step: 9, StepType: "pullOnline", StepName: "拉取在线镜像", Remote: "agent"},
	{Step: 10, StepType: "tagImages", StepName: "标记镜像", Remote: "agent"},
	{Step: 11, StepType: "pushLocal", StepName: "推送本地镜像", Remote: "agent"},
	{Step: 12, StepType: "checkImage", StepName: "检查镜像", Remote: "agent"},
	{Step: 13, StepType: "deployService", StepName: "应用服务部署", Remote: "agent"},
}

// Web10步模板
var WebSteps = []StepTemplate{
	{Step: 1, StepType: "init", StepName: "创建工作目录", Remote: "server"},
	{Step: 2, StepType: "clear", StepName: "清理工作目录", Remote: "server"},
	{Step: 3, StepType: "git", StepName: "Git代码克隆", Remote: "server"},
	{Step: 4, StepType: "npm", StepName: "安装Node.js依赖", Remote: "server"},
	{Step: 5, StepType: "build", StepName: "Web项目打包", Remote: "server"},
	{Step: 6, StepType: "product", StepName: "获取Web项目产物", Remote: "server"},
	{Step: 7, StepType: "downProduct", StepName: "下载产物", Remote: "agent"},
	{Step: 8, StepType: "extractProduct", StepName: "解压产物", Remote: "agent"},
	{Step: 9, StepType: "backupCurrent", StepName: "备份当前版本", Remote: "agent"},
	{Step: 10, StepType: "deployNew", StepName: "部署新版本", Remote: "agent"},
}

// GetStepsByType 根据任务类型获取步骤模板
// type: web=前端10步, single=后端单副本13步, double/空=后端双副本16步
func GetStepsByType(taskType string) []StepTemplate {
	switch taskType {
	case "web":
		return WebSteps
	case "single":
		return SingleReplicaSteps
	case "double", "":
		return DualReplicaSteps
	default:
		return DualReplicaSteps // 默认双副本
	}
}

// GetTotalStepsByType 根据任务类型获取总步骤数
// type: web=10步, single=13步, double/空=16步
func GetTotalStepsByType(taskType string) int {
	switch taskType {
	case "web":
		return 10
	case "single":
		return 13
	case "double", "":
		return 16
	default:
		return 16 // 默认双副本
	}
}

// getSystemURL 从数据库获取平台地址
func getSystemURL() string {
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		middleware.Logger.Errorf("获取系统设置失败: %v", err)
		return ""
	}
	return setting.SystemUrl
}

// SendTaskStartNotify 发送任务开始通知（只发一次，包含手机端链接）
// taskID: 任务ID
// taskRecord: Redis中的任务记录
// webhookURL: 飞书webhook地址
func SendTaskStartNotify(taskID string, taskRecord *assets.RedisTaskRecord, webhookURL string) error {
	if webhookURL == "" {
		middleware.Logger.Debug("未配置飞书webhook地址，跳过通知")
		return nil
	}

	// 从数据库获取平台地址
	baseURL := getSystemURL()
	if baseURL == "" {
		middleware.Logger.Error("未配置平台地址，跳过通知")
		return nil
	}

	// 构建手机端查看链接: http://localhost:3000/release?id=xxx
	mobileURL := fmt.Sprintf("%s/release?id=%s", baseURL, taskID)

	// 构建飞书卡片消息
	card := buildTaskStartCard(taskRecord, mobileURL)

	return sendFeishuCard(card, webhookURL)
}

// buildTaskStartCard 构建任务开始通知卡片
func buildTaskStartCard(taskRecord *assets.RedisTaskRecord, mobileURL string) map[string]interface{} {
	// 获取总步骤数
	totalSteps := GetTotalStepsByType(taskRecord.Type)

	// 构建内容
	var contentLines []string
	contentLines = append(contentLines, fmt.Sprintf("**项目名称**：%s", taskRecord.ProjectName))
	contentLines = append(contentLines, fmt.Sprintf("**发版类型**：%s", getTypeDisplayName(taskRecord.Type)))
	// 如果有发版说明，显示出来
	if taskRecord.Description != "" {
		contentLines = append(contentLines, fmt.Sprintf("**发版说明**：%s", taskRecord.Description))
	}
	contentLines = append(contentLines, fmt.Sprintf("**总步骤数**：%d 步", totalSteps))
	contentLines = append(contentLines, fmt.Sprintf("**开始时间**：%s", taskRecord.StartedAt))
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, "📱 **点击下方按钮查看实时进度**")

	// 构建卡片
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": fmt.Sprintf("【%s】生产发版开始", taskRecord.ProjectName),
			},
			"template": "blue",
		},
		"elements": []map[string]interface{}{
			{
				"tag":  "div",
				"text": map[string]interface{}{"tag": "lark_md", "content": joinLines(contentLines)},
			},
			{
				"tag": "action",
				"actions": []map[string]interface{}{
					{
						"tag": "button",
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": "👁️ 查看发版进度",
						},
						"type": "primary",
						"url":  mobileURL,
					},
				},
			},
		},
	}

	return card
}

// getTypeDisplayName 获取类型显示名称
func getTypeDisplayName(taskType string) string {
	switch taskType {
	case "web":
		return "前端项目"
	case "single":
		return "后端单副本"
	case "double", "":
		return "后端双副本"
	default:
		return "后端双副本"
	}
}

// formatDuration 格式化耗时
func formatDuration(seconds float64) string {
	if seconds < 1 {
		return fmt.Sprintf("%.0f毫秒", seconds*1000)
	} else if seconds < 60 {
		return fmt.Sprintf("%.0f秒", seconds)
	} else {
		minutes := int(seconds) / 60
		secs := int(seconds) % 60
		return fmt.Sprintf("%d分%d秒", minutes, secs)
	}
}

// sendFeishuCard 发送飞书卡片消息
func sendFeishuCard(card map[string]interface{}, webhookURL string) error {
	msg := map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		middleware.Logger.Errorf("飞书消息序列化失败: %v", err)
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		middleware.Logger.Errorf("发送飞书通知失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		middleware.Logger.Errorf("解析飞书响应失败: %v", err)
		return err
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		middleware.Logger.Errorf("飞书API返回错误: %v", result["msg"])
		return fmt.Errorf("飞书发送失败: %v", result["msg"])
	}

	middleware.Logger.Infof("飞书步骤通知发送成功")
	return nil
}
