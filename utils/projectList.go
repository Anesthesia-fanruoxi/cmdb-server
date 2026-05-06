package utils

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	ID          int64  `json:"id"`
	Project     string `json:"project"`
	ProjectName string `json:"project_name"`
}

// GetProjectListByMenuPath 根据菜单路径获取可访问的项目列表
// menuPath: 业务菜单路径，如 "/assets/proUpdate"
// includeOps: 可选参数，是否包含ops项目，默认不包含
// 返回项目列表（包含项目标识和项目名称）
func GetProjectListByMenuPath(w http.ResponseWriter, r *http.Request, menuPath string, includeOps ...bool) {
	// 是否包含ops项目，默认不包含
	showOps := false
	if len(includeOps) > 0 && includeOps[0] {
		showOps = true
	}
	// 获取当前用户信息
	userInfo, ok := r.Context().Value(UserInfoKey).(UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 创建权限检查实例，使用指定的菜单路径
	auth := ProjectAuth(
		uint(userInfo.UserID),
		uint(userInfo.DeptID),
		uint(userInfo.RoleID),
		role.Level,
		userInfo.OnlineAssets,
		userInfo.TestAssets,
		menuPath,
	)

	// 获取可访问项目列表
	projects, err := auth.GetAccessibleProjects("prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目失败")
		return
	}

	// 如果是超级管理员或没有指定菜单，返回所有项目
	if role.Level == 0 {
		var allProjects []system.ProjectDict
		query := database.DB
		if !showOps {
			query = query.Where("project != ?", "ops")
		}
		if err := query.Find(&allProjects).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询项目字典失败")
			return
		}
		result := make([]ProjectListResponse, 0, len(allProjects))
		for _, p := range allProjects {
			result = append(result, ProjectListResponse{
				ID:          p.ID,
				Project:     p.Project,
				ProjectName: p.ProjectName,
			})
		}
		middleware.ResponseSuccess(w, "获取成功", result)
		return
	}

	// 查询项目详情
	if len(projects) == 0 {
		middleware.ResponseSuccess(w, "获取成功", []ProjectListResponse{})
		return
	}

	var projectDicts []system.ProjectDict
	query := database.DB.Where("project IN ?", projects)
	if !showOps {
		query = query.Where("project != ?", "ops")
	}
	if err := query.Find(&projectDicts).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目字典失败")
		return
	}

	result := make([]ProjectListResponse, 0, len(projectDicts))
	for _, p := range projectDicts {
		result = append(result, ProjectListResponse{
			ID:          p.ID,
			Project:     p.Project,
			ProjectName: p.ProjectName,
		})
	}

	middleware.ResponseSuccess(w, "获取成功", result)
}
