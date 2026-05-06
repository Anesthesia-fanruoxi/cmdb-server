package project

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// ConfigUpdateRequest 配置增量更新请求
type ConfigUpdateRequest struct {
	Project      string                 `json:"project" validate:"required"`     // 项目标识
	PluginName   string                 `json:"plugin_name" validate:"required"` // 插件名称
	ConfigSet    map[string]interface{} `json:"config_set"`                      // 新增或修改的配置项（upsert）
	ConfigDelete []string               `json:"config_delete"`                   // 要删除的配置 key 列表
}

// AgentProjectConfig 插件配置增量更新
// POST /api/agent/project/config
func AgentProjectConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}
	if req.Project == "" || req.PluginName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project、plugin_name不能为空")
		return
	}
	if len(req.ConfigSet) == 0 && len(req.ConfigDelete) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "config_set 和 config_delete 不能同时为空")
		return
	}

	agentURL, err := getAgentURL(req.Project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	payload := map[string]interface{}{
		"name": req.PluginName,
	}
	if len(req.ConfigSet) > 0 {
		payload["config_set"] = req.ConfigSet
	}
	if len(req.ConfigDelete) > 0 {
		payload["config_delete"] = req.ConfigDelete
	}

	agentResp, err := callAgent("PUT", agentURL+"/api/plugins/update", payload)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendAgentResult(w, req.Project, req.PluginName, "config_update", agentResp)
}
