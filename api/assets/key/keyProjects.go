package key

import (
	"cmdb/utils"
	"net/http"
)

// KeyProjects 获取密钥管理页面可访问的项目列表
// GET /api/assets/key/projects
func KeyProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/key")
}
