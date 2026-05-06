package path

const (
	// 监控查询接口
	MonitorMetricsProjectsPath = "/api/monitor/metrics/projects"    // 获取可访问的项目列表
	MonitorQueryPath           = "/api/monitor/metrics/query"       // 即时查询
	MonitorQueryRangePath      = "/api/monitor/metrics/query_range" // 范围查询
	MonitorMetricsListPath     = "/api/monitor/metrics/list"        // 硬件监控视图列表
	MonitorMetricsListSSEPath  = "/sse/monitor/metrics/list"        // SSE硬件监控视图列表

	// 监控视图接口
	MonitorViewProjectsPath = "/api/monitor/view/projects" // 获取可访问的项目列表
	MonitorViewCreatePath   = "/api/monitor/view/create"   // 创建视图
	MonitorViewDetailPath   = "/api/monitor/view/detail"   // 视图详情
	MonitorViewListPath     = "/api/monitor/view/list"     // 视图列表
	MonitorViewUpdatePath   = "/api/monitor/view/update"   // 更新视图
	MonitorViewDeletePath   = "/api/monitor/view/delete"   // 删除视图

	// 告警接口
	MonitorAlertProjectsPath = "/api/monitor/alert/projects" // 获取可访问的项目列表
	MonitorAlertListPath     = "/api/monitor/alert/list"     // 获取活跃告警
)
