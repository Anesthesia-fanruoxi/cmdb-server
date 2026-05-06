package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// MenuResponse 菜单响应结构
type MenuResponse struct {
	ID        int64  `json:"id"`
	ParentID  int64  `json:"parent_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Sort      int    `json:"sort"`
	Icon      string `json:"icon"`
	IsEnabled bool   `json:"is_enabled"`
}

// MenusList 获取菜单列表
func MenuList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 查询所有菜单，按照排序字段升序排列
	var menus []MenuResponse
	if err := database.DB.Model(&system.Menu{}).Order("sort asc").Find(&menus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询菜单列表失败")
		return
	}

	middleware.ResponseSuccess(w, "查询成功", menus)
}
