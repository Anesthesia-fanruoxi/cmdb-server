package plugins

import (
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// AgentPluginsUpdateRequest 更新插件请求
type AgentPluginsUpdateRequest struct {
	ID             int64  `json:"id" validate:"required"` // 插件ID
	Version        string `json:"version"`                // 版本号（可修改）
	DisplayName    string `json:"display_name"`           // 显示名称
	Description    string `json:"description"`            // 插件描述
	Port           int    `json:"port"`                   // 服务端口
	ConfigTemplate string `json:"config_template"`        // 配置文件模板
}

// AgentPluginsUpdate 更新插件信息
func AgentPluginsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求参数
	var req AgentPluginsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}

	// 参数校验
	if req.ID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "id不能为空")
		return
	}

	// 查询插件是否存在
	var plugin agent.AgentStore
	if err := database.DB.First(&plugin, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "插件不存在")
		return
	}

	// 如果要修改版本号，需要检查新的name+version是否已被其他记录使用
	if req.Version != "" && req.Version != plugin.Version {
		var exists agent.AgentStore
		err := database.DB.Where("name = ? AND version = ? AND id != ?", plugin.Name, req.Version, req.ID).First(&exists).Error
		if err == nil {
			middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("版本 %s 已存在，不能重复", req.Version))
			return
		}
	}

	// 更新字段（只更新传了的字段）
	updates := make(map[string]interface{})

	if req.Version != "" {
		updates["version"] = req.Version
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Port > 0 {
		updates["port"] = req.Port
	}
	// ConfigTemplate 允许传空字符串（清空模板）
	if req.ConfigTemplate != "" || (req.ConfigTemplate == "" && r.FormValue("config_template") != "") {
		updates["config_template"] = req.ConfigTemplate
	}

	// 执行更新
	if len(updates) > 0 {
		if err := database.DB.Model(&plugin).Updates(updates).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "更新失败: "+err.Error())
			return
		}
	}

	// 重新查询获取最新数据
	database.DB.First(&plugin, req.ID)

	middleware.ResponseSuccess(w, "更新成功", plugin)
}
