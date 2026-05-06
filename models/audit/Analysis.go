package audit

import (
	"cmdb/models/sql"
	"time"
)

// 通用的时间范围查询参数
type TimeRangeQuery struct {
	StartTime string `json:"start_time"` // 开始时间
	EndTime   string `json:"end_time"`   // 结束时间
}

// UserStatsQuery 用户统计查询参数(通用于SQL和ES)
type UserStatsQuery struct {
	TimeRangeQuery
	Hour      int    `json:"hour"`       // 小时：0-23，表示一天中的某个小时，-1表示所有小时
	AuditType string `json:"audit_type"` // 审计类型：sql或es
}

// UserStatItem 用户统计项
type UserStatItem struct {
	UserID       int64  `json:"user_id"`       // 用户ID
	UserName     string `json:"user_name"`     // 用户名
	NickName     string `json:"nick_name"`     // 昵称
	Count        int    `json:"count"`         // 总执行次数
	WebCount     int    `json:"web_count"`     // Web平台执行次数
	DesktopCount int    `json:"desktop_count"` // Desktop平台执行次数
}

// SQLAuditRequest SQL审计查询请求结构
type SQLAuditRequest struct {
	UserName string `json:"UserName"` // 修改为匹配前端的大写形式
	Project  string `json:"project"`
	DbName   string `json:"db_name"`
	QueryId  string `json:"query_id"`
	TimeRangeQuery
	Page int `json:"page"` // 页码
}

// SQLAuditListItem SQL审计查询结果列表项
type SQLAuditListItem struct {
	QueryID        string    `json:"query_id" gorm:"column:query_id"`
	NickName       string    `json:"nick_name" gorm:"column:nick_name"`
	DbName         string    `json:"db_name" gorm:"column:db_name"`
	ExecutionTime  int       `json:"execution_time" gorm:"column:execution_time"`
	AffectedRows   int       `json:"affected_rows" gorm:"column:affected_rows"`
	Status         string    `json:"status" gorm:"column:status"`
	Operation      string    `json:"operation" gorm:"column:operation"`
	OperationCount int       `json:"operation_count" gorm:"column:operation_count"`
	ClientIP       string    `json:"client_ip" gorm:"column:client_ip"`
	City           string    `json:"city" gorm:"column:city"`
	Country        string    `json:"country" gorm:"column:country"`
	Region         string    `json:"region" gorm:"column:region"`
	District       string    `json:"district" gorm:"column:district"`
	Platform       string    `json:"platform" gorm:"column:platform"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
}

// SQLAuditResponse SQL审计查询响应结构
type SQLAuditResponse struct {
	Total int64              `json:"total"` // 总记录数
	List  []SQLAuditListItem `json:"list"`  // 记录列表
}

// SQLAuditDetailRequest 审计详情请求结构
type SQLAuditDetailRequest struct {
	QueryID string `json:"query_id"` // 查询ID
}

// SQLOperation SQL操作记录
type SQLOperation struct {
	NickName      string    `json:"nick_name"`      // 昵称
	Operation     string    `json:"operation"`      // 操作
	Page          int       `json:"page"`           // 页码
	ExecutionTime int       `json:"execution_time"` // 执行时间(毫秒)
	AffectedRows  int       `json:"affected_rows"`  // 影响行数
	Status        string    `json:"status"`         // 执行状态
	CreatedAt     time.Time `json:"created_at"`     // 操作时间
}

// 直接使用sql包中的SqlAuditLog作为详情
type SQLDetail sql.SqlAuditLog

// SQLAuditDetailResponse SQL审计详情响应结构
type SQLAuditDetailResponse struct {
	SearchDetail   SQLDetail      `json:"search_detail"`   // 查询操作详情
	PageOperations []SQLOperation `json:"page_operations"` // 相关翻页操作列表
}

// HourlyAuditQuery 小时审计查询参数
type HourlyAuditQuery struct {
	TimeRangeQuery
	Hour int `json:"hour"` // 指定小时，-1表示查询所有小时
}

// UserListQuery 用户查询列表参数(通用于SQL和ES)
type UserListQuery struct {
	UserID         int64  `json:"user_id"`    // 用户ID
	Hour           int    `json:"hour"`       // 小时：0-23，表示一天中的某个小时
	AuditType      string `json:"audit_type"` // 审计类型：sql或es
	TimeRangeQuery        // 时间范围查询参数
}
