package models

import (
	"time"
)

// ProjectGit 定义项目和Git仓库地址的对应关系表
type ProjectGit struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	ProjectID   uint      `json:"project_id" gorm:"not null;uniqueIndex;comment:项目ID"`
	GitURL      string    `json:"git_url" gorm:"type:varchar(255);not null;comment:Git仓库URL"`
	Description string    `json:"description" gorm:"type:text;comment:仓库描述"`
	CreatedBy   uint      `json:"created_by" gorm:"comment:创建人ID"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 设置表名
func (ProjectGit) TableName() string {
	return "project_git_repos"
}
