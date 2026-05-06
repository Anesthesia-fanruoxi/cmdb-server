package project

import (
	"cmdb/config"
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpgradeRequest 版本升级请求
type UpgradeRequest struct {
	Project    string `json:"project" validate:"required"`     // 项目标识
	PluginName string `json:"plugin_name" validate:"required"` // 插件名称
}

// AgentProjectUpgrade 插件版本升级（自动从数据库取最新版本）
// POST /api/agent/project/upgrade
func AgentProjectUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}
	if req.Project == "" || req.PluginName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project、plugin_name不能为空")
		return
	}

	agentURL, err := getAgentURL(req.Project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 从数据库查询最新版本
	var latestPlugin agent.AgentStore
	if err := database.DB.Where("name = ?", req.PluginName).
		Order("created_at DESC").
		First(&latestPlugin).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, fmt.Sprintf("未找到插件 %s 的版本信息", req.PluginName))
		return
	}

	agentConfig := config.GetAgentConfig()
	payload := map[string]interface{}{
		"name":    req.PluginName,
		"version": latestPlugin.Version,
	}
	if latestPlugin.PluginType == "container" {
		payload["image"] = latestPlugin.GetImageURL(agentConfig.ImageUrl)
	} else if latestPlugin.PluginType == "binary" {
		payload["download_url"] = latestPlugin.GetDownloadURL(agentConfig.DownloadUrl)
	}

	agentResp, err := callAgent("PUT", agentURL+"/api/plugins/upgrade", payload)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendAgentResult(w, req.Project, req.PluginName, "upgrade", agentResp)
}
