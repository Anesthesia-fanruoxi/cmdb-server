package path

// Agent插件管理路径
const (
	// 插件管理（后台CRUD表格）
	AgentPluginsListPath          = "/api/agent/plugins/list"           // 插件管理列表
	AgentPluginsDetailPath        = "/api/agent/plugins/detail"         // 插件详情
	AgentPluginsCreatePath        = "/api/agent/plugins/create"         // 创建插件
	AgentPluginsUpdatePath        = "/api/agent/plugins/update"         // 更新插件
	AgentPluginsDeletePath        = "/api/agent/plugins/delete"         // 删除插件
	AgentPluginsConfigPreviewPath = "/api/agent/plugins/config/preview" // 配置预览
	AgentPluginsVersionsPath      = "/api/agent/plugins/versions"       // 获取所有插件最新版本（供Agent调用）

	// 插件商店（用户展示和安装）
	AgentStoreProjectsPath = "/api/agent/store/projects" // 获取可访问的项目列表
	AgentStoreListPath     = "/api/agent/store/list"     // 插件商店列表展示
	AgentStoreInstallPath  = "/api/agent/store/install"  // 从商店安装插件到项目

	// 项目插件管理
	AgentProjectListPath       = "/api/agent/project/list"        // 查询已配置Agent的项目列表
	AgentProjectListDetailPath = "/api/agent/project/list/detail" // 查询项目已安装的插件列表
	AgentProjectControlPath    = "/api/agent/project/control"     // 插件控制（启动/停止/重启/卸载）
	AgentProjectUpgradePath    = "/api/agent/project/upgrade"     // 插件版本升级
	AgentProjectConfigPath     = "/api/agent/project/config"      // 插件配置更新
	AgentProjectLogsPath       = "/ws/agent/project/logs"         // WebSocket实时查看插件日志
)
