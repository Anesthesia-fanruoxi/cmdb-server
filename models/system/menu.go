package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// Menu 菜单表
type Menu struct {
	ID         int64     `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name" gorm:"type:varchar(50);not null;comment:菜单名称"`
	Path       string    `json:"path" gorm:"type:varchar(200);comment:前端路由路径"`
	Component  string    `json:"component" gorm:"type:varchar(255);comment:前端组件路径"`
	Permission string    `json:"permission" gorm:"type:varchar(100);comment:权限标识"`
	ParentID   int64     `json:"parent_id" gorm:"default:0;comment:父菜单ID"`
	Sort       int       `json:"sort" gorm:"default:0;comment:排序"`
	Icon       string    `json:"icon" gorm:"type:varchar(100);comment:图标"`
	IsVisible  bool      `json:"is_visible" gorm:"default:true;comment:是否可见"`
	IsEnabled  bool      `json:"is_enabled" gorm:"default:true;comment:是否启用"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return middleware.TableMenus
}

// MenuTree 菜单树结构
type MenuTree struct {
	Menu
	Children []MenuTree `json:"children,omitempty"`
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	ID        int64     `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	RoleID    int64     `json:"role_id" gorm:"index;comment:角色ID"`
	MenuID    int64     `json:"menu_id" gorm:"index;comment:菜单ID"`
	HasMenu   bool      `json:"has_menu" gorm:"default:false;comment:是否有菜单权限"`
	HasRead   bool      `json:"has_read" gorm:"default:false;comment:是否有只读权限"`
	HasWrite  bool      `json:"has_write" gorm:"default:false;comment:是否有写入权限"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return middleware.TableRoleMenus
}
