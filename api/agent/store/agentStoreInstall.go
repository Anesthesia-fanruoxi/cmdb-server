package store

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/agent"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AgentStoreInstallRequest 安装插件请求
type AgentStoreInstallRequest struct {
	PluginID      int64                  `json:"plugin_id" validate:"required"` // 插件ID（必填）
	Project       string                 `json:"project" validate:"required"`   // 项目标识（必填）
	ContainerPort int                    `json:"container_port"`                // 宿主机映射端口（容器类型必填）
	Config        map[string]interface{} `json:"config"`                        // 配置参数（容器类型可选）
	Command       string                 `json:"command"`                       // 启动命令（二进制类型可选）
	ConfigContent string                 `json:"config_content"`                // 配置文件内容（二进制类型可选，渲染后的配置）
}

// AgentInstallRequest Agent安装接口请求（对应agent的API）
type AgentInstallRequest struct {
	Name        string                 `json:"name"`                   // 插件名称
	Version     string                 `json:"version"`                // 版本号
	Category    string                 `json:"category"`               // 插件类型(container/binary)
	Image       string                 `json:"image,omitempty"`        // Docker镜像(container类型)
	DownloadURL string                 `json:"download_url,omitempty"` // 下载地址(binary类型)
	Port        int                    `json:"port,omitempty"`         // 宿主机端口(container类型必填，binary类型不传)
	Command     string                 `json:"command,omitempty"`      // 启动命令
	Config      map[string]interface{} `json:"config,omitempty"`       // 配置参数（环境变量等）
	ConfigFile  string                 `json:"config_file"`            // 配置文件内容（binary类型，渲染后的配置）
	Parameters  map[string]interface{} `json:"parameters,omitempty"`   // 扩展参数
}

// AgentInstallResponse Agent安装接口响应
type AgentInstallResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// AgentStoreInstall 从商店安装插件到项目
func AgentStoreInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求参数
	var req AgentStoreInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}

	// 参数校验
	if req.PluginID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "plugin_id不能为空")
		return
	}
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project不能为空")
		return
	}

	// 校验项目权限
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	projects, _, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil || !utils.Contains(projects, req.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权限操作该项目")
		return
	}

	// 查询插件信息
	var plugin agent.AgentStore
	if err := database.DB.First(&plugin, req.PluginID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "插件不存在")
		return
	}

	// 获取项目Agent URL
	agentURL, err := system.GetProjectAgentURL(req.Project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目Agent地址失败: %v", err))
		return
	}

	// 容器类型需要端口，二进制类型不需要
	containerPort := req.ContainerPort
	if plugin.PluginType == "container" && containerPort == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "容器类型插件必须指定宿主机映射端口(container_port)")
		return
	}

	// 获取Agent配置
	agentConfig := config.GetAgentConfig()

	// 构建Agent请求（按照agent API格式）
	agentReq := AgentInstallRequest{
		Name:        plugin.Name,
		Version:     plugin.Version,
		Category:    plugin.PluginType,                              // container 或 binary
		DownloadURL: plugin.GetDownloadURL(agentConfig.DownloadUrl), // 所有类型都传递download_url
	}

	// 根据插件类型传递不同参数
	if plugin.PluginType == "container" {
		// === 容器类型：传递 image_url, container_port, config ===

		// 1. 镜像地址（必填）
		if agentConfig.ImageUrl == "" {
			middleware.ResponseError(w, http.StatusInternalServerError, "配置错误：未配置container_registry")
			return
		}
		agentReq.Image = plugin.GetImageURL(agentConfig.ImageUrl)

		// 2. 宿主机映射端口（必填）
		agentReq.Port = containerPort

		// 3. 环境变量配置（可选）
		agentReq.Config = req.Config

		// 4. 容器内端口（添加到parameters）
		if plugin.Port > 0 {
			agentReq.Parameters = map[string]interface{}{
				"container_port": plugin.Port, // 容器内端口
			}
		}

	} else if plugin.PluginType == "binary" {
		// === 二进制类型：传递 download_url, command, config_file ===

		// 1. 下载地址（已在上面统一设置）
		if agentConfig.DownloadUrl == "" {
			middleware.ResponseError(w, http.StatusInternalServerError, "配置错误：未配置binary_base_url")
			return
		}

		// 2. 启动命令（可选，仅当用户传递时才设置）
		if req.Command != "" {
			agentReq.Command = req.Command
		}

		// 3. 配置文件内容（可选，渲染后的配置文件）
		agentReq.ConfigFile = req.ConfigContent
		if req.ConfigContent != "" {
			middleware.Logger.Infof("传递配置文件内容 [Plugin: %s, 长度: %d]", plugin.Name, len(req.ConfigContent))
		} else {
			middleware.Logger.Warnf("未收到配置文件内容 [Plugin: %s]", plugin.Name)
		}

	} else {
		middleware.ResponseError(w, http.StatusBadRequest, "不支持的插件类型: "+plugin.PluginType)
		return
	}

	// 序列化请求体
	payloadBytes, err := json.Marshal(agentReq)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "构建请求失败: "+err.Error())
		return
	}

	// 打印原始请求体
	middleware.Logger.Infof("============== 发送给Agent的原始请求体 ==============")
	middleware.Logger.Infof("插件类型: %s", plugin.PluginType)
	middleware.Logger.Infof("插件名称: %s", plugin.Name)
	middleware.Logger.Infof("Agent URL: %s", agentURL+"/api/plugins/install")
	middleware.Logger.Infof("请求体JSON: %s", string(payloadBytes))

	// 格式化打印（便于阅读）
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, payloadBytes, "", "  "); err == nil {
		middleware.Logger.Infof("格式化请求体:\n%s", prettyJSON.String())
	}
	middleware.Logger.Infof("======================================================")

	// 调用Agent安装接口
	agentAPIURL := agentURL + "/api/plugins/install"
	resp, err := http.Post(agentAPIURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("调用Agent失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 读取Agent响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取Agent响应失败: "+err.Error())
		return
	}

	// 解析Agent响应
	var agentResp AgentInstallResponse
	if err := json.Unmarshal(body, &agentResp); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析Agent响应失败: "+err.Error())
		return
	}

	// 检查Agent响应状态
	if agentResp.Code != 200 {
		msg := agentResp.Message
		if msg == "" {
			msg = "安装失败"
		}
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("Agent安装失败: %s", msg))
		return
	}

	// 过滤Agent返回的敏感配置信息
	safeAgentData := filterSensitiveData(agentResp.Data)

	// 构建成功响应
	response := map[string]interface{}{
		"plugin_name":  plugin.Name,
		"plugin_type":  plugin.PluginType,
		"version":      plugin.Version,
		"project":      req.Project,
		"agent_data":   safeAgentData, // 已过滤敏感信息的数据
		"installed_by": userInfo.UserID,
	}

	// 容器类型才返回端口信息
	if plugin.PluginType == "container" {
		response["container_port"] = containerPort // 宿主机端口
		response["service_port"] = plugin.Port     // 服务端口（容器内端口）
	}

	middleware.ResponseSuccess(w, "安装成功", response)
}

// filterSensitiveData 过滤敏感配置信息
func filterSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return data
	}

	// 创建副本避免修改原数据
	safeCopy := make(map[string]interface{})
	for k, v := range data {
		safeCopy[k] = v
	}

	// 删除敏感配置字段
	delete(safeCopy, "config")

	return safeCopy
}
