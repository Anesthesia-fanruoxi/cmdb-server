package routers

import (
	"cmdb/api/agent/plugins"
	"cmdb/api/agent/project"
	"cmdb/api/agent/store"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitAgentRoutes 初始化Agent插件管理路由
func InitAgentRoutes(mux *http.ServeMux) {
	// 插件管理（后台CRUD表格）
	mux.HandleFunc(path.AgentPluginsListPath, utils.JWTAuth(plugins.AgentPluginsList, "agent:plugins:r", "", "插件管理列表"))
	mux.HandleFunc(path.AgentPluginsDetailPath, utils.JWTAuth(plugins.AgentPluginsDetail, "agent:plugins:r", "", "插件详情"))
	mux.HandleFunc(path.AgentPluginsCreatePath, utils.JWTAuth(plugins.AgentPluginsCreate, "agent:plugins:w", "", "创建插件"))
	mux.HandleFunc(path.AgentPluginsUpdatePath, utils.JWTAuth(plugins.AgentPluginsUpdate, "agent:plugins:w", "", "更新插件"))
	mux.HandleFunc(path.AgentPluginsDeletePath, utils.JWTAuth(plugins.AgentPluginsDelete, "agent:plugins:w", "", "删除插件"))
	mux.HandleFunc(path.AgentPluginsConfigPreviewPath, utils.JWTAuth(plugins.AgentPluginsConfigPreview, "agent:plugins:r", "", "配置预览"))
	mux.HandleFunc(path.AgentPluginsVersionsPath, plugins.AgentPluginsVersions) // 公开接口，供Agent调用，无需JWT认证

	// 插件商店（用户展示和安装）
	mux.HandleFunc(path.AgentStoreProjectsPath, utils.JWTAuth(store.StoreProjects, "agent:store:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.AgentStoreListPath, utils.JWTAuth(store.AgentStoreList, "agent:store:r", "", "插件商店列表"))
	mux.HandleFunc(path.AgentStoreInstallPath, utils.JWTAuth(store.AgentStoreInstall, "agent:store:w", "o", "安装插件"))

	// 项目插件管理
	mux.HandleFunc(path.AgentProjectListPath, utils.JWTAuth(project.AgentProjectList, "agent:project:r", "", "项目列表"))
	mux.HandleFunc(path.AgentProjectListDetailPath, utils.JWTAuth(project.AgentProjectListDetail, "agent:project:r", "", "项目插件列表"))
	mux.HandleFunc(path.AgentProjectControlPath, utils.JWTAuth(project.AgentProjectControl, "agent:project:w", "o", "插件控制（启停重启卸载）"))
	mux.HandleFunc(path.AgentProjectUpgradePath, utils.JWTAuth(project.AgentProjectUpgrade, "agent:project:w", "o", "插件版本升级"))
	mux.HandleFunc(path.AgentProjectConfigPath, utils.JWTAuth(project.AgentProjectConfig, "agent:project:w", "o", "插件配置更新"))
	// WebSocket路由不使用JWT中间件，因为已在handler内部处理认证
	mux.HandleFunc(path.AgentProjectLogsPath, project.AgentProjectLogs)
}
