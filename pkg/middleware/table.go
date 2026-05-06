package middleware

// 资产相关表
const (
	TableProUpdate = "asset_release_record" // 发版记录表
)

// 系统管理
const (
	TableUsers        = "sys_users"         //用户表
	TableRoles        = "sys_roles"         // 角色表
	TableMenus        = "sys_menus"         // 菜单表
	TableRoleMenus    = "sys_role_menus"    // 角色菜单关联表
	TableDepartments  = "sys_departments"   // 部门表
	TableDeptProjects = "sys_dept_projects" // 部门项目关联表
	TableUserProjects = "sys_user_projects" // 用户项目关联表
	TableDictRecords  = "sys_dict_records"  // 字典表
	TableProjectDict  = "sys_project_dict"  // 项目字典表
	TableCategoryDict = "sys_category_dict" // 分类字典表
	TableSettings     = "sys_settings"      // 系统设置表
	TableSecurity     = "sys_security"      // 安全配置表
	TableUserDevice   = "sys_user_device"   // 用户设备绑定表
)

// 文档管理

const (
	TableDocuments       = "kb_documents"          // 文档表
	TableHistories       = "kb_document_histories" // 历史版本表
	TableRecycle         = "kb_document_recycle"   // 回收站表
	TablePersonal        = "kb_personal"           // 个人文档表
	TablePersonalRecycle = "kb_personal_recycle"   // 个人文档回收站表
	TablePublic          = "kb_public"             // 公开文档表
	TablePublicShare     = "kb_public_share"       // 公开文档分享表
	TablePublicHistory   = "kb_public_history"     // 公开文档历史表
	TablePublicRecycle   = "kb_public_recycle"     // 公开文档回收站表
)

// Job 相关表
const (
	TableJobInfo      = "job_info"      // 任务定义表
	TableJobProject   = "job_projects"  // 任务-项目关联表
	TableJobExecution = "job_execution" // 任务执行记录表
)

// es相关表
const (
	TableView          = "es_view"           // 视图表
	TableSharedKeyword = "es_shared_keyword" // 共享查询关键词表
)

// sql 相关表
const (
	TableSqlRule        = "sql_rule"         // sql 规则表
	TableSQLApply       = "sql_apply"        // sql 审批表
	TableSQLExport      = "sql_export"       // sql 导出表
	TableSQLApplyFlow   = "sql_apply_flow"   // sql 审批流程表
	TableSqlSharedQuery = "sql_shared_query" // sql 共享查询表
)

// 审计相关表
const (
	TableOperationLog      = "audit_operation_log" // 操作日志表
	TableLoginLog          = "audit_login_log"     // 登录日志表
	TableSqlAuditLog       = "audit_sql_log"       // SQL审计日志表
	TableEsAuditLog        = "audit_es_log"        // ES审计日志表
	TableIPLocationAudit   = "audit_ip_log"        // IP位置查询审计表
	TableEncryptDecryptLog = "audit_key_log"       // 加解密操作审计表
)

const (
	TableMonitorViews = "monitor_views" // 监控视图表
)

// Agent插件相关表
const (
	TablePluginStore = "agent_store" // 插件商店表
)

// 任务中心相关表
const (
	TableTaskCenter = "task_center" // 任务中心表
)
