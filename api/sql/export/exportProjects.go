package export

import (
	"cmdb/utils"
	"net/http"
)

// ExportProjects 获取SQL导出页面可访问的项目列表
// GET /api/sql/export/projects
func ExportProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/sql/export")
}
