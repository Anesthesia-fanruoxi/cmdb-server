package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// Department 部门表
type Department struct {
	ID          int64     `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"size:64;uniqueIndex;comment:部门名称"`
	Code        string    `json:"code" gorm:"size:64;uniqueIndex;comment:部门编码"`
	ParentID    int64     `json:"parent_id" gorm:"comment:父部门ID"`
	Description string    `json:"description" gorm:"size:255;comment:部门描述"`
	Sort        int       `json:"sort" gorm:"default:0;comment:排序"`
}

// TableName 指定表名
func (Department) TableName() string {
	return middleware.TableDepartments
}
