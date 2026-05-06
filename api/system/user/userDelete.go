package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// UserDelete 删除用户
func UserDelete(w http.ResponseWriter, r *http.Request) {
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

	// 从URL获取要删除的用户ID
	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	// 不能删除自己
	if userInfo.UserID == middleware.ToInt64(userID) {
		middleware.ResponseError(w, http.StatusBadRequest, "不能删除自己")
		return
	}

	// 获取当前用户的角色级别
	var currentRole system.Role
	if err := database.DB.First(&currentRole, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 检查要删除的用户是否存在，并获取其信息
	var targetUser system.User
	if err := database.DB.First(&targetUser, userID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
		return
	}

	// 检查要删除的用户角色级别
	var targetRole system.Role
	if err := database.DB.First(&targetRole, targetUser.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取用户角色失败")
		return
	}

	// 如果是超级管理员（level=0），则不允许删除
	if targetRole.Level == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "超级管理员用户不能删除")
		return
	}

	// 非超级管理员只能删除自己部门的用户
	if currentRole.Level > 0 {
		if targetUser.DeptID != userInfo.DeptID {
			middleware.ResponseError(w, http.StatusForbidden, "无权删除其他部门的用户")
			return
		}
	}

	// 删除用户
	if err := database.DB.Delete(&targetUser).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除用户失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
