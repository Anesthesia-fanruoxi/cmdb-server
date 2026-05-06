package middleware

import (
	"bytes"
	"cmdb/config"
	"cmdb/pkg/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AgentResponse Agent 统一响应结构
type AgentResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// CallAgentGET 调用 Agent GET 接口
// pluginPath: 插件路径,如 "/proxy/databi-plugs/spark/common/tables"
// agentURL: Agent 地址
// secret: 加密密钥
func CallAgentGET(agentURL, pluginPath, secret string) ([]byte, error) {
	// 构建完整 URL
	fullURL := agentURL + pluginPath
	Logger.Infof("[AgentClient GET] 请求URL: %s", fullURL)

	// 创建 GET 请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		Logger.Errorf("[AgentClient GET] 请求失败，HTTP状态码: %d, URL: %s", resp.StatusCode, fullURL)
		return nil, fmt.Errorf("Agent接口返回 %d，请检查路径是否正确", resp.StatusCode)
	}

	//Logger.Infof("[AgentClient GET] 原始响应: %s", string(respBody))

	// 解析加密响应
	var encryptedResp struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
		Logger.Errorf("[AgentClient GET] 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	//Logger.Infof("[AgentClient GET] Data长度: %d", len(encryptedResp.Data))

	// 检查 data 是否为空
	if encryptedResp.Data == "" {
		Logger.Errorf("[AgentClient GET] Agent返回的data为空")
		return nil, fmt.Errorf("Agent返回的data为空")
	}

	// 解密数据
	decryptedData, err := DecryptAndDecompress(encryptedResp.Data, secret)
	if err != nil {
		Logger.Errorf("[AgentClient GET] 解密失败: %v", err)
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	//Logger.Infof("[AgentClient GET] 解密后数据: %s", string(decryptedData))

	return decryptedData, nil
}

// CallAgentPOST 调用 Agent POST 接口
// pluginPath: 插件路径,如 "/proxy/databi-plugs/spark/common/query"
// agentURL: Agent 地址
// requestBody: 请求体(会自动加密)
// secret: 加密密钥
func CallAgentPOST(agentURL, pluginPath string, requestBody interface{}, secret string) ([]byte, error) {
	// 序列化请求体
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	Logger.Infof("[AgentClient POST] 请求体: %s", string(requestJSON))

	// 加密请求体
	encryptedData, err := CompressAndEncrypt(requestJSON, secret)
	if err != nil {
		return nil, fmt.Errorf("加密请求体失败: %w", err)
	}

	// 构建加密请求
	encryptedRequest := map[string]interface{}{
		"data": encryptedData,
	}
	encryptedRequestJSON, err := json.Marshal(encryptedRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化加密请求失败: %w", err)
	}

	// 构建完整 URL
	fullURL := agentURL + pluginPath
	Logger.Infof("[AgentClient POST] 请求URL: %s", fullURL)

	// 创建 POST 请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(encryptedRequestJSON))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		Logger.Errorf("[AgentClient POST] 请求失败，HTTP状态码: %d, URL: %s", resp.StatusCode, fullURL)
		return nil, fmt.Errorf("Agent接口返回 %d，请检查路径是否正确", resp.StatusCode)
	}

	//Logger.Infof("[AgentClient POST] 原始响应: %s", string(respBody))

	// 解析加密响应
	var encryptedResp struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
		Logger.Errorf("[AgentClient POST] 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	//Logger.Infof("[AgentClient POST] Data长度: %d", len(encryptedResp.Data))

	// 检查 data 是否为空
	if encryptedResp.Data == "" {
		Logger.Errorf("[AgentClient POST] Agent返回的data为空")
		return nil, fmt.Errorf("Agent返回的data为空")
	}

	// 解密数据
	decryptedData, err := DecryptAndDecompress(encryptedResp.Data, secret)
	if err != nil {
		Logger.Errorf("[AgentClient POST] 解密失败: %v", err)
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	//Logger.Infof("[AgentClient POST] 解密后数据: %s", string(decryptedData))

	return decryptedData, nil
}

// CallAgent 统一的 Agent 调用方法（GET 请求）
// project: 项目标识，用于获取 Agent 地址
// path: 插件路径，如 "/proxy/databi-plugs/spark/common/tables"
// 返回: 解密后的原始数据
func CallAgent(project, path string) ([]byte, error) {
	// 1. 获取 Agent 地址
	agentURL, err := getProjectAgentURL(project)
	if err != nil {
		return nil, fmt.Errorf("获取Agent地址失败: %w", err)
	}

	// 2. 获取加密密钥
	secret, err := getAgentSecret()
	if err != nil {
		return nil, fmt.Errorf("获取加密密钥失败: %w", err)
	}

	// 3. 调用 Agent GET 接口
	return CallAgentGET(agentURL, path, secret)
}

// CallAgentWithBody 统一的 Agent 调用方法（POST 请求）
// project: 项目标识，用于获取 Agent 地址
// path: 插件路径，如 "/proxy/databi-plugs/spark/common/query"
// body: 请求体（任意类型）
// 返回: 解密后的原始数据
func CallAgentWithBody(project, path string, body interface{}) ([]byte, error) {
	// 1. 获取 Agent 地址
	agentURL, err := getProjectAgentURL(project)
	if err != nil {
		return nil, fmt.Errorf("获取Agent地址失败: %w", err)
	}

	// 2. 获取加密密钥
	secret, err := getAgentSecret()
	if err != nil {
		return nil, fmt.Errorf("获取加密密钥失败: %w", err)
	}

	// 3. 调用 Agent POST 接口
	return CallAgentPOST(agentURL, path, body, secret)
}

// getProjectAgentURL 获取项目的 Agent 地址（内部辅助方法）
func getProjectAgentURL(project string) (string, error) {
	var projectDict struct {
		AgentURL string `gorm:"column:agent_url"`
	}

	// 使用全局 DB 实例查询
	if err := database.DB.Table(TableProjectDict).
		Select("agent_url").
		Where("project = ?", project).
		First(&projectDict).Error; err != nil {
		return "", fmt.Errorf("查询项目 %s 失败: %w", project, err)
	}

	if projectDict.AgentURL == "" {
		return "", fmt.Errorf("项目 %s 的Agent地址未配置", project)
	}

	return projectDict.AgentURL, nil
}

// getAgentSecret 获取加密密钥（内部辅助方法）
func getAgentSecret() (string, error) {
	cfg := config.GetConfig()
	secret := cfg.Agent.Secret

	if secret == "" {
		return "", fmt.Errorf("Agent密钥未配置")
	}

	return secret, nil
}
