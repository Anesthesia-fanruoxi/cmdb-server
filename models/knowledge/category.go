package knowledge

import (
	"gorm.io/gorm"
	"time"
)

// Category 文档分类
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Description string         `gorm:"size:200" json:"description"`
	ParentID    uint           `gorm:"default:0" json:"parent_id"`
	Sort        int            `gorm:"default:0" json:"sort"`
	user_name   uint           `gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "knowledge_categories"
}
