package knowledge

import (
	"cmdb/pkg/middleware"
	"time"
)

// Document 知识库文档
type Document struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null;comment:文档标题"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:文档内容"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;comment:关联的项目代码"`
	Category  string    `json:"category" gorm:"type:varchar(64);not null;comment:关联的分类"`
	Creator   string    `json:"creator" gorm:"type:varchar(64);not null;comment:创建者名称"`
	UserName  string    `json:"user_name" gorm:"type:varchar(64);not null;comment:最后更新者名称"`
	Version   int64     `json:"version" gorm:"type:int;not null;default:1;comment:当前版本号"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Document) TableName() string {
	return middleware.TableDocuments
}

// DocumentHistory 文档历史版本
type DocumentHistory struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	DocID     int64     `json:"doc_id" gorm:"type:bigint;not null;comment:关联的文档ID"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null;comment:历史版本标题"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:历史版本内容"`
	Category  string    `json:"category" gorm:"type:varchar(64);not null;comment:关联的分类"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;comment:所属项目代码"`
	Version   int64     `json:"version" gorm:"type:int;not null;comment:版本号"`
	Creator   string    `json:"creator" gorm:"type:varchar(64);not null;comment:创建者名称"`
	UserName  string    `json:"user_name" gorm:"type:varchar(64);not null;comment:更新者名称"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (DocumentHistory) TableName() string {
	return middleware.TableHistories
}

// DocumentRecycle 文档回收站
type DocumentRecycle struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	DocID       int64     `json:"doc_id" gorm:"type:bigint;not null;comment:文档ID"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null;comment:文档标题"`
	Content     string    `json:"content" gorm:"type:text;not null;comment:文档内容"`
	Category    string    `json:"category" gorm:"type:varchar(64);not null;comment:关联的分类"`
	Project     string    `json:"project" gorm:"type:varchar(64);not null;comment:所属项目代码"`
	Creator     string    `json:"creator" gorm:"type:varchar(64);not null;comment:创建者名称"`
	DeleterName string    `json:"deleter_name" gorm:"type:varchar(64);not null;comment:删除者名称"`
	DeletedAt   time.Time `json:"deleted_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (DocumentRecycle) TableName() string {
	return middleware.TableRecycle
}
