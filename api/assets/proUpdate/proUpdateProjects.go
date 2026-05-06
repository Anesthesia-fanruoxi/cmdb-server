package proUpdate

import (
	"cmdb/utils"
	"net/http"
)

// ProUpdateProjects 获取生产发版页面可访问的项目列表
// GET /api/assets/proUpdate/projects
func ProUpdateProjects(w http.ResponseWriter, r *http.Request) {
	// 使用当前业务菜单路径查询项目
	utils.GetProjectListByMenuPath(w, r, "/assets/proUpdate")
}
