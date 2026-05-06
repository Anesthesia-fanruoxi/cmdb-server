package store

import (
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// AgentStoreListResponse 插件商店列表响应
type AgentStoreListResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	DisplayName string `json:"display_name"`
	PluginType  string `json:"plugin_type"`
	Description string `json:"description"`
	Port        int    `json:"port"`
	IsConfig    bool   `json:"is_config"` // 是否配置了模板
}

// AgentStoreList 获取插件商店列表
// 支持关键词搜索，按name和version排序
func AgentStoreList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	keyword := query.Get("keyword") // 关键词搜索

	// 构建查询（排除agent，agent作为载体不在商店中展示）
	db := database.DB.Table(middleware.TablePluginStore).
		Where("id != ?", 1)

	// 关键词搜索（搜索名称或显示名称）
	if keyword != "" {
		db = db.Where("name LIKE ? OR display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询列表
	var plugins []agent.AgentStore
	if err := db.Order("name ASC, version DESC").Find(&plugins).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询插件列表失败")
		return
	}

	// 转换为响应格式
	var response []AgentStoreListResponse
	for _, p := range plugins {
		response = append(response, AgentStoreListResponse{
			ID:          p.ID,
			Name:        p.Name,
			Version:     p.Version,
			DisplayName: p.DisplayName,
			PluginType:  p.PluginType,
			Description: p.Description,
			Port:        p.Port,
			IsConfig:    p.ConfigTemplate != "", // 根据模板是否为空判断
		})
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
