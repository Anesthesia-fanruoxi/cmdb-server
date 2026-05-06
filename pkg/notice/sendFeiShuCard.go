package notice

import (
	"bytes"
	"cmdb/config"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 邮箱到用户ID的映射表
var emailToUserIDMap = map[string]string{
	"zhanghuajia@hzbxhd.com": "1gcd7cd8",
	"zhaowenda@hzbxhd.com":   "7a8d37gb",
	"guokai@hzbxhd.com":      "72bf1bg3",
	"wangxin@hzbxhd.com":     "3bdade88",
	"huxiaojun@hzbxhd.com":   "6696f6fd",
	"liyu@hzbxhd.com":        "47b555ad",
	"liyang@hzbxhd.com":      "5b5eb6bg",
	"liurong@hzbxhd.com":     "f6d2773d",
	"jiangqi@hzbxhd.com":     "bbbe9c64",
	// 可以添加更多的映射

}

// 飞书卡片消息相关结构
type feiShuCardConfig struct {
	AppID     string
	AppSecret string
	BaseURL   string
}

type feiShuTokenResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type feiShuCardMessage struct {
	ReceiveID string `json:"receive_id"` // 接收者ID
	MsgType   string `json:"msg_type"`   // 固定为 "interactive"
	Content   string `json:"content"`    // 卡片内容JSON字符串
	UUID      string `json:"uuid"`       // 消息UUID
}

type feiShuMessageResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

// 缓存access token，避免频繁请求
var (
	cardTokenCache   string
	cardTokenExpired time.Time
)

// getFeiShuCardConfig 获取飞书卡片配置
func getFeiShuCardConfig() (*feiShuCardConfig, error) {
	cfg := config.GetConfig()

	if cfg.FeiShu.AppID == "" || cfg.FeiShu.AppSecret == "" {
		return nil, fmt.Errorf("飞书应用配置不完整，请检查app_id和app_secret")
	}

	baseURL := cfg.FeiShu.BaseURL
	if baseURL == "" {
		baseURL = "https://open.feishu.cn"
	}

	return &feiShuCardConfig{
		AppID:     cfg.FeiShu.AppID,
		AppSecret: cfg.FeiShu.AppSecret,
		BaseURL:   baseURL,
	}, nil
}

// getCardAccessToken 获取access token
func getCardAccessToken() (string, error) {
	// 检查缓存的token是否还有效（提前5分钟刷新）
	if cardTokenCache != "" && time.Now().Before(cardTokenExpired.Add(-5*time.Minute)) {
		return cardTokenCache, nil
	}

	feishuCfg, err := getFeiShuCardConfig()
	if err != nil {
		return "", err
	}

	// 构建请求体
	reqBody := map[string]string{
		"app_id":     feishuCfg.AppID,
		"app_secret": feishuCfg.AppSecret,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 发送请求获取access token
	url := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", feishuCfg.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("获取access token请求失败: %v", err)
	}
	defer resp.Body.Close()

	var tokenResp feiShuTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("解析access token响应失败: %v", err)
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("获取access token失败: %s", tokenResp.Msg)
	}

	// 缓存token
	cardTokenCache = tokenResp.TenantAccessToken
	cardTokenExpired = time.Now().Add(time.Duration(tokenResp.Expire) * time.Second)

	middleware.Logger.Infof("获取飞书access token成功，token: %s", cardTokenCache)
	middleware.Logger.Infof("获取飞书access token成功，过期时间: %v", cardTokenExpired)
	return cardTokenCache, nil
}

// SendExportApprovalCard 发送SQL导出审批卡片消息
func SendExportApprovalCard(userEmail string, exportInfo ExportCardInfo) error {
	// 根据邮箱查找用户ID
	userID, err := getUserIDByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("根据邮箱查找用户失败: %v", err)
	}

	// 创建审批卡片内容
	card := createExportApprovalCard(exportInfo)

	// 发送卡片消息
	return sendCardMessage(userID, card)
}

// ExportCardInfo 导出卡片信息
type ExportCardInfo struct {
	ExportID        int64  `json:"export_id"`        // 导出ID
	SubmitterName   string `json:"submitter_name"`   // 提交人姓名
	ProjectName     string `json:"project_name"`     // 项目名称
	SubmitterRemark string `json:"submitter_remark"` // 申请说明
	ApprovalURL     string `json:"approval_url"`     // 审批链接
	Status          string `json:"status"`           // 状态
	CreatedAt       string `json:"created_at"`       // 创建时间
}

// createExportApprovalCard 创建导出审批卡片
func createExportApprovalCard(info ExportCardInfo) map[string]interface{} {
	// 根据状态设置不同的颜色和图标
	var headerColor, statusText, statusIcon string
	var showButtons bool

	switch info.Status {
	case "submitted":
		headerColor = "blue"
		statusText = "待审批"
		statusIcon = "📋"
		showButtons = true
	case "approved":
		headerColor = "green"
		statusText = "已审批"
		statusIcon = "✅"
		showButtons = false
	case "rejected":
		headerColor = "red"
		statusText = "已拒绝"
		statusIcon = "❌"
		showButtons = false
	case "completed":
		headerColor = "green"
		statusText = "已完成"
		statusIcon = "🎉"
		showButtons = false
	default:
		headerColor = "grey"
		statusText = "处理中"
		statusIcon = "⏳"
		showButtons = false
	}

	// 构建卡片元素
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"fields": []map[string]interface{}{
				{
					"is_short": true,
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**申请ID：** %d", info.ExportID),
						"tag":     "lark_md",
					},
				},
				{
					"is_short": true,
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**提交人：** %s", info.SubmitterName),
						"tag":     "lark_md",
					},
				},
				{
					"is_short": true,
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**项目：** %s", info.ProjectName),
						"tag":     "lark_md",
					},
				},
				{
					"is_short": true,
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**提交时间：** %s", info.CreatedAt),
						"tag":     "lark_md",
					},
				},
			},
		},
	}

	// 添加申请说明
	if info.SubmitterRemark != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**申请说明：**\n%s", info.SubmitterRemark),
				"tag":     "lark_md",
			},
		})
	}

	// 添加分割线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 如果需要显示按钮（待审批状态）
	if showButtons && info.ApprovalURL != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": "立即审批",
						"tag":     "plain_text",
					},
					"url":  info.ApprovalURL,
					"type": "primary",
				},
			},
		})
	} else if info.ApprovalURL != "" {
		// 其他状态显示查看详情按钮
		elements = append(elements, map[string]interface{}{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": "查看详情",
						"tag":     "plain_text",
					},
					"url":  info.ApprovalURL,
					"type": "default",
				},
			},
		})
	}

	// 构建完整卡片
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
			"enable_forward":   true,
		},
		"header": map[string]interface{}{
			"template": headerColor,
			"title": map[string]interface{}{
				"content": fmt.Sprintf("%s SQL导出申请 - %s", statusIcon, statusText),
				"tag":     "plain_text",
			},
		},
		"elements": elements,
	}

	return card
}

// sendCardMessage 发送卡片消息
func sendCardMessage(receiverID string, card map[string]interface{}) error {
	// 获取access token
	token, err := getCardAccessToken()
	if err != nil {
		return fmt.Errorf("获取access token失败: %v", err)
	}

	feishuCfg, err := getFeiShuCardConfig()
	if err != nil {
		return err
	}

	// 序列化卡片内容
	cardJSON, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("序列化卡片内容失败: %v", err)
	}

	// 构建消息体
	message := feiShuCardMessage{
		ReceiveID: receiverID,
		MsgType:   "interactive",
		Content:   string(cardJSON),
		UUID:      fmt.Sprintf("export_card_%d", time.Now().UnixNano()),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发送消息
	url := fmt.Sprintf("%s/open-apis/im/v1/messages", feishuCfg.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// 设置接收者类型为user_id
	query := req.URL.Query()
	query.Add("receive_id_type", "user_id")
	req.URL.RawQuery = query.Encode()

	middleware.Logger.Infof("使用user_id作为接收者类型: %s", receiverID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送消息请求失败: %v", err)
	}
	defer resp.Body.Close()

	var msgResp feiShuMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return fmt.Errorf("解析消息响应失败: %v", err)
	}

	if msgResp.Code != 0 {
		return fmt.Errorf("发送消息失败: %s", msgResp.Msg)
	}

	middleware.Logger.Infof("成功发送飞书卡片消息给用户 %s，消息ID: %s", receiverID, msgResp.Data.MessageID)
	return nil
}

// getUserIDByEmail 根据邮箱查找用户ID（使用映射表）
func getUserIDByEmail(email string) (string, error) {
	// 优先从映射表中查找
	if userID, ok := emailToUserIDMap[email]; ok {
		middleware.Logger.Infof("从映射表中获取用户ID: %s (邮箱: %s)", userID, email)
		return userID, nil
	}

	// 如果映射表中没有，记录警告并返回错误
	middleware.Logger.Warnf("邮箱 %s 未在映射表中找到对应的用户ID", email)
	return "", fmt.Errorf("邮箱 %s 未在映射表中找到对应的用户ID", email)
}

// ApprovalCardInfo 审批卡片信息
type ApprovalCardInfo struct {
	ExportID        int64  `json:"export_id"`        // 导出ID
	SubmitterName   string `json:"submitter_name"`   // 提交人姓名
	ProjectName     string `json:"project_name"`     // 项目名称
	DatabaseName    string `json:"database_name"`    // 数据库名称
	SubmitterRemark string `json:"submitter_remark"` // 申请说明
	SQLContent      string `json:"sql_content"`      // SQL内容
	CreatedAt       string `json:"created_at"`       // 创建时间
	Role            string `json:"role"`             // 当前用户角色: approver, reviewer, executor
	HasSQL          bool   `json:"has_sql"`          // 是否已有SQL
}

// SendApprovalActionCard 发送审批操作卡片
func SendApprovalActionCard(userEmail string, cardInfo ApprovalCardInfo) error {
	// 根据邮箱查找用户ID
	userID, err := getUserIDByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("根据邮箱查找用户失败: %v", err)
	}

	// 创建审批操作卡片内容
	card := createApprovalActionCard(cardInfo)

	// 发送卡片消息
	return sendCardMessage(userID, card)
}

// createApprovalActionCard 创建审批操作卡片
func createApprovalActionCard(info ApprovalCardInfo) map[string]interface{} {
	var headerColor, title, statusIcon string
	// 按钮功能已禁用，不再声明buttons变量

	// 根据角色设置标题和按钮
	switch info.Role {
	case "approver":
		headerColor = "blue"
		title = "SQL导出审批"
		statusIcon = "📋"
		// 按钮功能已禁用
	case "reviewer":
		headerColor = "orange"
		title = "SQL导出审核"
		statusIcon = "🔍"
		// 审核按钮已禁用
	case "executor":
		headerColor = "green"
		title = "SQL导出执行"
		statusIcon = "⚡"
		// 执行按钮已禁用
	default:
		headerColor = "grey"
		title = "SQL导出通知"
		statusIcon = "📄"
		// 默认情况下不显示按钮
	}

	// 构建卡片元素
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**提交人：** %s", info.SubmitterName),
				"tag":     "lark_md",
			},
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**项目名称：** %s", info.ProjectName),
				"tag":     "lark_md",
			},
		},
	}

	if info.DatabaseName != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**数据库：** %s", info.DatabaseName),
				"tag":     "lark_md",
			},
		})
	}

	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("**申请说明：** %s", info.SubmitterRemark),
			"tag":     "lark_md",
		},
	})

	if info.SQLContent != "" {
		// 显示完整的SQL内容，不进行截取
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**SQL内容：**\n%s", info.SQLContent),
				"tag":     "lark_md",
			},
		})
	}

	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("**创建时间：** %s", info.CreatedAt),
			"tag":     "lark_md",
		},
	})

	// 添加分隔线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 不添加操作按钮
	// 按钮已被禁用，不再显示

	// 构建完整卡片
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
			"enable_forward":   true,
		},
		"header": map[string]interface{}{
			"template": headerColor,
			"title": map[string]interface{}{
				"content": fmt.Sprintf("%s %s", statusIcon, title),
				"tag":     "plain_text",
			},
		},
		"elements": elements,
	}

	return card
}

// SQLApplyCardInfo SQL执行审批卡片信息
type SQLApplyCardInfo struct {
	ApplyID       int64  `json:"apply_id"`       // 审批ID
	SubmitterName string `json:"submitter_name"` // 提交人姓名
	ProjectName   string `json:"project_name"`   // 项目名称
	DatabaseName  string `json:"database_name"`  // 数据库名称
	Description   string `json:"description"`    // 申请说明
	ExecutionTime string `json:"execution_time"` // 执行时间，为空则立即执行
	CreatedAt     string `json:"created_at"`     // 创建时间
	Role          string `json:"role"`           // 当前用户角色: approver, executor
}

// SendSQLApplyCard 发送SQL执行审批卡片
func SendSQLApplyCard(userEmail string, cardInfo SQLApplyCardInfo) error {
	// 根据邮箱查找用户ID
	userID, err := getUserIDByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("根据邮箱查找用户失败: %v", err)
	}

	// 创建卡片内容
	card := createSQLApplyCard(cardInfo)

	// 发送卡片消息
	return sendCardMessage(userID, card)
}

// createSQLApplyCard 创建SQL执行审批卡片
func createSQLApplyCard(info SQLApplyCardInfo) map[string]interface{} {
	var headerColor, title, statusIcon string

	// 根据角色设置标题和颜色
	switch info.Role {
	case "approver":
		headerColor = "blue"
		title = "SQL执行审批"
		statusIcon = "📋"
	case "executor":
		headerColor = "green"
		title = "SQL执行通知"
		statusIcon = "⚡"
	default:
		headerColor = "grey"
		title = "SQL执行通知"
		statusIcon = "📄"
	}

	// 构建卡片元素
	elements := []map[string]interface{}{
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**提交人：** %s", info.SubmitterName),
				"tag":     "lark_md",
			},
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**项目名称：** %s", info.ProjectName),
				"tag":     "lark_md",
			},
		},
		{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**数据库：** %s", info.DatabaseName),
				"tag":     "lark_md",
			},
		},
	}

	// 添加申请说明
	if info.Description != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"content": fmt.Sprintf("**申请说明：** %s", info.Description),
				"tag":     "lark_md",
			},
		})
	}

	// 添加执行时间
	executionTimeText := "立即执行"
	if info.ExecutionTime != "" {
		executionTimeText = info.ExecutionTime
	}
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("**执行时间：** %s", executionTimeText),
			"tag":     "lark_md",
		},
	})

	// 添加创建时间
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("**创建时间：** %s", info.CreatedAt),
			"tag":     "lark_md",
		},
	})

	// 添加分隔线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 根据角色添加操作按钮
	if info.Role == "approver" {
		elements = append(elements, map[string]interface{}{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": "✅ 通过",
						"tag":     "plain_text",
					},
					"type": "primary",
					"value": map[string]string{
						"apply_id": fmt.Sprintf("%d", info.ApplyID),
						"action":   "approve",
					},
				},
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": "❌ 拒绝",
						"tag":     "plain_text",
					},
					"type": "danger",
					"value": map[string]string{
						"apply_id": fmt.Sprintf("%d", info.ApplyID),
						"action":   "reject",
					},
				},
			},
			"layout": "bisected",
		})
	} else if info.Role == "executor" {
		elements = append(elements, map[string]interface{}{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": "⚡ 立即执行",
						"tag":     "plain_text",
					},
					"type": "primary",
					"value": map[string]string{
						"apply_id": fmt.Sprintf("%d", info.ApplyID),
						"action":   "execute",
					},
				},
			},
		})
	}

	// 构建完整卡片
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
			"enable_forward":   true,
		},
		"header": map[string]interface{}{
			"template": headerColor,
			"title": map[string]interface{}{
				"content": fmt.Sprintf("%s %s", statusIcon, title),
				"tag":     "plain_text",
			},
		},
		"elements": elements,
	}

	return card
}
