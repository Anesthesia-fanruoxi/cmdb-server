package view

import (
	"cmdb/utils"
	"net/http"
)

// ViewProjects 获取监控视图页面可访问的项目列表
// GET /api/monitor/view/projects
func ViewProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/monitor/view")
}
