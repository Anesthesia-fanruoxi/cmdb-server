package plugins

import (
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// AgentPluginsCreateRequest 创建插件请求
type AgentPluginsCreateRequest struct {
	Name           string `json:"name" validate:"required"`         // 插件名称
	Version        string `json:"version" validate:"required"`      // 版本号
	DisplayName    string `json:"display_name" validate:"required"` // 显示名称
	PluginType     string `json:"plugin_type" validate:"required"`  // 插件类型(container/binary)
	Description    string `json:"description"`                      // 插件描述
	Port           int    `json:"port"`                             // 服务端口
	ConfigTemplate string `json:"config_template"`                  // 配置文件模板
}

// AgentPluginsCreate 创建插件
func AgentPluginsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求参数
	var req AgentPluginsCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}

	// 参数校验
	if req.Name == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "name不能为空")
		return
	}
	if req.Version == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "version不能为空")
		return
	}
	if req.DisplayName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "display_name不能为空")
		return
	}
	if req.PluginType == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "plugin_type不能为空")
		return
	}

	// 验证插件类型
	if req.PluginType != "container" && req.PluginType != "binary" {
		middleware.ResponseError(w, http.StatusBadRequest, "plugin_type只能是container或binary")
		return
	}

	// 检查插件是否已存在（name+version唯一）
	var exists agent.AgentStore
	if err := database.DB.Where("name = ? AND version = ?", req.Name, req.Version).First(&exists).Error; err == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "插件已存在，不能重复创建")
		return
	}

	// 创建插件记录
	plugin := agent.AgentStore{
		Name:           req.Name,
		Version:        req.Version,
		DisplayName:    req.DisplayName,
		PluginType:     req.PluginType,
		Description:    req.Description,
		Port:           req.Port,
		ConfigTemplate: req.ConfigTemplate,
	}

	if err := database.DB.Create(&plugin).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建插件失败: "+err.Error())
		return
	}

	// 获取用户信息
	userInfo, _ := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)

	// 构建响应
	response := map[string]interface{}{
		"id":           plugin.ID,
		"name":         plugin.Name,
		"version":      plugin.Version,
		"display_name": plugin.DisplayName,
		"plugin_type":  plugin.PluginType,
		"created_by":   userInfo.UserID,
	}

	middleware.ResponseSuccess(w, "创建成功", response)
}
