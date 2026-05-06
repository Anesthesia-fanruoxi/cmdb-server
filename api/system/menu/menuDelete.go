package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// MenuDelete 删除菜单
func MenuDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取菜单ID
	menuID := r.URL.Query().Get("id")
	if menuID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "菜单ID不能为空")
		return
	}

	// 转换菜单ID
	id, err := strconv.ParseInt(menuID, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的菜单ID")
		return
	}

	// 检查是否有子菜单
	var childCount int64
	if err := database.DB.Model(&system.Menu{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "检查子菜单失败")
		return
	}
	if childCount > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "存在子菜单，无法删除")
		return
	}

	// 查询菜单是否存在
	var menu system.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "菜单不存在")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 删除角色菜单关联
	if err := tx.Where("menu_id = ?", id).Delete(&system.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除角色菜单关联失败")
		return
	}

	// 删除菜单
	if err := tx.Delete(&menu).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除菜单失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除菜单失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
