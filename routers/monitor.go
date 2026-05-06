package routers

import (
	"cmdb/api/monitor/alert"
	"cmdb/api/monitor/metrics"
	"cmdb/api/monitor/view"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitMonitorRoutes 初始化指标相关的API路由
func InitMonitorRoutes(mux *http.ServeMux) {
	// 监控指标接口
	mux.HandleFunc(path.MonitorMetricsProjectsPath, utils.JWTAuth(metrics.MetricsProjects, "monitor:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.MonitorQueryPath, utils.JWTAuth(metrics.MonitorMetricsQuery, "monitor:r", "", "即时查询"))
	mux.HandleFunc(path.MonitorMetricsListSSEPath, utils.JWTAuth(metrics.MonitorMetricsListSSE, "monitor:r", "", "SSE批量查询"))

	// 范围查询接口
	mux.HandleFunc(path.MonitorQueryRangePath, utils.JWTAuth(metrics.MonitorMetricsRangeQuery, "monitor:r", "", "范围查询"))
	mux.HandleFunc(path.MonitorMetricsListPath, utils.JWTAuth(metrics.MonitorMetricsList, "monitor:r", "", "监控视图列表"))

	// 监控视图接口
	mux.HandleFunc(path.MonitorViewProjectsPath, utils.JWTAuth(view.ViewProjects, "monitor:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.MonitorViewCreatePath, utils.JWTAuth(view.ViewCreate, "monitor:r", "", "创建监控视图"))
	mux.HandleFunc(path.MonitorViewDetailPath, utils.JWTAuth(view.ViewDetail, "monitor:r", "", "监控视图详情"))
	mux.HandleFunc(path.MonitorViewListPath, utils.JWTAuth(view.ViewList, "monitor:r", "", "监控视图列表"))
	mux.HandleFunc(path.MonitorViewUpdatePath, utils.JWTAuth(view.ViewUpdate, "monitor:r", "", "更新监控视图"))
	mux.HandleFunc(path.MonitorViewDeletePath, utils.JWTAuth(view.ViewDelete, "monitor:r", "", "删除监控视图"))

	// 告警接口
	mux.HandleFunc(path.MonitorAlertProjectsPath, utils.JWTAuth(alert.AlertProjects, "monitor:alert:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.MonitorAlertListPath, utils.JWTAuth(alert.AlertList, "monitor:alert:r", "", "获取告警列表"))
}
