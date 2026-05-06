package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// ProjectDict 项目字典表
type CategoryDict struct {
	ID           int64     `json:"id" gorm:"primarykey"`
	Category     string    `json:"category" gorm:"size:64;uniqueIndex;comment:分类标识"`
	CategoryName string    `json:"category_name" gorm:"size:255;comment:分类名称"`
	CreatedBy    int64     `json:"created_by" gorm:"comment:创建人ID"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (CategoryDict) TableName() string {
	return middleware.TableCategoryDict
}
