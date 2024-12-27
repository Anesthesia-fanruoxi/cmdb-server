package models

import (
	"time"
)

// Department 部门模型
type Department struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Code        string    `json:"code" gorm:"not null;unique"`
	Description string    `json:"description"`
	ParentID    *uint     `json:"parent_id"`
	Sort        int       `json:"sort" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "departments"
}
