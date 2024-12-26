package routers

import (
	"net/http"
)

// InitRoutes 初始化所有路由
func InitRoutes(mux *http.ServeMux) {
	// 系统管理相关路由
	InitSystemRoutes(mux)

	// 资产管理相关路由 - 测试环境
	InitTestAssetRoutes(mux)

	// 监控中心相关路由
	//monitor.InitMonitorRoutes(mux)
}
