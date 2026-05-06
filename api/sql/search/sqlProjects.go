package search

import (
	"cmdb/utils"
	"net/http"
)

// SqlProjects 获取SQL搜索页面可访问的项目列表
// GET /api/sql/search/projects
func SqlProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/sql/search")
}
