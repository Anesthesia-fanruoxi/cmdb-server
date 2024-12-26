package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// GetMenuTree 获取菜单树
func GetMenuTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	db := initData.GetDB()
	var menus []models.Menu
	if err := db.Order("sort").Find(&menus).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取菜单列表失败")
		return
	}

	// 构建菜单树
	menuTree := buildMenuTree(menus, nil)
	utils.Success(w, menuTree)
}

// GetUserMenus 获取当前用户的菜单权限
func GetUserMenus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 从header中获取用户ID
	userIDStr := r.Header.Get("UserID")
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, utils.ERROR, "未获取到用户信息")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, utils.ERROR, "用户ID格式错误")
		return
	}

	// 查询用户信息获取角色ID
	db := initData.GetDB()
	var user models.User
	if err := db.First(&user, uint(userID)).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户信息失败")
		return
	}

	// 查询用户角色可以访问的菜单
	var menus []models.Menu
	if err := db.Table("menus").
		Select("menus.*").
		Joins("JOIN role_menus ON menus.id = role_menus.menu_id").
		Where("role_menus.role_id = ? AND menus.is_visible = 1 AND menus.is_enabled = 1", user.RoleID).
		Order("menus.sort").
		Find(&menus).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户菜单失败")
		return
	}

	// 构建菜单树
	menuTree := buildMenuTree(menus, nil)
	utils.Success(w, menuTree)
}

// CreateMenu 创建菜单
func CreateMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无��的请求参数")
		return
	}

	db := initData.GetDB()
	if err := db.Create(&menu).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建菜单失败")
		return
	}

	utils.Success(w, menu)
}

// UpdateMenu 更新菜单
func UpdateMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if menu.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "菜单ID不能为空")
		return
	}

	db := initData.GetDB()
	if err := db.Model(&menu).Updates(menu).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新菜单失败")
		return
	}

	utils.Success(w, menu)
}

// DeleteMenu 删除菜单
func DeleteMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 获取菜单ID
	menuID := r.URL.Query().Get("id")
	if menuID == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "菜单ID不能为空")
		return
	}

	id, err := strconv.ParseUint(menuID, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的菜单ID")
		return
	}

	db := initData.GetDB()

	// 检查是否有子菜单
	var count int64
	if err := db.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "检查子菜单失败")
		return
	}

	if count > 0 {
		utils.Error(w, http.StatusBadRequest, utils.ERROR, "该菜单下有子菜单，不能直接删除")
		return
	}

	// 开启事务
	tx := db.Begin()

	// 删除角色-菜单关联关系
	if err := tx.Delete(&models.RoleMenu{}, "menu_id = ?", id).Error; err != nil {
		tx.Rollback()
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除菜单权限关联失败")
		return
	}

	// 删除菜单
	if err := tx.Delete(&models.Menu{}, id).Error; err != nil {
		tx.Rollback()
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除菜单失败")
		return
	}

	tx.Commit()
	utils.Success(w, nil)
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []models.Menu, parentID *uint) []models.MenuTree {
	var tree []models.MenuTree
	for _, menu := range menus {
		if (parentID == nil && menu.ParentID == nil) || (parentID != nil && menu.ParentID != nil && *menu.ParentID == *parentID) {
			node := models.MenuTree{
				Menu:     menu,
				Children: buildMenuTree(menus, &menu.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}
