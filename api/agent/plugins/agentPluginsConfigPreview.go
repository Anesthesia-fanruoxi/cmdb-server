package plugins

import (
	"bytes"
	"cmdb/models/agent"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"text/template"
)

// ConfigPreviewRequest 配置预览请求
type ConfigPreviewRequest struct {
	PluginID       int64                  `json:"plugin_id"`                         // 插件ID（二选一）
	ConfigTemplate string                 `json:"config_template"`                   // 配置模板（二选一）
	ConfigParams   map[string]interface{} `json:"config_params" validate:"required"` // 配置参数
	CustomConfig   string                 `json:"custom_config"`                     // 自定义额外配置（可选）
}

// ConfigPreviewResponse 配置预览响应
type ConfigPreviewResponse struct {
	OriginalTemplate string                 `json:"original_template"` // 原始模板
	RenderedConfig   string                 `json:"rendered_config"`   // 渲染后的配置
	ConfigParams     map[string]interface{} `json:"config_params"`     // 使用的参数
}

// AgentPluginsConfigPreview 预览插件配置（渲染模板）
func AgentPluginsConfigPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req ConfigPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.ConfigParams == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "config_params不能为空")
		return
	}

	var configTemplate string

	// 获取配置模板
	if req.PluginID > 0 {
		// 从数据库查询插件配置模板
		var plugin agent.AgentStore
		if err := database.DB.First(&plugin, req.PluginID).Error; err != nil {
			middleware.ResponseError(w, http.StatusNotFound, "插件不存在")
			return
		}
		if plugin.ConfigTemplate == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "该插件未配置模板")
			return
		}
		configTemplate = plugin.ConfigTemplate
	} else if req.ConfigTemplate != "" {
		// 使用传入的配置模板
		configTemplate = req.ConfigTemplate
	} else {
		middleware.ResponseError(w, http.StatusBadRequest, "plugin_id和config_template至少提供一个")
		return
	}

	// 如果有自定义额外配置，追加到模板后面
	finalTemplate := configTemplate
	if req.CustomConfig != "" {
		finalTemplate = configTemplate + "\n" + req.CustomConfig
		middleware.Logger.Infof("追加自定义配置 [长度: %d]", len(req.CustomConfig))
	}

	// 渲染配置模板
	tmpl, err := template.New("config").Parse(finalTemplate)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "模板解析失败: "+err.Error())
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.ConfigParams); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "模板渲染失败: "+err.Error())
		return
	}

	// 构建响应
	response := ConfigPreviewResponse{
		OriginalTemplate: configTemplate, // 返回原始模板（不含额外配置）
		RenderedConfig:   buf.String(),   // 返回完整渲染结果（含额外配置）
		ConfigParams:     req.ConfigParams,
	}

	middleware.ResponseSuccess(w, "预览成功", response)
}
