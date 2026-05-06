package plugins

import (
	"cmdb/config"
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// PluginVersionInfo 插件版本信息
type PluginVersionInfo struct {
	Name        string `json:"name"`                   // 插件名称
	Version     string `json:"version"`                // 最新版本号
	Image       string `json:"image,omitempty"`        // 完整镜像地址（容器类插件）
	DownloadURL string `json:"download_url,omitempty"` // 完整下载地址（二进制插件）
	PluginType  string `json:"plugin_type"`            // 插件类型(container/binary)
	DisplayName string `json:"display_name"`           // 显示名称
}

// AgentInfo Agent信息
type AgentInfo struct {
	Version     string `json:"version"`      // Agent版本
	DownloadURL string `json:"download_url"` // Agent下载地址
}

// VersionsResponseData 版本查询响应数据
type VersionsResponseData struct {
	Plugins []PluginVersionInfo `json:"plugins"` // 业务插件版本列表
	Total   int                 `json:"total"`   // 业务插件总数
	Agent   *AgentInfo          `json:"agent"`   // Agent信息（可能为空）
}

// AgentPluginsVersions 获取所有插件的最新版本信息（供Agent调用，响应加密）
func AgentPluginsVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取加密密钥
	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	if taskSalt == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "加密密钥未配置")
		return
	}

	// 步骤1：查询业务插件（id != 1），获取每个插件的最新版本
	var businessPlugins []agent.AgentStore
	subQuery := database.DB.Table(middleware.TablePluginStore).
		Select("name, MAX(created_at) as max_created_at").
		Where("id != ?", 1).
		Group("name")

	err := database.DB.
		Table(middleware.TablePluginStore+" as p1").
		Joins("JOIN (?) as p2 ON p1.name = p2.name AND p1.created_at = p2.max_created_at", subQuery).
		Where("p1.id != ?", 1).
		Find(&businessPlugins).Error

	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询业务插件版本失败")
		return
	}

	// 获取Agent配置，用于生成完整地址
	agentConfig := config.GetAgentConfig()

	// 步骤2：查询Agent（id = 1）
	var agentPlugin agent.AgentStore
	var agentInfo *AgentInfo
	err = database.DB.
		Select("name, version").
		Where("id = ?", 1).
		First(&agentPlugin).Error

	if err == nil {
		// 找到Agent插件，构建Agent信息
		agentInfo = &AgentInfo{
			Version:     agentPlugin.Version,
			DownloadURL: agentPlugin.GetDownloadURL(agentConfig.DownloadUrl),
		}
	} else {
		// 没有找到Agent插件
		middleware.Logger.Warnf("未找到Agent插件 (id=1): %v", err)
		agentInfo = nil
	}

	// 转换为业务插件版本列表
	var pluginsInfo []PluginVersionInfo
	for _, plugin := range businessPlugins {
		info := PluginVersionInfo{
			Name:        plugin.Name,
			Version:     plugin.Version,
			PluginType:  plugin.PluginType,
			DisplayName: plugin.DisplayName,
		}

		// 根据插件类型添加完整地址
		if plugin.PluginType == "container" {
			info.Image = plugin.GetImageURL(agentConfig.ImageUrl)
		} else if plugin.PluginType == "binary" {
			info.DownloadURL = plugin.GetDownloadURL(agentConfig.DownloadUrl)
		}

		pluginsInfo = append(pluginsInfo, info)
	}

	// 构建响应数据
	responseData := VersionsResponseData{
		Plugins: pluginsInfo,
		Total:   len(pluginsInfo),
		Agent:   agentInfo,
	}

	// 序列化为JSON
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化响应失败")
		return
	}

	// 加密响应数据
	encryptedData, err := middleware.CompressAndEncrypt(responseJSON, taskSalt)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "加密响应失败")
		return
	}

	// 直接返回加密数据（不嵌套JSON）
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"code":200,"msg":"查询成功","data":"` + encryptedData + `"}`))
}
