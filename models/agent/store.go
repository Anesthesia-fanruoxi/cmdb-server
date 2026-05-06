package agent

import (
	"cmdb/pkg/middleware"
	"fmt"
	"time"
)

// AgentStore 插件商店模型
type AgentStore struct {
	ID             int64     `json:"id" gorm:"primarykey;comment:主键ID"`
	Name           string    `json:"name" gorm:"size:64;not null;comment:插件名称"`
	Version        string    `json:"version" gorm:"size:32;not null;comment:版本号"`
	DisplayName    string    `json:"display_name" gorm:"size:128;not null;comment:显示名称"`
	PluginType     string    `json:"plugin_type" gorm:"size:32;not null;comment:插件类型(container/binary)"`
	Description    string    `json:"description" gorm:"type:text;comment:插件描述"`
	Port           int       `json:"port" gorm:"comment:插件服务端口(容器内端口,仅container类型)"`
	ConfigTemplate string    `json:"config_template" gorm:"type:text;comment:配置文件模板(支持Go template语法)"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

// GetImageURL 获取镜像地址（动态拼接）
// 格式: registry/name:version
// 例如: hub.hzbxhd.com/test/sql-plugs:2.0
func (s *AgentStore) GetImageURL(registry string) string {
	if s.PluginType == "container" {
		return fmt.Sprintf("%s/%s:%s", registry, s.Name, s.Version)
	}
	return ""
}

// GetDownloadURL 获取下载地址（动态拼接）
// 格式: baseURL/name （不包含版本号）
// 例如: http://192.168.3.12:8080/plugs/mysql-client
// 注意：所有类型都可能需要下载地址（容器类型用于配置文件等）
func (s *AgentStore) GetDownloadURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", baseURL, s.Name)
}

// TableName 设置表名
func (AgentStore) TableName() string {
	return middleware.TablePluginStore
}
