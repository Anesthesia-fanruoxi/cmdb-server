package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// ProjectMenuDetail 项目菜单关联详情
type ProjectMenuDetail struct {
	Project string  `json:"project"`
	MenuIDs []int64 `json:"menu_ids"`
}

// UserProjectDetailResponse 用户项目关联详情响应
type UserProjectDetailResponse struct {
	UserID       int64               `json:"user_id"`
	UserName     string              `json:"user_name"`
	DeptID       int64               `json:"dept_id"`
	ProjectMenus []ProjectMenuDetail `json:"project_menus"` // 用户直接关联的项目（按项目分组菜单ID）
	DeptProjects []string            `json:"dept_projects"` // 部门关联的项目
}

// UserProjectDetail 获取用户项目关联详情
// 返回用户直接关联的项目和部门关联的项目，方便前端分别展示
func UserProjectDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取用户ID参数
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少user_id参数")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的user_id")
		return
	}

	// 检查用户是否存在
	var user system.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
		return
	}

	// 获取用户直接关联的项目（含菜单ID）
	var userProjects []system.UserProject
	if err := database.DB.Where("user_id = ?", userID).Find(&userProjects).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户项目失败")
		return
	}

	// 按项目分组菜单ID
	projectMenuMap := make(map[string][]int64)
	for _, up := range userProjects {
		projectMenuMap[up.Project] = append(projectMenuMap[up.Project], up.MenuID)
	}

	// 转换为响应格式
	projectMenus := make([]ProjectMenuDetail, 0, len(projectMenuMap))
	for project, menuIDs := range projectMenuMap {
		projectMenus = append(projectMenus, ProjectMenuDetail{
			Project: project,
			MenuIDs: menuIDs,
		})
	}

	// 获取用户部门关联的项目
	var deptProjects []string
	if err := database.DB.Model(&system.DeptProject{}).
		Where("dept_id = ?", user.DeptID).
		Pluck("project", &deptProjects).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询部门项目失败")
		return
	}

	// 构造响应
	response := UserProjectDetailResponse{
		UserID:       user.ID,
		UserName:     user.UserName,
		DeptID:       user.DeptID,
		ProjectMenus: projectMenus,
		DeptProjects: deptProjects,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
