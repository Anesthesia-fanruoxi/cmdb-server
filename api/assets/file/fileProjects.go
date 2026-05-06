package file

import (
	"cmdb/utils"
	"net/http"
)

// FileProjects 获取文件管理页面可访问的项目列表
// GET /api/assets/file/projects
func FileProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/file")
}
