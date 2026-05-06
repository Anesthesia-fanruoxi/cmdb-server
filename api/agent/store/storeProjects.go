package store

import (
	"cmdb/utils"
	"net/http"
)

// StoreProjects 获取Agent商店页面可访问的项目列表
// GET /api/agent/store/projects
func StoreProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/agent/store")
}
