package elfk

import (
	"cmdb/pkg/middleware"
	"time"
)

// SharedKeyword 共享查询关键词表
type SharedKeyword struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;index;comment:项目"`
	Category  string    `json:"category" gorm:"type:varchar(64);not null;comment:分类"`
	ViewID    int64     `json:"view_id" gorm:"not null;index;comment:视图ID"`
	Keyword   string    `json:"keyword" gorm:"type:varchar(500);not null;index:idx_unique_keyword;comment:关键词"`
	Remark    string    `json:"remark" gorm:"type:varchar(255);comment:备注"`
	Creator   string    `json:"creator" gorm:"type:varchar(64);not null;comment:添加人"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
}

// TableIndex 定义复合唯一索引（project + category + view_id + keyword）
func (SharedKeyword) TableIndex() []string {
	return []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_shared_keyword_unique ON es_shared_keyword(project, category, view_id, keyword)"}
}

// TableName 指定表名
func (SharedKeyword) TableName() string {
	return middleware.TableSharedKeyword
}
