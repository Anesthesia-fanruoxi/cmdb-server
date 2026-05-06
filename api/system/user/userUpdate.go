package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserUpdate 更新用户信息
func UserUpdate(w http.ResponseWriter, r *http.Request) {
	// 添加调试日志

	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		fmt.Printf("获取用户信息失败，Context: %+v\n", r.Context())
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求体为map，支持任意字段更新
	var reqMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqMap); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 获取要更新的用户ID
	userIDFloat, ok := reqMap["id"].(float64)
	if !ok {
		// 尝试从上下文获取当前用户ID
		userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
		if !ok {
			middleware.ResponseError(w, http.StatusBadRequest, "用户ID不能为空")
			return
		}
		userIDFloat = float64(userInfo.UserID)
	}

	userID := int64(userIDFloat)

	// 获取当前用户的角色级别
	var currentRole system.Role
	if err := database.DB.Select("level").First(&currentRole, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 查询要更新的用户
	var user system.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
		return
	}
	// 不允许修改超管
	if userID == 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "不能修改超级管理员用户")
		return
	}

	// 权限检查
	if currentRole.Level > 0 {
		// 非超级管理员只能修改同部门用户
		if middleware.ToInt64(user.DeptID) != userInfo.DeptID {
			middleware.ResponseError(w, http.StatusForbidden, "无权修改其他部门的用户")
			return
		}

		// 检查角色级别
		var targetRole system.Role
		if err := database.DB.Select("level").First(&targetRole, user.RoleID).Error; err != nil {
			middleware.ResponseError(w, http.StatusForbidden, "目标用户角色无效")
			return
		}
		if targetRole.Level <= currentRole.Level {
			middleware.ResponseError(w, http.StatusForbidden, "无权修改同级别或更高级别的用户")
			return
		}
	}

	// 开始事务
	tx := database.DB.Begin()

	// 构建更新map
	updates := make(map[string]interface{})

	// 处理每个可能的字段
	if v, ok := reqMap["password"].(string); ok && v != "" {
		// 加密新密码
		hashedPassword, err := utils.EncryptPassword(v)
		if err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "密码加密失败")
			return
		}
		updates["password"] = hashedPassword
	}
	if v, ok := reqMap["nick_name"].(string); ok && v != "" {
		updates["nick_name"] = v
	}
	if v, ok := reqMap["email"].(string); ok && v != "" {
		updates["email"] = v
	}
	if v, ok := reqMap["phone"].(string); ok && v != "" {
		updates["phone"] = v
	}
	if v, ok := reqMap["is_enabled"].(bool); ok {
		updates["is_enabled"] = v
	}
	if v, ok := reqMap["allow_password_login"].(bool); ok {
		updates["allow_password_login"] = v
	}
	if v, ok := reqMap["online_assets"].(bool); ok {
		updates["online_assets"] = v
	}
	if v, ok := reqMap["test_assets"].(bool); ok {
		updates["test_assets"] = v
	}
	if v, ok := reqMap["role_id"].(float64); ok && v > 0 {
		roleID := int64(v)
		// 不允许分配为系统角色（ID=1）
		if roleID == 1 {
			middleware.ResponseError(w, http.StatusBadRequest, "不允许分配系统角色")
			return
		}
		// 检查新角色是否存在
		var newRole system.Role
		if err := tx.First(&newRole, roleID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "指定的角色不存在")
			return
		}
		// 检查是否是系统级别角色（level=0）
		if newRole.Level == 0 {
			middleware.ResponseError(w, http.StatusBadRequest, "不允许分配系统级别角色")
			return
		}
		updates["role_id"] = roleID
	}
	if v, ok := reqMap["dept_id"].(float64); ok && v > 0 {
		deptID := int64(v)
		// 不允许分配为系统部门（ID=1）
		if deptID == 1 {
			middleware.ResponseError(w, http.StatusBadRequest, "不允许分配系统部门")
			return
		}
		// 检查新部门是否存在
		var dept system.Department
		if err := tx.First(&dept, deptID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "指定的部门不存在")
			return
		}
		updates["dept_id"] = deptID
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 自动更新时间
	updates["updated_at"] = time.Now()

	// 执行更新
	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新用户失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "提交事务失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
