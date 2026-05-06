package path

// 审计管理路径
const (
	// 搜索相关
	AuditOperationListPath = "/api/audit/operation/list" // 操作日志列表
	AuditLoginListPath     = "/api/audit/login/list"     // 登录日志列表
	AuditSearchListPath    = "/api/audit/search/list"    // ES审计日志列表
	AuditSqlListPath       = "/api/audit/sql/list"       // SQL审计日志列表
	AuditSearchDetailPath  = "/api/audit/search/detail"  // ES审计日志详情
	AuditSqlListDetailPath = "/api/audit/sql/detail"     // SQL审计日志详情

	// 按小时查询
	AuditEsHourlyPath = "/api/audit/es/hourly" // ES审计按小时详情

	// 用户SQL统计
	AuditSqlUserStatsPath = "/api/audit/sql/user/stats" // 用户SQL使用统计
	AuditSqlUserListPath  = "/api/audit/sql/user/list"  // 用户SQL列表

	// 用户ES统计
	AuditEsUserStatsPath    = "/api/audit/es/user/stats"    // 用户ES使用统计
	AuditEsUserListPath     = "/api/audit/es/user/list"     // 用户ES列表
	AuditSqlQueryDetailPath = "/api/audit/sql/query/detail" // SQL查询详情

	// 审计分析
	AuditDataAnalysisPath = "/api/audit/analysis" // 数据分析

	// 加解密审计
	AuditKeyListPath = "/api/audit/key/list" // 加解密审计日志列表

	// IP审计
	AuditIpListPath = "/api/audit/ip/list" // IP位置查询审计日志列表
)
