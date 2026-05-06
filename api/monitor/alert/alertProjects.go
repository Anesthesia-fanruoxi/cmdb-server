package alert

import (
	"cmdb/utils"
	"net/http"
)

// AlertProjects 获取监控告警页面可访问的项目列表
// GET /api/monitor/alert/projects
func AlertProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/monitor/alert")
}
