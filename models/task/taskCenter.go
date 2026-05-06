package task

import (
	"cmdb/pkg/middleware"
	"time"
)

// TaskCenter 任务中心表模型
type TaskCenter struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID   string `json:"task_id" gorm:"type:varchar(64);not null;index;comment:任务ID"`
	Type     string `json:"type" gorm:"type:varchar(32);not null;index;comment:任务类型"`
	Status   string `json:"status" gorm:"type:varchar(16);not null;default:pending;comment:状态"`
	UserID   int64  `json:"user_id" gorm:"not null;index;comment:用户ID"`
	UserName string `json:"user_name" gorm:"type:varchar(64);comment:用户名"`
	NickName string `json:"nick_name" gorm:"-"` // 不映射到数据库字段，仅用于查询时接收数据
	Title    string `json:"title" gorm:"type:varchar(128);comment:任务标题"`

	// 文件相关
	FilePath string `json:"file_path" gorm:"type:varchar(512);comment:文件路径"`
	FileName string `json:"file_name" gorm:"type:varchar(128);comment:文件名"`
	FileSize int64  `json:"file_size" gorm:"default:0;comment:文件大小"`

	// 结果相关
	TotalCount   int    `json:"total_count" gorm:"default:0;comment:总记录数"`
	ErrorMessage string `json:"error_message" gorm:"type:text;comment:错误信息"`

	// 时间
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	CompletedAt *time.Time `json:"completed_at" gorm:"comment:完成时间"`
	ExpireAt    *time.Time `json:"expire_at" gorm:"comment:过期时间"`
}

// TableName 设置表名
func (TaskCenter) TableName() string {
	return middleware.TableTaskCenter
}

// 任务类型常量
const (
	TypeAnalysis  = "analysis"   // ES数据分析
	TypeESExport  = "es_export"  // ES日志导出
	TypeSQLExport = "sql_export" // SQL查询导出
)

// 任务状态常量
const (
	StatusPending  = "pending"  // 等待中
	StatusRunning  = "running"  // 执行中
	StatusSuccess  = "success"  // 成功
	StatusFailed   = "failed"   // 失败
	StatusCanceled = "canceled" // 已取消
)

// GetTypeText 获取类型文本
func GetTypeText(t string) string {
	switch t {
	case TypeAnalysis:
		return "数据分析"
	case TypeESExport:
		return "日志导出"
	case TypeSQLExport:
		return "SQL导出"
	default:
		return t
	}
}
