package search

import (
	"cmdb/utils"
	"net/http"
)

// SearchProjects 获取ELFK搜索页面可访问的项目列表
// GET /api/elfk/search/projects
func SearchProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/elfk/search")
}
