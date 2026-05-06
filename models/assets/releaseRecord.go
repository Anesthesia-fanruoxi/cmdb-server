package assets

import (
	"time"
)

// ReleaseRecord 发版记录模型
type ReleaseRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	TaskID       string    `json:"task_id" gorm:"type:varchar(32);uniqueIndex;comment:任务ID"`
	Project      string    `json:"project" gorm:"type:varchar(64);index;comment:项目名称"`
	Type         string    `json:"type" gorm:"type:varchar(32);default:'';index;comment:发版类型(web/double/single/空)"`
	Category     string    `json:"category" gorm:"type:varchar(32);default:'';comment:分类标签(如manager)"`
	Status       string    `json:"status" gorm:"type:varchar(16);default:'pending';comment:任务状态"`
	Step         int       `json:"step" gorm:"default:0;comment:发版步骤"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	StartedAt    time.Time `json:"started_at" gorm:"comment:开始时间"`
	CompletedAt  time.Time `json:"completed_at" gorm:"comment:完成时间;default:null"`
	GitURL       string    `json:"git_url" gorm:"type:varchar(255);comment:Git仓库地址"`
	Description  string    `json:"description" gorm:"type:text;comment:任务描述"`
	UpdateFeishu string    `json:"update_feishu" gorm:"type:varchar(255);comment:任务通知飞书URL"`
	NotifyFeishu string    `json:"notify_feishu" gorm:"type:varchar(255);comment:步骤通知飞书URL"`
	Result       string    `json:"result" gorm:"type:text;comment:任务结果"`
	ErrorMessage string    `json:"error_message" gorm:"type:text;comment:错误信息"`
}

// TableName 设置表名
func (ReleaseRecord) TableName() string {
	return "asset_release_record"
}
