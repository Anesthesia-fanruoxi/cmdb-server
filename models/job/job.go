package job

import (
	"cmdb/pkg/middleware"
	"time"
)

// JobInfo 任务定义表
type JobInfo struct {
	ID            int64     `json:"id" gorm:"primary_key"`
	Name          string    `json:"name" gorm:"not null"`            // 任务名称
	TaskKey       string    `json:"task_key" gorm:"unique"`          // 任务标识符
	Cron          string    `json:"notice" gorm:"not null"`          // cron表达式
	Description   string    `json:"description"`                     // 任务描述
	ScriptID      string    `json:"script_id" gorm:"not null"`       // 脚本ID（Agent端脚本标识）
	ScriptType    string    `json:"script_type"`                     // 脚本类型（仅用于展示）
	ScriptContent string    `json:"script_content" gorm:"type:text"` // 脚本内容（仅用于展示和编辑）
	Status        int       `json:"status" gorm:"default:0"`         // 状态 0:停止 1:运行
	Creator       string    `json:"creator" gorm:"not null"`         // 创建人
	CreatedAt     time.Time `json:"created_at" gorm:"not null"`      // 创建时间
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null"`      // 更新时间
}

// TableName 指定表名
func (JobInfo) TableName() string {
	return middleware.TableJobInfo
}
