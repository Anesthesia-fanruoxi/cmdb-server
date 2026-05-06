package system

import "time"

// AgentDict agent字典
type AgentDict struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	Project   string    `json:"project"`    // 项目标识
	AgentURL  string    `json:"agent_url"`  // agent地址
	FeishuUrl string    `json:"feishu_url"` // 飞书告警地址
	Color     string    `json:"color"`      // 字典颜色
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (AgentDict) TableName() string {
	return "sys_agent_dict"
}
