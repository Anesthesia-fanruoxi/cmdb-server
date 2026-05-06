package apply

import (
	"cmdb/utils"
	"net/http"
)

// ApplyProjects 获取SQL审批页面可访问的项目列表
// GET /api/sql/apply/projects
func ApplyProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/sql/apply")
}
