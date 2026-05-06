package knowledge

import (
	"cmdb/pkg/middleware"
	"time"
)

// Personal 个人文档
type Personal struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null;comment:文档标题"`
	Category  string    `json:"category" gorm:"type:varchar(50);not null;comment:文档分类"`
	Content   string    `json:"content" gorm:"type:text;comment:文档内容"`
	UserName  string    `json:"user_name" gorm:"type:varchar(64);not null;comment:最后更新者名称"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Personal) TableName() string {
	return middleware.TablePersonal
}

// PersonalRecycle 个人文档回收站
type PersonalRecycle struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	DocID     int64     `json:"doc_id" gorm:"type:bigint unsigned;not null;comment:文档ID"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null;comment:文档标题"`
	Category  string    `json:"category" gorm:"type:varchar(50);not null;comment:文档分类"`
	Content   string    `json:"content" gorm:"type:text;comment:文档内容"`
	UserName  string    `json:"user_name" gorm:"type:varchar(64);not null;comment:删除者名称"`
	DeletedAt time.Time `json:"deleted_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (PersonalRecycle) TableName() string {
	return middleware.TablePersonalRecycle
}
