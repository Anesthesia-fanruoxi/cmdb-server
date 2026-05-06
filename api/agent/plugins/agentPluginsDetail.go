package plugins

import (
	"cmdb/config"
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// AgentPluginsDetail 获取插件详情
func AgentPluginsDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取插件ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "id参数不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "id参数格式错误")
		return
	}

	// 查询插件信息
	var plugin agent.AgentStore
	if err := database.DB.First(&plugin, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "插件不存在")
		return
	}

	// 获取Agent配置
	agentConfig := config.GetAgentConfig()

	// 构建响应，添加动态生成的image/download_url
	response := map[string]interface{}{
		"id":           plugin.ID,
		"name":         plugin.Name,
		"version":      plugin.Version,
		"display_name": plugin.DisplayName,
		"plugin_type":  plugin.PluginType,
		"description":  plugin.Description,
		"port":         plugin.Port,
		"created_at":   plugin.CreatedAt,
		"updated_at":   plugin.UpdatedAt,
	}

	// 根据类型添加对应的URL和配置
	if plugin.PluginType == "container" {
		response["image_url"] = plugin.GetImageURL(agentConfig.ImageUrl)
	} else if plugin.PluginType == "binary" {
		response["download_url"] = plugin.GetDownloadURL(agentConfig.DownloadUrl)
		// 只有二进制插件才返回配置模板
		response["config_template"] = plugin.ConfigTemplate
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
