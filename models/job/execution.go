package job

import (
	"cmdb/pkg/middleware"
	"time"
)

// JobExecution 任务执行记录表
type JobExecution struct {
	ID         int64      `json:"id" gorm:"primary_key"`
	JobID      int64      `json:"job_id" gorm:"not null"`                // 任务ID
	Project    string     `json:"project" gorm:"not null"`               // 项目标识
	TaskKey    string     `json:"task_key" gorm:"not null"`              // 任务标识符
	TaskName   string     `json:"task_name" gorm:"not null"`             // 任务名称
	ExecLog    string     `json:"exec_log"`                              // 执行日志（可为空）
	ExecStatus int        `json:"exec_status" gorm:"not null;default:0"` // 执行状态 0:执行中 1:成功 2:失败
	Result     string     `json:"result"`                                // 执行结果（可为空）
	ErrorMsg   string     `json:"error_msg"`                             // 错误信息（可为空）
	StartTime  time.Time  `json:"start_time" gorm:"not null"`            // 开始时间
	EndTime    *time.Time `json:"end_time"`                              // 结束时间（可为空）
	Duration   int        `json:"duration" gorm:"default:0"`             // 执行时长(秒)
	CreatedAt  time.Time  `json:"created_at" gorm:"not null"`            // 创建时间
}

// TableName 指定表名
func (JobExecution) TableName() string {
	return middleware.TableJobExecution
}
