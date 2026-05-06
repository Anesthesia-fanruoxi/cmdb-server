package role

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// RoleDelete 删除角色
func RoleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取角色ID
	roleID := r.URL.Query().Get("id")
	if roleID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "角色ID不能为空")
		return
	}

	// 查询角色是否存在
	var role system.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "角色不存在")
		return
	}

	// 不允许删除超级管理员角色
	if role.Level == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "不能删除超级管理员角色")
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 1. 将使用该角色的用户角色更新为普通用户角色
	if err := tx.Model(&system.User{}).
		Where("role_id = ?", roleID).
		Update("role_id", 2). // 假设 2 是普通用户角色的ID
		Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新用户角色失败")
		return
	}

	// 2. 删除角色的菜单权限关联
	if err := tx.Where("role_id = ?", roleID).Delete(&system.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除角色权限关联失败")
		return
	}

	// 3. 删除角色
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除角色失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "提交事务失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
