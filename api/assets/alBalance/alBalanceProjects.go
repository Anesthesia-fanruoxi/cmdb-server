package alBalance

import (
	"cmdb/utils"
	"net/http"
)

// AlBalanceProjects 获取阿里云余额页面可访问的项目列表
// GET /api/assets/alBalance/projects
func AlBalanceProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/alBalance")
}
