package proUpdate

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProUpdateCreateRequest 创建任务请求结构
type ProUpdateCreateRequest struct {
	Project     string `json:"project" validate:"required"` // 项目名称
	Type        string `json:"type"`                        //
	Category    string `json:"category"`
	Description string `json:"description"`
}

// ProUpdateCreateResponse 创建任务响应结构
type ProUpdateCreateResponse struct {
	ID      string `json:"id"`      // 任务ID
	Project string `json:"project"` // 项目名称
	Type    string `json:"type"`    // 任务类型
}

// ProUpdateCreate 创建项目发版任务
// POST /api/assets/proUpdate/create
func ProUpdateCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求参数
	var req ProUpdateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Logger.Errorf("解析创建任务请求失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	// 验证参数
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目名称不能为空")
		return
	}

	middleware.Logger.Infof("收到创建任务请求 - Project: %s, Type: %s, Category: %s, Description: %s", req.Project, req.Type, req.Category, req.Description)

	// 调用代理服务创建任务（直接传递原始参数，由agent处理）
	taskResp, err := createTaskOnProxy(req)
	if err != nil {
		middleware.Logger.Errorf("调用代理服务创建任务失败 - Project: %s, Error: %v", req.Project, err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("创建任务失败: %v", err))
		return
	}

	middleware.Logger.Infof("任务创建成功 - Project: %s, TaskID: %s", req.Project, taskResp.ID)
	middleware.ResponseSuccess(w, "任务创建成功", taskResp)
}

// createTaskOnProxy 调用代理服务创建任务
func createTaskOnProxy(req ProUpdateCreateRequest) (*ProUpdateCreateResponse, error) {
	// 从数据库获取项目的 agent_url
	agentURL, err := system.GetProjectAgentURL(req.Project)
	if err != nil {
		return nil, fmt.Errorf("获取项目 %s 的Agent地址失败: %v", req.Project, err)
	}

	middleware.Logger.Infof("项目 %s 使用Agent地址: %s", req.Project, agentURL)

	createURL := fmt.Sprintf("%s/proxy/cicd-agent/update", agentURL)

	// 构建传递给代理服务的请求体
	requestBody := map[string]string{
		"project": req.Project,
	}
	if req.Type != "" {
		requestBody["type"] = req.Type
	}
	if req.Category != "" {
		requestBody["category"] = req.Category
	}
	if req.Description != "" {
		requestBody["description"] = req.Description
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	middleware.Logger.Infof("调用代理服务创建任务 - URL: %s, Project: %s, Type: %s", createURL, req.Project, req.Type)

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送HTTP POST请求
	resp, err := client.Post(createURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("发送创建请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 先解析外层响应（可能包含加密数据）
	var encryptedResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"` // agent模式返回加密字符串
	}

	if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
		// 如果解析失败且HTTP状态码不是200，直接返回HTTP错误
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("代理服务返回错误状态: %d, 响应: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("解析响应数据失败: %v", err)
	}

	// 解密响应数据（使用Agent加密密钥）
	cfg := config.GetConfig()
	salt := cfg.Agent.Secret
	if salt == "" {
		return nil, fmt.Errorf("Agent加密密钥未配置")
	}

	decryptedData, err := middleware.DecryptAndDecompress(encryptedResp.Data, salt)
	if err != nil {
		return nil, fmt.Errorf("解密响应数据失败: %v (HTTP状态: %d)", err, resp.StatusCode)
	}

	middleware.Logger.Infof("使用Agent密钥解密成功 - 解密后的响应数据: %s", string(decryptedData))

	// 解析解密后的实际数据
	var finalResp struct {
		Code    int                     `json:"code"`
		Message string                  `json:"message"`
		Data    ProUpdateCreateResponse `json:"data"`
	}

	if err := json.Unmarshal(decryptedData, &finalResp); err != nil {
		return nil, fmt.Errorf("解析解密后的数据失败: %v", err)
	}

	if finalResp.Code != 200 {
		return nil, fmt.Errorf("代理服务返回错误: %s", finalResp.Message)
	}

	middleware.Logger.Infof("代理服务创建任务成功 - ProjectName: %s, TaskID: %s", req.Project, finalResp.Data.ID)
	return &finalResp.Data, nil
}

// extractProjectIdentifier 提取项目标识
// 规则：如果项目名有 -web 后缀，去掉后缀；否则返回完整项目名
// 示例：jxh-web → jxh, jxh → jxh, jxh-api → jxh-api
func extractProjectIdentifier(projectName string) string {
	if strings.HasSuffix(projectName, "-web") {
		return strings.TrimSuffix(projectName, "-web")
	}
	return projectName
}
