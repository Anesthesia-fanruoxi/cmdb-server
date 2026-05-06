package view

import (
	"cmdb/utils"
	"net/http"
)

// ViewProjects 获取ELFK视图页面可访问的项目列表
// GET /api/elfk/view/projects
func ViewProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/elfk/view")
}
