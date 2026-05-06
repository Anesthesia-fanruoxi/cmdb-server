package system

import (
	"cmdb/pkg/middleware"
)

// UserProject 用户项目关联表
// 用于跨部门协作时，给用户单独绑定项目并指定可访问的菜单
type UserProject struct {
	UserID  int64  `json:"user_id" gorm:"primaryKey;comment:用户ID"`
	Project string `json:"project" gorm:"primaryKey;size:64;comment:项目标识"`
	MenuID  int64  `json:"menu_id" gorm:"primaryKey;comment:关联菜单ID"`
}

// TableName 指定表名
func (UserProject) TableName() string {
	return middleware.TableUserProjects
}
