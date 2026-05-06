package project

import (
	"bytes"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AgentOperateResponse Agent操作接口响应
type AgentOperateResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// callAgent 统一调用 Agent HTTP 接口
func callAgent(method, agentAPIURL string, payload interface{}) (*AgentOperateResponse, error) {
	var httpReq *http.Request
	var err error

	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("构建请求失败: %w", err)
		}
		httpReq, err = http.NewRequest(method, agentAPIURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
	} else {
		httpReq, err = http.NewRequest(method, agentAPIURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Agent响应失败: %w", err)
	}

	var agentResp AgentOperateResponse
	if err := json.Unmarshal(body, &agentResp); err != nil {
		return nil, fmt.Errorf("解析Agent响应失败: %w", err)
	}

	return &agentResp, nil
}

// getAgentURL 获取项目 Agent 地址
func getAgentURL(project string) (string, error) {
	agentURL, err := system.GetProjectAgentURL(project)
	if err != nil {
		return "", fmt.Errorf("获取项目Agent地址失败: %w", err)
	}
	return agentURL, nil
}

// filterSensitiveConfig 过滤敏感配置信息
func filterSensitiveConfig(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return data
	}
	safeCopy := make(map[string]interface{})
	for k, v := range data {
		safeCopy[k] = v
	}
	delete(safeCopy, "config")
	return safeCopy
}

// sendAgentResult 统一发送 Agent 操作结果
func sendAgentResult(w http.ResponseWriter, project, pluginName, action string, agentResp *AgentOperateResponse) {
	if agentResp.Code != 200 {
		msg := agentResp.Message
		if msg == "" {
			msg = "操作失败"
		}
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("Agent操作失败: %s", msg))
		return
	}

	middleware.ResponseSuccess(w, fmt.Sprintf("%s 成功", action), map[string]interface{}{
		"project":     project,
		"plugin_name": pluginName,
		"action":      action,
		"agent_data":  filterSensitiveConfig(agentResp.Data),
	})
}
