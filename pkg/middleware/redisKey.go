package middleware

import (
	"fmt"
)

// Redis键名路径常量 - 按功能模块分类
const (
	// ======== 用户认证相关 ========
	KeyPathUserToken       = "cmdb:auth:user:%d:token:pc"     // 用户PC端Token: cmdb:auth:user:1:token
	KeyPathMobileUserToken = "cmdb:auth:user:%d:token:mobile" // 用户移动端Token: cmdb:auth:mobile:user:1:token
	KeyPathLoginFailed     = "cmdb:auth:login:failed:%s"      // 登录失败次数: cmdb:auth:login:failed:admin

	// ======== 角色权限相关 ========
	KeyPathRolePermissions = "cmdb:rbac:role:%d:permissions" // 角色权限: cmdb:rbac:role:1:permissions

	// ======== 安全防护相关 ========
	KeyPathIPLoginFailed  = "cmdb:security:ip:%s:login:failed" // IP登录失败: cmdb:security:ip:192.168.1.1:login:failed
	KeyPathIPBlacklist    = "cmdb:security:blacklist:ip:%s"    // IP黑名单: cmdb:security:blacklist:ip:192.168.1.1
	KeyPathIPRateLimit    = "cmdb:security:ratelimit:ip:%s"    // IP限流: cmdb:security:ratelimit:ip:192.168.1.1
	KeyPathSecurityConfig = "cmdb:security:config"

	// ======== 知识库文档相关 ========
	KeyPathDocShare    = "cmdb:doc:share:%d" // 文档分享: cmdb:doc:share:1
	KeyPathDocShareStr = "cmdb:doc:share:%s" // 文档分享: cmdb:doc:share:1

	// ======== SQL相关 ========
	KeyPathSQLApply          = "cmdb:sql:apply:%d" // SQL申请: cmdb:sql:apply:1
	KeyPathSqlExport         = "cmdb:sql:export:%s"
	KeyPathSqlSearch         = "cmdb:sql:search:%s"
	KeyPathSqlExportProgress = "cmdb:sql:export:progress:%d" // SQL导出进度: cmdb:sql:export:progress:1

	// SQL查询分页缓存相关
	KeyPathSqlSearchMain   = "cmdb:sql:search:main:%s"   // SQL查询主信息: cmdb:sql:search:main:qid-xxx
	KeyPathSqlSearchBatch  = "cmdb:sql:search:batch:%s"  // SQL查询批次: cmdb:sql:search:batch:qid-xxx-1
	KeyPathSqlSearchCancel = "cmdb:sql:search:cancel:%s" // SQL查询取消标记: cmdb:sql:search:cancel:qid-xxx

	// ======== ES相关 ========
	KeyPathESView   = "cmdb:es:view:%d"
	KeyPathESSearch = "cmdb:es:search:%s"

	// ======== DataBI相关 ========
	KeyPathDataBITables        = "cmdb:databi:tables:%s"         // DataBI表列表缓存: cmdb:databi:tables:project_name
	KeyPathDataBIRefreshStatus = "cmdb:databi:refresh:status:%s" // DataBI刷新状态: cmdb:databi:refresh:status:project_name
	KeyPathDataBIRefreshLock   = "cmdb:databi:refresh:lock:%s"   // DataBI刷新锁: cmdb:databi:refresh:lock:project_name

	// ======== 作业平台相关 ========
	KeyPathAgent = "cmdb:agent:%s" // Agent信息: cmdb:agent:agent_id
	KeyPathNode  = "cmdb:node:%s"  // 节点信息: cmdb:node:hostname

	// ======== 项目发版相关 ========
	KeyPathTaskRecord = "cmdb:task:record:%s" // 任务记录: cmdb:task:record:task_id

	// ======== IP项目关联相关 ========
	KeyPathIPProject      = "cmdb:ip_project:ip:%s"       // IP关联的项目集合: cmdb:ip_project:ip:192.168.1.1
	KeyPathProjectIP      = "cmdb:ip_project:project:%s"  // 项目关联的IP集合: cmdb:ip_project:project:cmdb
	KeyPathIPLocationStat = "cmdb:ip_location:stat:%s:%s" // IP位置查询统计: cmdb:ip_location:stat:20251028:192.168.1.1

	// ======== 设备自动登录相关 ========
	KeyPathDeviceChallenge     = "cmdb:device:challenge:%s"      // 设备挑战码: cmdb:device:challenge:xxx
	KeyPathDeviceChallengeRate = "cmdb:device:challenge:rate:%s" // 挑战码频率限制: cmdb:device:challenge:rate:machine_id
	KeyPathDeviceLoginFail     = "cmdb:device:login:fail:%s"     // 设备登录失败计数: cmdb:device:login:fail:machine_id

	// ======== 应用更新相关 ========
	KeyPathAppVersion = "cmdb:app:version" // 应用版本信息
)

// ======== 用户认证相关函数 ========

// GetUserTokenKey 获取PC端用户Token的Redis键名
func GetUserTokenKey(userID int64) string {
	return fmt.Sprintf(KeyPathUserToken, userID)
}

// GetMobileUserTokenKey 获取移动端用户Token的Redis键名
func GetMobileUserTokenKey(userID int64) string {
	return fmt.Sprintf(KeyPathMobileUserToken, userID)
}

// GetLoginFailedKey 获取用户登录失败次数的Redis键名
func GetLoginFailedKey(username string) string {
	return fmt.Sprintf(KeyPathLoginFailed, username)
}

// GetTokenInfoKey 获取token信息的Redis键名
func GetTokenInfoKey(token string) string {
	return "cmdb:token:" + token + ":info"
}

// ======== 角色权限相关函数 ========

// GetRolePermissionsKey 获取角色权限的Redis键名
func GetRolePermissionsKey(roleID uint) string {
	return fmt.Sprintf(KeyPathRolePermissions, roleID)
}

// ======== 安全防护相关函数 ========
// GetIPLoginFailedKey 获取IP登录失败次数的Redis键名
func GetIPLoginFailedKey(ip string) string {
	return fmt.Sprintf(KeyPathIPLoginFailed, ip)
}

// GetIPBlacklistKey 获取IP黑名单的Redis键名
func GetIPBlacklistKey(ip string) string {
	return fmt.Sprintf(KeyPathIPBlacklist, ip)
}

// GetIPRateLimitKey 获取IP限流的Redis键名queryID
func GetIPRateLimitKey(ip string) string {
	return fmt.Sprintf(KeyPathIPRateLimit, ip)
}

// ======== 知识库文档相关函数 ========
// GetDocShareKeyStr 获取文档分享的Redis键名(字符串ID)
func GetDocShareKey(docID int64) string {
	return fmt.Sprintf(KeyPathDocShare, docID)
}
func GetDocShareKeyStr(docID string) string {
	return fmt.Sprintf(KeyPathDocShareStr, docID)
}

// ======== SQL相关函数 ========

func GetSqlExportKey(exportID string) string {
	return fmt.Sprintf(KeyPathSqlExport, exportID)
}
func GetSqlSearchKey(searchID string) string {
	return fmt.Sprintf(KeyPathSqlSearch, searchID)
}

// GetSqlExportProgressKey 获取SQL导出进度的Redis键名
func GetSqlExportProgressKey(exportID int64) string {
	return fmt.Sprintf(KeyPathSqlExportProgress, exportID)
}

// ======== SQL查询分页缓存相关函数 ========

// GetSqlSearchMainKey 获取SQL查询主信息的Redis键名
func GetSqlSearchMainKey(queryID string) string {
	return fmt.Sprintf(KeyPathSqlSearchMain, queryID)
}

// GetSqlSearchBatchKey 获取SQL查询批次的Redis键名
func GetSqlSearchBatchKey(batchID string) string {
	return fmt.Sprintf(KeyPathSqlSearchBatch, batchID)
}

// GetSqlSearchCancelKey 获取SQL查询取消标记的Redis键名
func GetSqlSearchCancelKey(queryID string) string {
	return fmt.Sprintf(KeyPathSqlSearchCancel, queryID)
}

// ======== ES相关函数 ========
func GetESViewKey(viewID uint) string {
	return fmt.Sprintf(KeyPathESView, viewID)
}
func GetESSearchKey(searchID string) string {
	return fmt.Sprintf(KeyPathESSearch, searchID)
}

// ======== DataBI相关函数 ========

// GetDataBITablesKey 获取DataBI表列表缓存的Redis键名
func GetDataBITablesKey(project string) string {
	return fmt.Sprintf(KeyPathDataBITables, project)
}

// GetDataBIRefreshStatusKey 获取DataBI刷新状态的Redis键名
func GetDataBIRefreshStatusKey(project string) string {
	return fmt.Sprintf(KeyPathDataBIRefreshStatus, project)
}

// GetDataBIRefreshLockKey 获取DataBI刷新锁的Redis键名
func GetDataBIRefreshLockKey(project string) string {
	return fmt.Sprintf(KeyPathDataBIRefreshLock, project)
}

// GetTaskRecordKey 获取任务记录的Redis键名
func GetTaskRecordKey(taskID string) string {
	return fmt.Sprintf(KeyPathTaskRecord, taskID)
}

// ======== IP项目关联相关函数 ========

// GetIPProjectKey 获取IP关联的项目集合Redis键名
func GetIPProjectKey(ip string) string {
	return fmt.Sprintf(KeyPathIPProject, ip)
}

// GetProjectIPKey 获取项目关联的IP集合Redis键名
func GetProjectIPKey(project string) string {
	return fmt.Sprintf(KeyPathProjectIP, project)
}

// GetIPLocationStatKey 获取IP位置查询统计Redis键名
func GetIPLocationStatKey(date, requestIP string) string {
	return fmt.Sprintf(KeyPathIPLocationStat, date, requestIP)
}

// ======== 设备自动登录相关函数 ========

// GetDeviceChallengeKey 获取设备挑战码的Redis键名
func GetDeviceChallengeKey(challenge string) string {
	return fmt.Sprintf(KeyPathDeviceChallenge, challenge)
}

// GetDeviceChallengeRateKey 获取设备挑战码频率限制的Redis键名
func GetDeviceChallengeRateKey(machineID string) string {
	return fmt.Sprintf(KeyPathDeviceChallengeRate, machineID)
}

// GetDeviceLoginFailKey 获取设备登录失败计数的Redis键名
func GetDeviceLoginFailKey(machineID string) string {
	return fmt.Sprintf(KeyPathDeviceLoginFail, machineID)
}
