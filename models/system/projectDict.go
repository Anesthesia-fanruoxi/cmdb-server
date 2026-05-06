package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// ProjectDict 项目字典表
type ProjectDict struct {
	ID               int64     `json:"id" gorm:"primarykey"`
	Project          string    `json:"project" gorm:"size:64;uniqueIndex;comment:项目标识"`
	ProjectName      string    `json:"project_name" gorm:"size:255;comment:项目名称"`
	AgentURL         string    `json:"agent_url" gorm:"size:255;default:'';comment:Agent代理地址"`
	BackenDomain     string    `json:"backen_domain" gorm:"size:255;default:'';comment:后台管理地址"`
	ApiDomain        string    `json:"api_domain" gorm:"size:255;default:'';comment:三方调用地址"`
	GitVue           string    `json:"git_vue" gorm:"size:255;default:'';comment:前端Git仓库"`
	GitBackend       string    `json:"git_backend" gorm:"size:255;default:'';comment:后端Git仓库"`
	AlterFeishu      string    `json:"alter_feishu" gorm:"size:255;default:'';comment:飞书告警地址"`
	UpdateFeishu     string    `json:"update_feishu" gorm:"size:255;default:'';comment:发版通知地址(开发组内)"`
	NotifyFeishu     string    `json:"notify_feishu" gorm:"size:255;default:'';comment:通知地址(所有人)"`
	Description      string    `json:"description" gorm:"size:500;default:'';comment:项目描述"`
	BackendTool      string    `json:"backend_tool" gorm:"size:255;default:'';comment:后端工具（java17等）"`
	FrontendTool     string    `json:"frontend_tool" gorm:"size:255;default:'';comment:前端工具（node14等）"`
	EnableSkywalking bool      `json:"enable_skywalking" gorm:"default:false;comment:是否启用skywalking"`
	Color            string    `json:"color" gorm:"size:20;default:'';comment:字典颜色"`
	Logo             string    `json:"logo" gorm:"size:500;default:'';comment:项目Logo地址"`
	CreatedBy        int64     `json:"created_by" gorm:"comment:创建人ID"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ProjectDict) TableName() string {
	return middleware.TableProjectDict
}
