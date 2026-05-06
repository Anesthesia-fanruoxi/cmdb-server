package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// Permission 权限表
type Permission struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name" gorm:"size:50;uniqueIndex;comment:权限名称"` // 如：system:user:add
	Title     string    `json:"title" gorm:"size:50;comment:权限标题"`            // 如：创建用户
	Module    string    `json:"module" gorm:"size:50;comment:所属模块"`           // 如：system
	Remark    string    `json:"remark" gorm:"size:200;comment:备注"`            // 权限说明
}

// Role 角色表
type Role struct {
	ID          int64     `json:"id" gorm:"primaryKey;comment:角色ID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"size:32;uniqueIndex;comment:角色名称"` // 如：管理员
	Code        string    `json:"code" gorm:"size:32;uniqueIndex;comment:角色编码"` // 如：admin
	Description string    `json:"description" gorm:"size:128;comment:角色描述"`     // 角色描述
	Level       int64     `json:"level" gorm:"comment:角色级别"`                    // 新增字段：0超级管理员，1管理员，2普通用户
}

// TableName 指定表名
func (Role) TableName() string {
	return middleware.TableRoles
}
