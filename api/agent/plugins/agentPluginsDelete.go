package plugins

import (
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// AgentPluginsDelete 删除插件
func AgentPluginsDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	// 查询插件是否存在
	var plugin agent.AgentStore
	if err := database.DB.First(&plugin, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "插件不存在")
		return
	}

	// 删除插件
	if err := database.DB.Delete(&plugin).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"id":      plugin.ID,
		"name":    plugin.Name,
		"version": plugin.Version,
	}

	middleware.ResponseSuccess(w, "删除成功", response)
}
