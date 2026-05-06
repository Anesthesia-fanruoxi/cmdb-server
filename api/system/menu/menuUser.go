package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"sort"
)

// UserMenuResponse 用户菜单响应
type UserMenuResponse struct {
	system.Menu
	Children []UserMenuResponse `json:"children"`
	Perms    []string           `json:"perms"`
}

// MenuUser 获取用户菜单树
func MenuUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 查询用户的菜单权限
	var roleMenus []system.RoleMenu
	if err := database.DB.Where("role_id = ? AND has_menu = ?", userInfo.RoleID, true).Find(&roleMenus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户菜单权限失败")
		return
	}

	// 获取有权限的菜单ID列表
	menuIDs := make([]int64, 0, len(roleMenus))
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuID)
	}

	// 查询这些菜单的详细信息
	var menus []system.Menu
	if err := database.DB.Where("id IN ? AND is_enabled = ?", menuIDs, true).
		Order("sort").Find(&menus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询菜单失败")
		return
	}

	// 构建菜单树
	menuTree := buildUserMenuTree(menus, nil)

	middleware.ResponseSuccess(w, "获取成功", menuTree)
}

// buildUserMenuTree 构建用户菜单树
func buildUserMenuTree(menus []system.Menu, parentID *uint) []system.MenuTree {
	var tree []system.MenuTree
	for _, menu := range menus {
		// 将 int64 转换为 *uint 进行比较
		var menuParentID *uint
		if menu.ParentID != 0 {
			tmp := uint(menu.ParentID)
			menuParentID = &tmp
		}

		if (parentID == nil && menuParentID == nil) ||
			(parentID != nil && menuParentID != nil && *menuParentID == *parentID) {
			uMenuId := uint(menu.ID)
			node := system.MenuTree{
				Menu:     menu,
				Children: buildUserMenuTree(menus, &uMenuId),
			}
			tree = append(tree, node)
		}
	}

	// 按 sort 字段排序
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Sort < tree[j].Sort
	})

	return tree
}
