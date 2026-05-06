package sql

import (
	"cmdb/pkg/middleware"
	"time"
)

// Apply流程状态常量
const (
	ApplyStatusPending   = 0 // 待审批
	ApplyStatusApproved  = 1 // 待执行
	ApplyStatusExecuting = 2 // 执行中
	ApplyStatusCompleted = 3 // 执行成功
	ApplyStatusFailed    = 4 // 执行失败
	ApplyStatusRejected  = 5 // 已驳回
	ApplyStatusCanceled  = 6 // 已撤销
)

// 角色状态常量
const (
	ApplyRoleCannotProcess = 0 // 不能处理
	ApplyRoleCanProcess    = 1 // 可以处理
)

// 状态映射
var ApplyStatusTextMap = map[int]string{
	ApplyStatusPending:   "待审批",
	ApplyStatusApproved:  "待执行",
	ApplyStatusExecuting: "执行中",
	ApplyStatusCompleted: "执行成功",
	ApplyStatusFailed:    "执行失败",
	ApplyStatusRejected:  "已驳回",
	ApplyStatusCanceled:  "已撤销",
}

// UpdateApplyRequest 更新审批流程请求
type UpdateApplyRequest struct {
	ID          int64  `json:"id"`           // 流程ID
	ProcessType int    `json:"process_type"` // 处理类型：1通过，0拒绝
	Comment     string `json:"comment"`      // 备注信息
}

// ApplyDetailResponse 审批详情响应
type ApplyDetailResponse struct {
	ID                  int64  `json:"id"`
	Project             string `json:"project"`
	SubmitterID         int64  `json:"submitter_id"`
	SubmitterName       string `json:"submitter_name"`
	SubmitterStatus     int    `json:"submitter_status"`
	SQLContent          string `json:"sql_content"`
	DatabaseName        string `json:"database_name"`
	ExecutionTime       string `json:"execution_time"`
	Description         string `json:"description"`
	RuleCheckResult     string `json:"rule_check_result"`
	HasViolation        bool   `json:"has_violation"`
	HasBlocker          bool   `json:"has_blocker"`
	ApplyID             int64  `json:"apply_id"`
	ApplyName           string `json:"apply_name"`
	ApplyStatus         int    `json:"apply_status"`
	ApplyComment        string `json:"apply_comment"`
	ApplyTime           string `json:"apply_time"`
	ApplySource         string `json:"apply_source"`
	ExecutorID          int64  `json:"executor_id"`
	ExecutorName        string `json:"executor_name"`
	ExecutorStatus      int    `json:"executor_status"`
	ExecutorSource      string `json:"executor_source"`
	ExecutionResult     string `json:"execution_result"`
	ActualExecutionTime string `json:"actual_execution_time"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Status              int    `json:"status"`
	StatusText          string `json:"status_text"`
	CurrentOperator     string `json:"current_operator"`
}

// GetApplyStatusText 获取状态的中文描述
func GetApplyStatusText(status int) string {
	text, ok := ApplyStatusTextMap[status]
	if !ok {
		return "未知状态"
	}
	return text
}

// CalculateApplyCurrentOperator 计算当前操作人
func CalculateApplyCurrentOperator(apply *SQLApply) string {
	switch apply.Status {
	case ApplyStatusPending: // 待审批
		return apply.ApplyName
	case ApplyStatusApproved: // 待执行
		return apply.ExecutorName
	case ApplyStatusExecuting: // 执行中
		return apply.ExecutorName
	}
	return "" // 其他状态没有当前操作人
}

// formatTimePtr 格式化时间指针
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return middleware.FormatTimeToChina(*t)
}
