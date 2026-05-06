package search

import "time"

// 操作类型常量
const (
	OperationExport = "export" // 导出操作
	OperationSearch = "search" // 查询操作
	OperationPage   = "page"   // 分页操作
)

// FilteredTables 需要过滤的表名（完全匹配）
var FilteredTables = []string{
	"sys_user",
	"sys_role",
	"sys_dept",
	"sys_menu",
	"sys_oper_log",
	"sys_logininfor",
	"sys_upload",
	"sys_config",
}

// SystemDatabases 系统数据库列表
var SystemDatabases = []string{
	"information_schema",
	"performance_schema",
	"INFORMATION_SCHEMA",
	"MYSQL",
	"mysql",
	"sys",
}

// 缓存配置
const (
	// PageCacheSize 每页缓存的大小（固定为20条，与前端分页一致）
	PageCacheSize = 20

	// BatchSize 每批次数据大小（100条）
	BatchSize = 100
)

// 使用time.Duration类型的常量
var (
	// QueryCacheTTL Redis缓存过期时间（1小时）
	QueryCacheTTL = 1 * time.Hour
)
