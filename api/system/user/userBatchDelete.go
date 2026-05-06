package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

type UserBatchDeleteRequest struct {
	UserIDs []uint `json:"user_ids" validate:"required,min=1"`
}

// UserBatchDelete 批量删除用户
func UserBatchDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求体
	var req UserBatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 获取当前用户的角色级别
	var currentRole system.Role
	if err := database.DB.First(&currentRole, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 查询要删除的用户列表
	var users []system.User
	if err := tx.Find(&users, req.UserIDs).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户信息失败")
		return
	}

	// 检查每个用户的权限
	for _, user := range users {
		// 不能删除自己
		if userInfo.UserID == user.ID {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "不能删除自己")
			return
		}

		// 获取目标用户的角色信息
		var targetRole system.Role
		if err := tx.First(&targetRole, user.RoleID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "获取用户角色失败")
			return
		}

		// 不能删除超级管理员
		if targetRole.Level == 0 {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "不能删除超级管理员用户")
			return
		}

		// 非超级管理员只能删除自己部门的用户
		if currentRole.Level > 0 && user.DeptID != userInfo.DeptID {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusForbidden, "无权删除其他部门的用户")
			return
		}
	}

	// 批量删除用户
	if err := tx.Delete(&system.User{}, req.UserIDs).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除用户失败")
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(w, "删除成功", nil)
}
