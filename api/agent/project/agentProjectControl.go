package project

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// ControlRequest 启停重启/卸载请求
type ControlRequest struct {
	Project    string `json:"project" validate:"required"`     // 项目标识
	PluginName string `json:"plugin_name" validate:"required"` // 插件名称
	Action     string `json:"action" validate:"required"`      // start/stop/restart/uninstall
}

// AgentProjectControl 插件控制操作（启动/停止/重启/卸载）
// POST /api/agent/project/control
func AgentProjectControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req ControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}
	if req.Project == "" || req.PluginName == "" || req.Action == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project、plugin_name、action不能为空")
		return
	}

	validActions := map[string]bool{
		"start": true, "stop": true, "restart": true, "uninstall": true,
	}
	if !validActions[req.Action] {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的action，支持: start, stop, restart, uninstall")
		return
	}

	agentURL, err := getAgentURL(req.Project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	var apiURL, method string
	switch req.Action {
	case "start", "stop", "restart":
		apiURL = fmt.Sprintf("%s/api/plugins/operate?name=%s&action=%s", agentURL, req.PluginName, req.Action)
		method = "PUT"
	case "uninstall":
		apiURL = fmt.Sprintf("%s/api/plugins/uninstall?name=%s", agentURL, req.PluginName)
		method = "DELETE"
	}

	agentResp, err := callAgent(method, apiURL, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendAgentResult(w, req.Project, req.PluginName, req.Action, agentResp)
}
