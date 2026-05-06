package sql

import (
	"cmdb/pkg/middleware"
	"time"
)

// SqlAuditLog SQL审计日志模型
type SqlAuditLog struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        int64     `json:"user_id" gorm:"column:user_id;not null;comment:用户ID"`
	UserName      string    `json:"user_name" gorm:"column:user_name;not null;comment:用户名"`
	Project       string    `json:"project" gorm:"column:project;not null;comment:项目名称"`
	ClientIP      string    `json:"client_ip" gorm:"column:client_ip;not null;comment:客户端IP"`
	City          string    `json:"city" gorm:"column:city;not null;comment:城市"`
	Country       string    `json:"country" gorm:"column:country;not null;comment:国家"`
	Region        string    `json:"region" gorm:"column:region;not null;comment:省份"`
	District      string    `json:"district" gorm:"column:district;not null;comment:区县"`
	QueryID       string    `json:"query_id" gorm:"column:query_id;comment:查询ID"`
	Operation     string    `gorm:"column:operation;type:varchar(20);comment:操作类型"`
	Page          int       `gorm:"column:page;comment:页码"`
	Platform      string    `gorm:"column:platform;type:varchar(20);comment:平台标识:desktop/web"`
	DBName        string    `json:"db_name" gorm:"column:db_name;not null;comment:数据库名称"`
	QuerySql      string    `json:"query_sql" gorm:"column:query_sql;type:text;not null;comment:查询语句"`
	AffectedRows  int       `json:"affected_rows" gorm:"column:affected_rows;default:0;comment:影响行数"`
	ExecutionTime int       `json:"execution_time" gorm:"column:execution_time;default:0;comment:执行时间(毫秒)"`
	Status        string    `json:"status" gorm:"column:status;default:success;comment:执行状态: success/failed"`
	ErrorMessage  string    `json:"error_message" gorm:"column:error_message;type:text;comment:错误信息"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (SqlAuditLog) TableName() string {
	return middleware.TableSqlAuditLog
}

type SqlRule struct {
	ID              int64  `json:"id" gorm:"column:id;primaryKey"`
	RuleName        string `json:"rule_name" gorm:"column:rule_name"`
	RuleCategory    string `json:"rule_category" gorm:"column:rule_category"`
	RuleType        string `json:"rule_type" gorm:"column:rule_type"`
	SqlTag          string `json:"sql_tag" gorm:"column:sql_tag"`
	RuleDescription string `json:"rule_description" gorm:"column:rule_description"`
	IsEnabled       int    `json:"is_enabled" gorm:"column:is_enabled"`
	RuleContent     string `json:"rule_content" gorm:"column:rule_content;type:text"`
	ErrorMessage    string `json:"error_message" gorm:"column:error_message;type:text"`
}

func (SqlRule) TableName() string {
	return middleware.TableSqlRule
}

type SQLApplyFlow struct {
	ID         int64  `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	ProjectId  *int64 `json:"project_id" gorm:"column:project_id"`   // 项目ID
	ApplyID    *int64 `json:"apply_id" gorm:"column:apply_id"`       // 审批人ID
	ExecutorID *int64 `json:"executor_id" gorm:"column:executor_id"` // 执行人ID
}

// TableName 设置表名
func (SQLApplyFlow) TableName() string {
	return middleware.TableSQLApplyFlow
}

// SQLApply 表示SQL审批表
type SQLApply struct {
	ID                  int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Project             string     `json:"project" gorm:"column:project;not null"`
	SubmitterID         int64      `json:"submitter_id" gorm:"column:submitter_id;not null"`
	SubmitterName       string     `json:"submitter_name" gorm:"column:submitter_name;not null"`
	SubmitterStatus     int        `json:"submitter_status" gorm:"column:submitter_status;default:0"` // 0-不能处理，1-可以处理
	SQLContent          string     `json:"sql_content" gorm:"column:sql_content;type:text;not null"`
	DatabaseName        string     `json:"database_name" gorm:"column:database_name;not null"`
	ExecutionTime       *time.Time `json:"execution_time" gorm:"column:execution_time"`
	Description         string     `json:"description" gorm:"column:description;type:text"`
	RuleCheckResult     string     `json:"rule_check_result" gorm:"column:rule_check_result;type:text"`
	HasViolation        int        `json:"has_violation" gorm:"column:has_violation;default:0"`
	HasBlocker          int        `json:"has_blocker" gorm:"column:has_blocker;default:0"`
	ApplyStatus         int        `json:"apply_status" gorm:"column:apply_status;default:0"` // 0-不能处理，1-可以处理
	ApplyID             int64      `json:"apply_id" gorm:"column:apply_id"`
	ApplyName           string     `json:"apply_name" gorm:"column:apply_name"`
	ApplyComment        string     `json:"apply_comment" gorm:"column:apply_comment;type:text"`
	ApplyTime           *time.Time `json:"apply_time" gorm:"column:apply_time"`
	ApplySource         string     `json:"apply_source" gorm:"column:apply_source"`                 // 审批人来源：web/feishu
	ExecutorStatus      int        `json:"executor_status" gorm:"column:executor_status;default:0"` // 0-不能处理，1-可以处理
	ExecutorID          int64      `json:"executor_id" gorm:"column:executor_id"`
	ExecutorName        string     `json:"executor_name" gorm:"column:executor_name"`
	ExecutorSource      string     `json:"executor_source" gorm:"column:executor_source"` // 执行人来源：web/feishu
	ExecutionResult     string     `json:"execution_result" gorm:"column:execution_result;type:text"`
	ActualExecutionTime *time.Time `json:"actual_execution_time" gorm:"column:actual_execution_time"`
	CreatedAt           time.Time  `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"column:updated_at;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	Status              int        `json:"status" gorm:"column:status;default:0"` // 0-待审批，1-待执行，2-执行中，3-执行完成，4-执行失败，5-已驳回，6-已撤销
}

// TableName 指定表名
func (SQLApply) TableName() string {
	return middleware.TableSQLApply
}

// 简化的规则信息
type RuleInfo struct {
	RuleName     string `json:"rule_name"`     // 规则名称
	RuleLevel    int    `json:"rule_level"`    // 规则级别
	ErrorMessage string `json:"error_message"` // 错误信息
	Passed       bool   `json:"passed"`        // 是否通过规则检查
}

// 简化的SQL检查结果，用于序列化为JSON存储
type SimpleSQLCheckResult struct {
	SQLIndex     int        `json:"sql_index"`     // SQL在审批中的索引
	SQLContent   string     `json:"sql_content"`   // SQL内容
	TableNames   []string   `json:"table_names"`   // 表名
	HasViolation bool       `json:"has_violation"` // 是否有违规
	HasBlocker   bool       `json:"has_blocker"`   // 是否有阻断
	RuleInfos    []RuleInfo `json:"rule_infos"`    // 违反的规则信息
}

// QueryMainInfo 查询主信息
type QueryMainInfo struct {
	// 查询总记录数
	TotalCount int64 `json:"total_count"`

	// 每页大小（固定为20）
	PageSize int `json:"page_size"`

	// 总页数
	TotalPages int `json:"total_pages"`

	// 子qid数量（批次总数）
	BatchCount int `json:"batch_count"`

	// 每批次大小（固定为1000）
	BatchSize int `json:"batch_size"`

	// 查询语句
	Query string `json:"query"`

	// 数据库名称
	DBName string `json:"db_name"`

	// 项目/代理名称
	Agent string `json:"agent"`

	// 列名
	Columns []string `json:"columns"`

	// 缓存策略：full(全部缓存), partial(部分缓存), minimal(最小缓存)
	CacheStrategy string `json:"cache_strategy,omitempty"`
}

// BatchInfo 批次信息
type BatchInfo struct {
	// 批次索引（从1开始）
	BatchIndex int `json:"batch_index"`

	// 该批次的记录总数
	Count int `json:"count"`

	// 主键ID
	MainID string `json:"main_id"`

	// 查询时间
	QueryTime time.Time `json:"query_time"`
}
