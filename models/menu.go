package models

import (
	"time"
)

// Menu 菜单表
type Menu struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name" gorm:"type:varchar(50);not null;comment:菜单名称"`
	Path       string    `json:"path" gorm:"type:varchar(100);comment:前端路由路径"`
	Component  string    `json:"component" gorm:"type:varchar(100);comment:前端组件路径"`
	Permission string    `json:"permission" gorm:"type:varchar(50);comment:权限标识"`
	ParentID   *uint     `json:"parent_id" gorm:"comment:父菜单ID"`
	Sort       int       `json:"sort" gorm:"default:0;comment:排序"`
	Icon       string    `json:"icon" gorm:"type:varchar(50);comment:图标"`
	IsVisible  bool      `json:"is_visible" gorm:"default:true;comment:是否可见"`
	IsEnabled  bool      `json:"is_enabled" gorm:"default:true;comment:是否启用"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}

// RoleMenu 角色-菜单关联表
type RoleMenu struct {
	RoleID    uint      `json:"role_id" gorm:"primarykey;comment:角色ID"`
	MenuID    uint      `json:"menu_id" gorm:"primarykey;comment:菜单ID"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "role_menus"
}

// MenuTree 菜单树结构
type MenuTree struct {
	Menu
	Children []MenuTree `json:"children,omitempty"`
}
