package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/initialize"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=50"`
	Path       string `json:"path" validate:"max=200"`
	Component  string `json:"component" validate:"max=255"`
	Permission string `json:"permission" validate:"max=100"`
	ParentID   uint   `json:"parent_id"`
	Sort       int    `json:"sort" validate:"min=0"`
	Icon       string `json:"icon" validate:"max=100"`
	IsVisible  bool   `json:"is_visible"`
	IsEnabled  bool   `json:"is_enabled"`
}

// MenuCreate 创建菜单
func MenuCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req CreateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 如果有父菜单，检查父菜单是否存在
	if req.ParentID != 0 {
		var parentCount int64
		database.DB.Model(&system.Menu{}).Where("id = ?", req.ParentID).Count(&parentCount)
		if parentCount == 0 {
			middleware.ResponseError(w, http.StatusBadRequest, "父菜单不存在")
			return
		}
	}

	// 创建菜单
	menu := system.Menu{
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Permission: req.Permission,
		ParentID:   int64(req.ParentID),
		Sort:       req.Sort,
		Icon:       req.Icon,
		IsVisible:  req.IsVisible,
		IsEnabled:  req.IsEnabled,
	}

	if err := database.DB.Create(&menu).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建菜单失败")
		return
	}
	var superAdminRole system.Role
	if err := database.DB.Where("level = ?", 0).First(&superAdminRole).Error; err != nil {
		return
	}
	// 创建菜单后
	if err := initialize.InitSuperAdminMenus(superAdminRole.ID); err != nil {
		fmt.Printf("初始化超级管理员菜单权限失败: %v\n", err)
	}

	middleware.ResponseSuccess(w, "创建成功", map[string]interface{}{
		"menu_id": menu.ID,
	})
}
