package task

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(32);comment:任务ID"`
	Name         string    `json:"name" gorm:"type:varchar(64);comment:项目名称"`
	Type         string    `json:"type" gorm:"type:varchar(32);default:'default';comment:任务类型"`
	Status       string    `json:"status" gorm:"type:varchar(16);default:'pending';comment:任务状态"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:创建时间"`
	StartedAt    time.Time `json:"started_at" gorm:"comment:开始时间"`
	CompletedAt  time.Time `json:"completed_at" gorm:"comment:完成时间"`
	GitURL       string    `json:"git_url" gorm:"type:varchar(255);comment:Git仓库地址"`
	Description  string    `json:"description" gorm:"type:text;comment:任务描述"`
	ProFeishuURL string    `json:"pro_feishu_url" gorm:"type:varchar(255);comment:项目飞书URL"`
	OpsFeishuURL string    `json:"ops_feishu_url" gorm:"type:varchar(255);comment:运维飞书URL"`
	Result       string    `json:"result" gorm:"type:text;comment:任务结果"`
	ErrorMessage string    `json:"error_message" gorm:"type:text;comment:错误信息"`
}

// TableName 设置表名
func (Task) TableName() string {
	return "t_tasks"
}
