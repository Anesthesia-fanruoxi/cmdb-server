package role

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"time"
)

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleList 获取角色列表
func RoleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取当前用户的角色级别
	var currentRole system.Role
	if err := database.DB.Select("level").First(&currentRole, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 查询角色列表
	var roles []RoleListResponse
	query := database.DB.Model(&system.Role{})

	// 非超级管理员只能查看级别比自己低的角色
	if currentRole.Level > 0 {
		query = query.Where("level > ?", currentRole.Level)
	}

	if err := query.Find(&roles).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询角色列表失败")
		return
	}

	middleware.ResponseSuccess(w, "查询成功", roles)
}

// RoleDetail 角色详情
func RoleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取角色ID
	roleID := r.URL.Query().Get("id")
	if roleID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "角色ID不能为空")
		return
	}

	// 查询角色
	var role system.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "角色不存在")
		return
	}

	// 返回角色信息
	middleware.ResponseSuccess(w, "查询成功", role)
}
