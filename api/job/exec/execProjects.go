package exec

import (
	"cmdb/utils"
	"net/http"
)

// ExecProjects 获取Job执行记录页面可访问的项目列表
// GET /api/job/exec/projects
func ExecProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/job/exec")
}
