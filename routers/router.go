package routers

import (
	"net/http"
)

// InitRouter 初始化所有路由
func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 初始化系统管理路由
	InitSystemRoutes(mux)    // 添加系统管理路由
	InitKnowledgeRoutes(mux) // 添加知识库路由
	InitJobRoutes(mux)       // 添加任务路由
	InitEsRoutes(mux)        // 添加ES路由
	InitAuditRoutes(mux)     // 添加审计路由
	InitRulesRoutes(mux)     // 添加规则路由
	InitMonitorRoutes(mux)   // 添加监控路由
	InitTaskRoutes(mux)      // 添加任务路由
	InitAssetsRoutes(mux)    // 添加资产管理路由
	InitAgentRoutes(mux)     // 添加Agent插件管理路由
	InitRedisRoutes(mux)     // 添加Redis管理路由
	return mux
}
