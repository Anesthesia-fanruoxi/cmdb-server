package knowledge

import (
	"cmdb/pkg/middleware"
	"time"
)

// Public 公开文档
type Public struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null;comment:文档标题"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:文档内容"`
	Category  string    `json:"category" gorm:"type:varchar(50);not null;comment:文档分类"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;default:'';comment:所属项目"`
	ShareCode string    `json:"share_code" gorm:"type:varchar(32);default:'';comment:分享码"`
	Creator   string    `json:"creator" gorm:"type:varchar(50);not null;comment:创建者名称"`
	UserName  string    `json:"user_name" gorm:"type:varchar(50);not null;comment:最后修改者名称"`
	Version   int64     `json:"version" gorm:"type:bigint;not null;comment:版本号"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Public) TableName() string {
	return middleware.TablePublic
}

// PublicShare 公开文档分享
type PublicShare struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	DocID     int64     `json:"doc_id" gorm:"type:bigint unsigned;not null;comment:文档ID"`
	ShareCode string    `json:"share_code" gorm:"type:varchar(32);not null;uniqueIndex;comment:分享码"`
	ShareUrl  string    `json:"share_url" gorm:"type:varchar(255);not null;comment:分享链接"`
	Creator   string    `json:"creator" gorm:"type:varchar(50);not null;comment:分享者名称"`
	ExpiredAt time.Time `json:"expired_at" gorm:"type:datetime;not null;comment:过期时间"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (PublicShare) TableName() string {
	return middleware.TablePublicShare
}

// PublicHistory 公开文档历史记录
type PublicHistory struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	DocID     int64     `json:"doc_id" gorm:"type:bigint unsigned;not null;comment:文档ID"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null;comment:文档标题"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:文档内容"`
	Category  string    `json:"category" gorm:"type:varchar(50);not null;comment:文档分类"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;default:'';comment:所属项目"`
	Creator   string    `json:"creator" gorm:"type:text;comment:创建者名称"`
	Version   int64     `json:"version" gorm:"type:bigint;not null;comment:版本号"`
	UserName  string    `json:"user_name" gorm:"type:varchar(50);not null;comment:操作者名称"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (PublicHistory) TableName() string {
	return middleware.TablePublicHistory
}

// PublicRecycle 公开文档回收站
type PublicRecycle struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	DocID     int64     `json:"doc_id" gorm:"type:bigint unsigned;not null;comment:文档ID"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null;comment:文档标题"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:文档内容"`
	Category  string    `json:"category" gorm:"type:varchar(50);not null;comment:文档分类"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;default:'';comment:所属项目"`
	Creator   string    `json:"creator" gorm:"type:text;comment:创建者名称"`
	UserName  string    `json:"user_name" gorm:"type:varchar(50);not null;comment:删除者名称"`
	DeletedAt time.Time `json:"deleted_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (PublicRecycle) TableName() string {
	return middleware.TablePublicRecycle
}
