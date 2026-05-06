package main

import (
	"cmdb/api/elfk/search"
	"cmdb/api/job/scheduler"
	"cmdb/api/monitor/alert"
	"cmdb/api/sql/apply"
	"cmdb/api/sql/databi"
	sqlsearch "cmdb/api/sql/search"
	"cmdb/api/task"
	"cmdb/config"
	"cmdb/pkg/database"
	"cmdb/pkg/initialize"
	"cmdb/pkg/middleware"
	"cmdb/routers"
	"cmdb/utils"
	"fmt"
	"net/http"
	"os"
)

func init() {
	// 初始化配置
	if err := config.Init(); err != nil {
		// 配置初始化失败时直接使用标准输出，确保错误信息显示
		_, err := fmt.Fprintf(os.Stderr, "致命错误: 初始化配置失败: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	// 初始化日志系统 - 提前初始化，以便使用StartupLogger
	middleware.InitLogger()

	// 初始化数据库连接
	if err := database.InitMySQL(&config.GetConfig().MySQL); err != nil {
		middleware.StartupLogger.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化Redis连接
	redisCfg := config.GetConfig().Redis
	addr := fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port)
	if err := database.InitRedis(addr, redisCfg.Password, redisCfg.DB); err != nil {
		middleware.StartupLogger.Fatalf("Redis初始化失败: %v", err)
	}

	// 初始化Prometheus客户端
	if err := database.InitPrometheus(&config.GetConfig().Prometheus); err != nil {
		middleware.StartupLogger.Errorf("Prometheus初始化失败: %v", err)
	}

	// 初始化SLS客户端管理器
	if err := search.InitSLSManager(&config.GetConfig().SLS); err != nil {
		middleware.StartupLogger.Errorf("SLS客户端管理器初始化失败: %v", err)
	}

	// 初始化角色权限缓存
	if err := utils.InitRolePermissions(); err != nil {
		middleware.StartupLogger.Errorf("初始化角色权限缓存失败: %v", err)
	}

	// 初始化验证器
	middleware.InitValidator()

	initialize.StartViewCacheRefreshScheduler()

	// 初始化系统权限
	if err := initialize.InitSystemPrivileges(); err != nil {
		middleware.StartupLogger.Errorf("初始化系统权限失败: %v", err)
	}

	// 初始化任务管理系统
	if err := task.InitTaskSystem(); err != nil {
		middleware.StartupLogger.Errorf("初始化任务系统失败: %v", err)
	}

	// 注册数据分析执行器
	search.RegisterAnalysisExecutor()

	// 注册ES导出执行器
	search.RegisterESExportExecutor()

	// 注册SQL导出执行器
	sqlsearch.RegisterSQLExportExecutor()

	// 注册DataBI导出执行器
	databi.RegisterDatabiExportExecutor()

	// 注册Pod执行器
	task.RegisterPodExecutors()

	// 启动Job调度器
	jobScheduler := scheduler.GetScheduler()
	jobScheduler.Start()
	middleware.StartupLogger.Info("Job任务调度器已启动")

	// 启动SQL审批定时执行调度器
	apply.StartApplyScheduler()
}

// main函数是程序的入口
func main() {
	// 初始化API路由
	r := routers.InitRouter()

	// 初始化告警系统
	alert.InitAlert()

	// 启动服务器
	addr := ":8080"
	middleware.StartupLogger.Infof("服务运行在 http://localhost%s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		middleware.StartupLogger.Fatalf("服务启动失败: %v", err)
	}
}
