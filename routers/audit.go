package routers

import (
	"cmdb/api/audit"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitAuditRoutes 初始化审计相关路由
func InitAuditRoutes(mux *http.ServeMux) {
	// 操作日志相关路由
	mux.HandleFunc(path.AuditOperationListPath, utils.JWTAuth(audit.AuditOperationLogList, "audit:operation:r", "", ""))
	// 登录日志相关路由
	mux.HandleFunc(path.AuditLoginListPath, utils.JWTAuth(audit.AuditLoginLogList, "audit:login:r", "", ""))
	// ES审计日志相关路由
	mux.HandleFunc(path.AuditSearchListPath, utils.JWTAuth(audit.AuditSearchList, "audit:search:r", "", ""))
	mux.HandleFunc(path.AuditSearchDetailPath, utils.JWTAuth(audit.AuditSearchDetail, "audit:search:r", "", ""))
	mux.HandleFunc(path.AuditEsHourlyPath, utils.JWTAuth(audit.EsAuditHourlyDetails, "audit:search:r", "", "ES审计按小时详情"))
	// SQL审计日志相关路由
	mux.HandleFunc(path.AuditSqlListPath, utils.JWTAuth(audit.SqlAuditList, "audit:sql:r", "", ""))
	mux.HandleFunc(path.AuditSqlListDetailPath, utils.JWTAuth(audit.SqlAuditDetail, "audit:sql:r", "", ""))
	// 已将SqlAuditHourlyDetails功能合并到SqlUserStats和SqlUserSqlList中，此路由保留但实际不再使用

	// 用户SQL统计相关路由
	mux.HandleFunc(path.AuditSqlUserStatsPath, utils.JWTAuth(audit.SqlUserStats, "audit:sql:r", "", "用户SQL使用统计"))
	mux.HandleFunc(path.AuditSqlUserListPath, utils.JWTAuth(audit.SqlUserSqlList, "audit:sql:r", "", "用户SQL列表"))

	// 用户ES统计相关路由
	mux.HandleFunc(path.AuditEsUserStatsPath, utils.JWTAuth(audit.EsUserStats, "audit:search:r", "", "用户ES使用统计"))
	mux.HandleFunc(path.AuditEsUserListPath, utils.JWTAuth(audit.EsUserList, "audit:search:r", "", "用户ES列表"))
	mux.HandleFunc(path.AuditSqlQueryDetailPath, utils.JWTAuth(audit.SqlAuditDetail, "audit:sql:r", "", "SQL查询详情"))
	// 数据分析相关路由
	mux.HandleFunc(path.AuditDataAnalysisPath, utils.JWTAuth(audit.GetAuditStatsHandler, "audit:sql:r", "", ""))

	// 加解密审计相关路由
	mux.HandleFunc(path.AuditKeyListPath, utils.JWTAuth(audit.KeyAuditList, "audit:key:r", "", ""))

	// IP审计相关路由
	mux.HandleFunc(path.AuditIpListPath, utils.JWTAuth(audit.IpAuditList, "audit:ip:r", "", "IP位置查询审计"))
}
