package plugins

import (
	"cmdb/config"
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// AgentPluginsListResponse 插件列表响应
type AgentPluginsListResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	DisplayName string `json:"display_name"`
	PluginType  string `json:"plugin_type"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`        // 容器镜像地址（动态生成）
	DownloadURL string `json:"download_url,omitempty"` // 下载地址（动态生成）
	Port        int    `json:"port"`
	IsConfig    bool   `json:"is_config"` // 是否配置了模板
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// AgentPluginsList 获取插件管理列表（后台表格）
func AgentPluginsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	keyword := query.Get("keyword") // 关键词搜索
	pluginType := query.Get("type") // 插件类型过滤

	// 构建查询（显示所有插件，包括agent）
	db := database.DB.Table(middleware.TablePluginStore)

	// 关键词搜索
	if keyword != "" {
		db = db.Where("name LIKE ? OR display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型过滤
	if pluginType != "" {
		db = db.Where("plugin_type = ?", pluginType)
	}

	// 查询列表
	var plugins []agent.AgentStore
	if err := db.Order("created_at DESC").Find(&plugins).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询插件列表失败")
		return
	}

	// 获取Agent配置
	agentConfig := config.GetAgentConfig()
	middleware.Logger.Infof("Agent配置 - ImageUrl: %s, DownloadUrl: %s", agentConfig.ImageUrl, agentConfig.DownloadUrl)

	// 转换为响应格式
	var response []AgentPluginsListResponse
	for _, p := range plugins {
		item := AgentPluginsListResponse{
			ID:          p.ID,
			Name:        p.Name,
			Version:     p.Version,
			DisplayName: p.DisplayName,
			PluginType:  p.PluginType,
			Description: p.Description,
			Port:        p.Port,
			IsConfig:    p.ConfigTemplate != "", // 根据模板是否为空判断
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 根据类型添加对应的URL
		if p.PluginType == "container" {
			item.Image = p.GetImageURL(agentConfig.ImageUrl)
		} else if p.PluginType == "binary" {
			item.DownloadURL = p.GetDownloadURL(agentConfig.DownloadUrl)
		}

		response = append(response, item)
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
