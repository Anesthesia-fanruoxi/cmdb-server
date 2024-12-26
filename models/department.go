package models

import (
	"time"
)

// Department 部门模型
type Department struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null;comment:部门名称"`
	Code        string    `json:"code" gorm:"type:varchar(50);not null;uniqueIndex;comment:部门编码"`
	Description string    `json:"description" gorm:"type:varchar(255);comment:描述"`
	ParentID    *uint     `json:"parent_id" gorm:"comment:父部门ID"`
	Sort        int       `json:"sort" gorm:"type:int;default:0;comment:排序"`
	IsEnabled   bool      `json:"is_enabled" gorm:"type:tinyint(1);default:1;comment:是否启用"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "departments"
}
