package sql

import (
	"cmdb/pkg/middleware"
	"time"
)

// SharedQuery SQL共享查询表
type SharedQuery struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Project   string    `json:"project" gorm:"type:varchar(64);not null;index;comment:项目"`
	DBName    string    `json:"db_name" gorm:"type:varchar(64);not null;comment:数据库"`
	Query     string    `json:"query" gorm:"type:text;not null;comment:查询语句"`
	Remark    string    `json:"remark" gorm:"type:varchar(255);comment:备注"`
	Creator   string    `json:"creator" gorm:"type:varchar(64);not null;comment:添加人"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
}

// TableName 指定表名
func (SharedQuery) TableName() string {
	return middleware.TableSqlSharedQuery
}
