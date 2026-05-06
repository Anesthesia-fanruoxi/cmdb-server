package path

// 任务管理路径
const (
	// sql查询
	SqlSearchProjectsPath = "/api/sql/search/projects" // 获取可访问的项目列表
	SqlSearchDataPath     = "/api/sql/search/data"
	SqlHistoryListPath    = "/api/sql/search/history"
	SqlSearchPagePath     = "/api/sql/search/page"
	SqlSearchDBPath       = "/api/sql/search/db"
	SqlSearchTBPath       = "/api/sql/search/tb"
	SqlSearchTBDetailPath = "/api/sql/search/tb-detail"
	SqlExportDataPath     = "/api/sql/search/export"
	SqlSearchPhonePath    = "/api/sql/search/phone"
	SqlSearchCancelPath   = "/api/sql/search/cancel" // 取消SQL查询

	// sql审批
	SqlApplyProjectsPath       = "/api/sql/apply/projects" // 获取可访问的项目列表
	SqlApplyListPath           = "/sse/sql/apply/list"
	SqlApplyDetailPath         = "/api/sql/apply/detail"
	SqlApplyCreatePath         = "/api/sql/apply/create"
	SqlApplyUpdatePath         = "/api/sql/apply/update"
	SqlApplyFeiShuCallbackPath = "/api/sql/apply/feishu/callback" // 飞书卡片回调接口

	// sql导出审批
	SqlExportProjectsPath         = "/api/sql/export/projects" // 获取可访问的项目列表
	SqlExportListPath             = "/sse/sql/export/list"
	SqlExportCreatePath           = "/api/sql/export/create"
	SqlExportUpdatePath           = "/api/sql/export/update"
	SqlExportDetailPath           = "/api/sql/export/detail"
	SqlExportUrlUpdatePath        = "/api/sql/export/url/update"
	SqlExportUrlDetailPath        = "/api/sql/export/url/detail"
	SqlExportFeiShuUpdatePath     = "/api/sql/export/feishu/update"     // 飞书机器人专用更新接口
	SqlExportDownloadPath         = "/api/sql/export/download"          // 下载导出文件（支持两种方式）
	SqlExportGenerateDownloadPath = "/api/sql/export/download/generate" // 生成临时下载链接（JWT鉴权）

	// sql流程
	SqlProcessListPath   = "/api/sql/process/list"
	SqlProcessUsersPath  = "/api/sql/process/users"
	SqlProcessCreatePath = "/api/sql/process/create"
	SqlProcessUpdatePath = "/api/sql/process/update"
	SqlProcessDeletePath = "/api/sql/process/delete"

	// sql规则
	SqlRulesListPath   = "/api/sql/rules/list"
	SqlRulesUpdatePath = "/api/sql/rules/update"
	SqlRulesCheckPath  = "/api/sql/rules/check"

	// SQL分析
	SqlAnalyzePath = "/api/sql/analyze"

	// SQL共享查询
	SqlSharedQueryListPath   = "/api/sql/shared/list"   // 共享查询列表
	SqlSharedQueryCreatePath = "/api/sql/shared/create" // 创建共享查询
	SqlSharedQueryUpdatePath = "/api/sql/shared/update" // 更新共享查询
	SqlSharedQueryDeletePath = "/api/sql/shared/delete" // 删除共享查询

	// DataBI Spark 接口
	DataBIProjectsPath     = "/api/sql/databi/projects"      // 获取可访问的项目列表
	DataBITablesPath       = "/sse/sql/databi/tables"        // SSE 实时推送表列表（读取 Redis）
	DataBIQueryPath        = "/api/sql/databi/query"         // 执行 Spark 查询
	DataBIRefreshPath      = "/api/sql/databi/refresh"       // 触发刷新任务（写入 Redis）
	DataBIExportPath       = "/api/sql/databi/export"        // 异步导出任务
	DataBIColumnListPath   = "/api/sql/databi/column/list"   // 获取表字段信息
	DataBIColumnUpdatePath = "/api/sql/databi/column/update" // 修改列注释
)
