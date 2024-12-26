package models

import (
	"time"
)

// Role 角色表
type Role struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string `gorm:"type:varchar(32);uniqueIndex;not null"`
	Code        string `gorm:"type:varchar(32);uniqueIndex;not null"`
	Description string `gorm:"type:varchar(128)"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
