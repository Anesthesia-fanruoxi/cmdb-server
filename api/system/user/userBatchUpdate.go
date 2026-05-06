package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

type UserBatchUpdateRequest struct {
	UserIDs            []uint `json:"user_ids" validate:"required,min=1"`
	AllowPasswordLogin *bool  `json:"allow_password_login"`
	DeptID             *uint  `json:"dept_id"`
	IsEnabled          *bool  `json:"is_enabled"`
	OnlineAssets       *bool  `json:"online_assets"`
	TestAssets         *bool  `json:"test_assets"`
}

// UserBatchUpdate 批量更新用户
func UserBatchUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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
	var req UserBatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 至少需要更新一个字段
	if req.DeptID == nil && req.IsEnabled == nil && req.OnlineAssets == nil && req.TestAssets == nil && req.AllowPasswordLogin == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "至少需要更新一个字段")
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

	// 查询要更新的用户列表
	var users []system.User
	if err := tx.Find(&users, req.UserIDs).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户信息失败")
		return
	}

	// 检查每个用户的权限
	for _, user := range users {
		// 获取目标用户的角色信息
		var targetRole system.Role
		if err := tx.First(&targetRole, user.RoleID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "获取用户角色失败")
			return
		}

		// 不能修改超级管理员
		if targetRole.Level == 0 {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "不能修改超级管理员用户")
			return
		}

		// 非超级管理员只能修改自己部门的用户
		if currentRole.Level > 0 && user.DeptID != userInfo.DeptID {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusForbidden, "无权修改其他部门的用户")
			return
		}
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.DeptID != nil {
		// 如果要更新部门，需要检查部门是否存在
		var dept system.Department
		if err := tx.First(&dept, *req.DeptID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "部门不存在")
			return
		}
		updates["dept_id"] = *req.DeptID
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.OnlineAssets != nil {
		updates["online_assets"] = *req.OnlineAssets
	}
	if req.TestAssets != nil {
		updates["test_assets"] = *req.TestAssets
	}
	if req.AllowPasswordLogin != nil {
		updates["allow_password_login"] = *req.AllowPasswordLogin
	}

	// 批量更新用户
	if err := tx.Model(&system.User{}).Where("id IN ?", req.UserIDs).Updates(updates).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新用户失败")
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(w, "更新成功", nil)
}
