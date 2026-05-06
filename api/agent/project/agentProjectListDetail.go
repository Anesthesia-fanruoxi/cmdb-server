package project

import (
	"cmdb/models/agent"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AgentPluginInfo Agent返回的插件信息（对应agent API的响应）
type AgentPluginInfo struct {
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Category      string                 `json:"category"`       // container/binary
	Status        string                 `json:"status"`         // running/stopped/error/not_found/paused
	Port          int                    `json:"port"`           // 宿主机端口
	ContainerPort int                    `json:"container_port"` // 容器内端口
	Uptime        string                 `json:"uptime"`         // 运行时间
	Config        map[string]interface{} `json:"config"`         // 配置参数
	InstalledAt   string                 `json:"installed_at"`   // 安装时间
}

// AgentListResponse Agent列表接口响应
type AgentListResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// AgentProjectListDetail 查询项目已安装的插件列表
func AgentProjectListDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取项目参数
	project := r.URL.Query().Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 获取项目Agent URL
	agentURL, err := system.GetProjectAgentURL(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目Agent地址失败: %v", err))
		return
	}

	// 获取查询参数（传递给Agent）
	query := r.URL.Query()
	category := query.Get("category") // 插件类型过滤
	status := query.Get("status")     // 状态过滤
	name := query.Get("name")         // 名称搜索

	// 构建Agent API URL
	agentAPIURL := agentURL + "/api/plugins/list"

	// 添加查询参数
	req, err := http.NewRequest("GET", agentAPIURL, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "构建请求失败")
		return
	}

	q := req.URL.Query()
	if category != "" {
		q.Add("category", category)
	}
	if status != "" {
		q.Add("status", status)
	}
	if name != "" {
		q.Add("name", name)
	}
	req.URL.RawQuery = q.Encode()

	// 调用Agent列表接口
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("调用Agent失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 读取Agent响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取Agent响应失败")
		return
	}

	// 解析Agent响应
	var agentResp AgentListResponse
	if err := json.Unmarshal(body, &agentResp); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析Agent响应失败")
		return
	}

	// 检查Agent响应状态
	if agentResp.Code != 200 {
		msg := agentResp.Message
		if msg == "" {
			msg = "查询失败"
		}
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("Agent查询失败: %s", msg))
		return
	}

	// 增强插件数据：对比最新版本
	enhancedData := enhancePluginsWithVersionCheck(agentResp.Data)

	// 返回增强后的数据
	middleware.ResponseSuccess(w, "查询成功", enhancedData)
}

// enhancePluginsWithVersionCheck 增强插件数据，添加版本对比信息
func enhancePluginsWithVersionCheck(data map[string]interface{}) map[string]interface{} {
	// 获取plugins列表
	plugins, ok := data["plugins"].([]interface{})
	if !ok {
		return data
	}

	// 查询所有插件的最新版本（一次性查询，减少数据库访问）
	var storePlugins []agent.AgentStore
	database.DB.Select("name, version").
		Group("name").
		Order("name ASC, version DESC").
		Find(&storePlugins)

	// 构建插件名称到最新版本的映射
	latestVersionMap := make(map[string]string)
	for _, sp := range storePlugins {
		if _, exists := latestVersionMap[sp.Name]; !exists {
			latestVersionMap[sp.Name] = sp.Version
		}
	}

	// 从Agent响应中获取Agent版本信息
	agentVersion := ""
	if agentVersionData, exists := data["agent_version"]; exists {
		if version, ok := agentVersionData.(string); ok {
			agentVersion = version
		}
	}

	// 遍历已安装的插件，添加is_update字段，对配置进行脱敏处理
	for i, pluginData := range plugins {
		plugin, ok := pluginData.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取插件名称和当前版本
		name, _ := plugin["name"].(string)
		currentVersion, _ := plugin["version"].(string)

		// 对比版本
		if latestVersion, exists := latestVersionMap[name]; exists {
			// 如果数据库有最新版本记录
			plugin["latest_version"] = latestVersion
			plugin["is_update"] = currentVersion != latestVersion
		} else {
			// 数据库没有这个插件（可能是手动安装的）
			plugin["latest_version"] = ""
			plugin["is_update"] = false
		}

		// 对配置进行脱敏处理
		if config, ok := plugin["config"].(map[string]interface{}); ok {
			plugin["config"] = maskSensitiveConfig(config)
		}

		plugins[i] = plugin
	}

	// 更新Agent版本信息
	data["plugins"] = plugins
	data["agent_version"] = agentVersion
	return data
}

// maskSensitiveConfig 对敏感配置字段进行脱敏处理
func maskSensitiveConfig(config map[string]interface{}) map[string]interface{} {
	// 敏感字段关键词列表（不区分大小写）
	sensitiveKeywords := []string{
		"password", "passwd", "pwd",
		"secret", "token", "key",
		"apikey", "api_key",
		"access_key", "secret_key", "private_key",
	}

	// 创建副本避免修改原数据
	maskedConfig := make(map[string]interface{})
	for k, v := range config {
		maskedConfig[k] = v
	}

	// 对敏感字段进行脱敏（不区分大小写）
	for configKey := range maskedConfig {
		lowerKey := strings.ToLower(configKey)

		// 检查字段名是否包含敏感关键词
		isSensitive := false
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(lowerKey, keyword) {
				isSensitive = true
				break
			}
		}

		// 如果是敏感字段且值不为空，替换为 ******
		if isSensitive {
			if strVal, ok := maskedConfig[configKey].(string); ok && strVal != "" {
				maskedConfig[configKey] = "******"
			}
		}
	}

	return maskedConfig
}
