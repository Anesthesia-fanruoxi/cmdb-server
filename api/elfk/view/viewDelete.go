package view

import (
	"cmdb/models/elfk"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"fmt"
	"net/http"
	"strconv"
)

// DeleteView 删除视图
func DeleteView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
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

	// 获取视图ID
	viewID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的视图ID")
		return
	}

	// 查询视图是否存在
	var view elfk.View
	if err := database.DB.Where("id = ?", viewID).First(&view).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "视图不存在")
		return
	}

	// 获取项目权限并验证
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 验证是否有权限访问该项目
	if !isSuperAdmin && !utils.Contains(projects, view.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 删除视图
	if err := database.DB.Delete(&view).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("删除视图失败: %v", err))
		return
	}
	middleware.ResponseSuccess(w, "删除视图成功", nil)
}
