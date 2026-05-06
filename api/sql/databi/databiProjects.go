package databi

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// DataBIProjects 获取可访问的项目列表（仅返回dwd项目）
// GET /api/sql/databi/projects
func DataBIProjects(w http.ResponseWriter, r *http.Request) {
	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 检查用户是否有dwd项目权限
	hasDwdPermission := isSuperAdmin || utils.Contains(projects, "dwd")
	if !hasDwdPermission {
		middleware.ResponseSuccess(w, "获取成功", []utils.ProjectListResponse{})
		return
	}

	// 查询dwd项目信息
	var dwdProject system.ProjectDict
	if err := database.DB.Where("project = ?", "dwd").First(&dwdProject).Error; err != nil {
		middleware.ResponseSuccess(w, "获取成功", []utils.ProjectListResponse{})
		return
	}

	// 返回dwd项目信息
	result := []utils.ProjectListResponse{
		{
			ID:          dwdProject.ID,
			Project:     dwdProject.Project,
			ProjectName: dwdProject.ProjectName,
		},
	}

	middleware.ResponseSuccess(w, "获取成功", result)
}
