package routers

import (
	"cmdb/api/sql/analyze"
	"cmdb/api/sql/apply"
	"cmdb/api/sql/databi"
	"cmdb/api/sql/export"
	"cmdb/api/sql/process"
	"cmdb/api/sql/rule"
	"cmdb/api/sql/search"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitToolRoutes 初始化工具相关路由
func InitRulesRoutes(mux *http.ServeMux) {
	// sql查询
	mux.HandleFunc(path.SqlSearchProjectsPath, utils.JWTAuth(search.SqlProjects, "sql:search:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.SqlSearchDataPath, utils.JWTAuth(search.SqlSearchData, "sql:search:r", "", "SQL查询"))
	mux.HandleFunc(path.SqlHistoryListPath, utils.JWTAuth(search.SqlHistoryList, "sql:search:r", "", "sql查询历史记录"))
	mux.HandleFunc(path.SqlSearchDBPath, utils.JWTAuth(search.SqlSearchDB, "sql:search:r", "", "sql查询数据库列表"))
	mux.HandleFunc(path.SqlSearchTBPath, utils.JWTAuth(search.SqlSearchTB, "sql:search:r", "", "sql查询表列表"))
	mux.HandleFunc(path.SqlSearchTBDetailPath, utils.JWTAuth(search.SqlSearchTBDetail, "sql:search:r", "", "sql查询表详情"))
	mux.HandleFunc(path.SqlSearchPagePath, utils.JWTAuth(search.SqlSearchPage, "sql:search:r", "", "SQL分页查询"))
	mux.HandleFunc(path.SqlSearchCancelPath, utils.JWTAuth(search.SqlSearchCancel, "sql:search:r", "", "取消SQL查询"))
	mux.HandleFunc(path.SqlExportDataPath, utils.JWTAuth(search.SqlExportData, "sql:search:w", "", "SQL导出"))
	mux.HandleFunc(path.SqlSearchPhonePath, search.SearchPhone)

	// sql审批
	mux.HandleFunc(path.SqlApplyProjectsPath, utils.JWTAuth(apply.ApplyProjects, "sql:apply:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.SqlApplyListPath, utils.JWTAuth(apply.SqlApplyList, "sql:apply:r", "", "sql审批列表"))
	mux.HandleFunc(path.SqlApplyDetailPath, utils.JWTAuth(apply.SqlApplyDetail, "sql:apply:r", "", "sql审批详情"))
	mux.HandleFunc(path.SqlApplyCreatePath, utils.JWTAuth(apply.SqlApplyCreate, "sql:apply:w", "o", "创建sql审批"))
	mux.HandleFunc(path.SqlApplyUpdatePath, utils.JWTAuth(apply.SqlApplyUpdate, "sql:apply:w", "o", "更新sql审批"))
	// 飞书卡片回调接口（飞书服务器调用，使用IP白名单验证）
	mux.HandleFunc(path.SqlApplyFeiShuCallbackPath, utils.IPWhitelistMiddleware(apply.FeiShuApplyCallback))

	// sql 导出审批
	mux.HandleFunc(path.SqlExportProjectsPath, utils.JWTAuth(export.ExportProjects, "sql:export:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.SqlExportListPath, utils.JWTAuth(export.SqlExportList, "sql:export:r", "", "sql导出列表"))
	mux.HandleFunc(path.SqlExportDetailPath, utils.JWTAuth(export.SqlExportDetail, "sql:export:r", "", "sql导出详情"))
	mux.HandleFunc(path.SqlExportCreatePath, utils.JWTAuth(export.SqlExportCreate, "sql:export:w", "o", "创建sql导出"))
	mux.HandleFunc(path.SqlExportUpdatePath, utils.JWTAuth(export.SqlExportUpdate, "sql:export:w", "o", "更新sql导出"))
	mux.HandleFunc(path.SqlExportDownloadPath, utils.JWTAuth(export.SqlExportDownload, "sql:export:r", "", "下载导出文件"))
	// 生成临时下载链接（JWT鉴权）
	mux.HandleFunc(path.SqlExportGenerateDownloadPath, utils.JWTAuth(export.SqlExportGenerateDownload, "sql:export:r", "", "生成临时下载链接"))
	// URL操作不需要JWT认证，通过验证码验证
	mux.HandleFunc(path.SqlExportUrlDetailPath, utils.MobileJWTAuth(export.SqlExportUrlDetail, "sql:export:r", "", "sqlUrl导出详情"))
	mux.HandleFunc(path.SqlExportUrlUpdatePath, utils.MobileJWTAuth(export.SqlExportUrlUpdate, "sql:export:r", "", "sqlUrl导出更新"))
	// 飞书机器人专用接口，使用IP白名单验证
	mux.HandleFunc(path.SqlExportFeiShuUpdatePath, utils.IPWhitelistMiddleware(export.FeiShuExportUpdate))

	// sql流程
	mux.HandleFunc(path.SqlProcessListPath, utils.JWTAuth(process.SqlProcessList, "sql:process:r", "", "sql流程列表"))
	mux.HandleFunc(path.SqlProcessUsersPath, utils.JWTAuth(process.SqlProcessUsers, "sql:process:r", "", "获取审批人和执行人列表"))
	mux.HandleFunc(path.SqlProcessCreatePath, utils.JWTAuth(process.SqlProcessCreate, "sql:process:w", "o", "创建sql流程"))
	mux.HandleFunc(path.SqlProcessUpdatePath, utils.JWTAuth(process.SqlProcessUpdate, "sql:process:w", "o", "更新sql流程"))
	mux.HandleFunc(path.SqlProcessDeletePath, utils.JWTAuth(process.SqlProcessDelete, "sql:process:w", "o", "删除sql流程"))

	// sql规则
	mux.HandleFunc(path.SqlRulesListPath, utils.JWTAuth(rule.SqlRulesList, "sql:rules:r", "", "sql规则列表"))
	mux.HandleFunc(path.SqlRulesUpdatePath, utils.JWTAuth(rule.SqlRulesUpdate, "sql:rules:w", "o", "更新sql规则"))
	mux.HandleFunc(path.SqlRulesCheckPath, utils.JWTAuth(rule.RuleCheckAPI, "sql:rules:r", "", "SQL规则检查"))

	// sql分析
	mux.HandleFunc(path.SqlAnalyzePath, utils.JWTAuth(analyze.SQLAnalyzeHandler, "", "", "SQL分析"))

	// sql共享查询
	mux.HandleFunc(path.SqlSharedQueryListPath, utils.JWTAuth(search.SharedQueryList, "sql:search:r", "", "共享查询列表"))
	mux.HandleFunc(path.SqlSharedQueryCreatePath, utils.JWTAuth(search.SharedQueryCreate, "sql:search:r", "o", "创建共享查询"))
	mux.HandleFunc(path.SqlSharedQueryUpdatePath, utils.JWTAuth(search.SharedQueryUpdate, "sql:search:r", "o", "更新共享查询"))
	mux.HandleFunc(path.SqlSharedQueryDeletePath, utils.JWTAuth(search.SharedQueryDelete, "sql:search:r", "o", "删除共享查询"))

	// DataBI Spark 接口
	mux.HandleFunc(path.DataBIProjectsPath, utils.JWTAuth(databi.DataBIProjects, "sql:databi:r", "", "获取DataBI可访问项目列表"))
	mux.HandleFunc(path.DataBITablesPath, utils.JWTAuth(databi.DataBITables, "sql:databi:r", "", "SSE实时推送表列表（读取Redis）"))
	mux.HandleFunc(path.DataBIQueryPath, utils.JWTAuth(databi.DataBIQuery, "sql:databi:r", "", "执行Spark查询"))
	mux.HandleFunc(path.DataBIRefreshPath, utils.JWTAuth(databi.DataBIRefresh, "sql:databi:r", "", "触发刷新任务（写入Redis）"))
	mux.HandleFunc(path.DataBIExportPath, utils.JWTAuth(databi.DataBIExport, "sql:databi:w", "", "异步导出任务"))
	mux.HandleFunc(path.DataBIColumnListPath, utils.JWTAuth(databi.DataBIColumnList, "sql:databi:r", "", "获取表字段信息"))
	mux.HandleFunc(path.DataBIColumnUpdatePath, utils.JWTAuth(databi.DataBIColumnUpdate, "sql:databi:w", "o", "修改列注释"))
}
