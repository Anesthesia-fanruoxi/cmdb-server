package elfk

import (
	"cmdb/pkg/middleware"
	"time"
)

// SysEsLog ES审计日志表结构
type SysEsLog struct {
	ID            int64     `json:"id" gorm:"primary_key;auto_increment;comment:'主键ID'"`
	QueryID       string    `json:"query_id" gorm:"type:varchar(64);comment:'查询ID'"`
	UserID        int64     `json:"user_id" gorm:"type:varchar(64);comment:'用户ID'"`
	UserName      string    `json:"user_name" gorm:"type:varchar(100);comment:'用户名'"`
	ClientIP      string    `json:"client_ip" gorm:"type:varchar(64);comment:'客户端IP'"`
	Country       string    `json:"country" gorm:"type:varchar(255);comment:'国家'"`
	City          string    `json:"city" gorm:"type:varchar(255);comment:'城市'"`
	Region        string    `json:"region" gorm:"type:varchar(255);comment:'地区'"`
	District      string    `json:"district" gorm:"type:varchar(255);comment:'区县'"`
	Project       string    `json:"project" gorm:"type:varchar(255);comment:'项目名称'"`
	ViewName      string    `json:"view_name" gorm:"type:varchar(255);comment:'视图名称'"`
	Operation     string    `json:"operation" gorm:"type:varchar(255);comment:'操作类型'"`
	Page          int       `json:"page" gorm:"comment:'页码'"`
	QIndexPattern string    `json:"q_index_pattern" gorm:"type:varchar(255);comment:'查询索引模式'"`
	QTimeField    string    `json:"q_time_field" gorm:"type:varchar(255);comment:'查询时间字段'"`
	QStartTime    string    `json:"q_start_time" gorm:"type:varchar(255);comment:'查询开始时间'"`
	QEndTime      string    `json:"q_end_time" gorm:"type:varchar(255);comment:'查询结束时间'"`
	QKeyword      string    `json:"q_keyword" gorm:"type:varchar(255);comment:'搜索关键词'"`
	ResponseCode  int       `json:"response_code" gorm:"comment:'响应状态码'"`
	QueryTimeMs   int64     `json:"query_time_ms" gorm:"comment:'查询耗时(毫秒)'"`
	DocCount      int       `json:"doc_count" gorm:"comment:'查询文档数量'"`
	ErrorMessage  string    `json:"error_message" gorm:"type:text;comment:'错误信息'"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP;comment:'创建时间'"`
}

// TableName 指定表名
func (SysEsLog) TableName() string {
	return middleware.TableEsAuditLog
}
