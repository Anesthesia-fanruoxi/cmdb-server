package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"sort"
)

// MenuTree 获取菜单树
func MenuTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 查询所有菜单
	var menus []system.Menu
	if err := database.DB.Order("sort").Find(&menus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询菜单失败")
		return
	}

	// 构建菜单树
	menuTree := buildMenuTree(menus, nil)
	middleware.ResponseSuccess(w, "获取成功", menuTree)
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []system.Menu, parentID *uint) []system.MenuTree {
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
				Children: buildMenuTree(menus, &uMenuId),
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
