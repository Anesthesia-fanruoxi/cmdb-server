package kubePod

import (
	"cmdb/utils"
	"net/http"
)

// KubePodProjects 获取Pod管理页面可访问的项目列表
// GET /api/assets/kubePod/projects
func KubePodProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/kubePod")
}
