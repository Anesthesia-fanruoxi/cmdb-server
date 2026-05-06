package notice

import (
	"bytes"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FeiShuMessage 飞书消息结构
type FeiShuMessage struct {
	MsgType string                 `json:"msg_type"`       // 消息类型：text/post/interactive/share_chat/image/audio/media/file
	Content map[string]interface{} `json:"content"`        // 消息内容
	Card    map[string]interface{} `json:"card,omitempty"` // 消息卡片，当msg_type为interactive时使用
}

// FeiShuResponse 飞书响应结构
type FeiShuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

var (
	webhookURL string        // 缓存的webhook地址
	lastCheck  time.Time     // 上次检查webhook的时间
	checkTTL   = time.Minute // webhook地址缓存时间
)

// getWebhookURL 获取飞书webhook地址
func getWebhookURL(projectName ...string) (string, error) {
	// 如果传入了项目名，优先从项目字典中获取
	if len(projectName) > 0 && projectName[0] != "" {
		var projectDict system.ProjectDict
		// 先尝试按 project_name（中文名称）查询
		if err := database.DB.Where("project_name = ?", projectName[0]).First(&projectDict).Error; err == nil {
			if projectDict.AlterFeishu != "" {
				middleware.Logger.Infof("使用项目 %s 的专用飞书地址: %s\n", projectName[0], projectDict.AlterFeishu)
				return projectDict.AlterFeishu, nil
			}
		} else {
			// 如果按中文名称查不到，尝试按 project（英文简称）查询
			if err := database.DB.Where("project = ?", projectName[0]).First(&projectDict).Error; err == nil {
				if projectDict.AlterFeishu != "" {
					middleware.Logger.Infof("使用项目 %s 的专用飞书地址: %s\n", projectName[0], projectDict.AlterFeishu)
					return projectDict.AlterFeishu, nil
				}
			}
		}
		middleware.Logger.Infof("项目 %s 未配置专用飞书地址，使用系统默认地址\n", projectName[0])
	}

	// 如果缓存的webhook地址存在且未过期，直接返回
	if webhookURL != "" && time.Since(lastCheck) < checkTTL {
		return webhookURL, nil
	}

	// 从数据库获取系统设置
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		middleware.Logger.Infof("获取系统设置失败: %v\n", err)
		return "", fmt.Errorf("获取系统设置失败: %v", err)
	}

	if setting.FeiShuWebhookUrl == "" {
		return "", fmt.Errorf("飞书webhook地址未配置")
	}

	// 更新缓存
	webhookURL = setting.FeiShuWebhookUrl
	lastCheck = time.Now()

	return webhookURL, nil
}

// SendFeiShuText 发送文本消息
func SendFeiShuText(content string) error {
	msg := FeiShuMessage{
		MsgType: "text",
		Content: map[string]interface{}{
			"text": content,
		},
	}
	return SendFeiShuMessage(msg)
}

// SendFeiShuCard 发送卡片消息
func SendFeiShuCard(card map[string]interface{}) error {
	msg := FeiShuMessage{
		MsgType: "interactive",
		Card:    card,
	}
	return SendFeiShuMessage(msg)
}

// SendFeiShuCardWithProject 发送卡片消息（支持项目参数）
func SendFeiShuCardWithProject(card map[string]interface{}, projectName string) error {
	msg := FeiShuMessage{
		MsgType: "interactive",
		Card:    card,
	}
	return SendFeiShuMessageWithProject(msg, projectName)
}

// SendFeiShuMessage 发送飞书消息
func SendFeiShuMessage(msg FeiShuMessage) error {
	// 获取webhook地址
	webhook, err := getWebhookURL()
	if err != nil {
		middleware.Logger.Errorf("获取webhook地址失败: %v\n", err)
		return err
	}
	return sendMessage(msg, webhook)
}

// SendFeiShuMessageWithProject 发送飞书消息（支持项目参数）
func SendFeiShuMessageWithProject(msg FeiShuMessage, projectName string) error {
	// 获取webhook地址（优先使用项目配置）
	webhook, err := getWebhookURL(projectName)
	if err != nil {
		middleware.Logger.Errorf("获取webhook地址失败: %v\n", err)
		return err
	}
	return sendMessage(msg, webhook)
}

// sendMessage 发送消息的通用方法
func sendMessage(msg FeiShuMessage, webhook string) error {
	// 转换消息为JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		middleware.Logger.Errorf("消息序列化失败: %v\n", err)
		return fmt.Errorf("消息序列化失败: %v", err)
	}

	// 发送请求
	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		middleware.Logger.Errorf("发送请求失败: %v\n", err)
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result FeiShuResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v\n", err)
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if result.Code != 0 {
		middleware.Logger.Errorf("飞书API返回错误: %s\n", result.Msg)
		return fmt.Errorf("发送失败: %s", result.Msg)
	}
	return nil
}
