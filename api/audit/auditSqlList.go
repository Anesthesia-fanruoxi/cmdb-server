package audit

import (
	"bytes"
	"cmdb/models/audit"
	"cmdb/models/elfk"
	"cmdb/models/sql"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// SqlAuditList SQL审计日志查询接口
func SqlAuditList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取当前用户的角色级别
	var role system.Role
	if err := database.DB.Table(middleware.TableRoles).
		Select("level").
		Where("id = ?", userInfo.RoleID).
		First(&role).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 解析请求参数
	var req audit.SQLAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 添加参数验证和日志
	middleware.Logger.Infof("分页参数: page=%d", req.Page)

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1 // 默认第1页
	}

	// 计算分页偏移量，固定每页20条
	const pageSize = 20
	offset := (req.Page - 1) * pageSize
	middleware.Logger.Infof("分页参数: page=%d, offset=%d", req.Page, offset)

	// 构建查询条件
	db := database.DB.
		Table(middleware.TableSqlAuditLog + " s").
		Select("s.query_id, s.city, s.country, s.client_ip, s.region, s.district, s.platform, u.nick_name, s.db_name, s.created_at, " +
			"s.execution_time, s.affected_rows, s.status, s.operation, " +
			"COUNT(a.id) AS operation_count").
		Joins("LEFT JOIN " + middleware.TableSqlAuditLog + " a ON s.query_id = a.query_id AND a.id != s.id").
		Joins("LEFT JOIN " + middleware.TableUsers + " u ON u.id = s.user_id").
		Where("s.operation = 'search'"). // 假设SQL执行的操作类型是execute
		Group("s.id, s.query_id, s.user_name, s.db_name, s.created_at, " +
			"s.execution_time, s.affected_rows, s.status, s.platform").
		Order("s.created_at DESC")

	// 超级管理员可以查看所有审计记录，普通用户只能查看其部门有权限的记录
	if role.Level > 0 {
		// 这里可以根据实际情况调整普通用户的可见范围
		// 例如可以改为只能查看自己的记录
		db = db.Where("s.user_id = ?", userInfo.UserID)
	}

	// 添加查询条件
	if req.DbName != "" {
		db = db.Where("s.db_name = ?", req.DbName)
	}
	if req.UserName != "" {
		db = db.Where("s.user_name LIKE ?", "%"+req.UserName+"%")
	}
	if req.QueryId != "" {
		db = db.Where("s.query_id = ?", req.QueryId)
	}

	// 处理时间查询条件
	if req.StartTime != "" {
		db = db.Where("s.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("s.created_at <= ?", req.EndTime)
	}

	// 获取总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录总数失败")
		return
	}
	middleware.Logger.Infof("总记录数: total=%d", total)

	// 分页查询，使用固定的pageSize
	var logs []audit.SQLAuditListItem
	if err := db.
		Limit(pageSize). // 使用固定的pageSize
		Offset(offset).
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录失败")
		return
	}
	middleware.Logger.Infof("当前页记录数: count=%d", len(logs))

	// 返回结果
	response := audit.SQLAuditResponse{
		Total: total,
		List:  logs,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// SqlAuditDetail 获取SQL审计详情接口
func SqlAuditDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取当前用户的角色级别
	var role system.Role
	if err := database.DB.Table(middleware.TableRoles).
		Select("level").
		Where("id = ?", userInfo.RoleID).
		First(&role).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 解析请求参数
	var req audit.SQLAuditDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 验证参数
	if req.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "查询ID不能为空")
		return
	}

	// 查询主记录信息 - 查询所有字段
	var searchDetail audit.SQLDetail
	db := database.DB.Table(middleware.TableSqlAuditLog).
		Where("query_id = ? AND operation = 'search'", req.QueryID) // 查询操作类型为search的记录

	// 超级管理员可以查看所有审计记录，普通用户只能查看自己的记录
	if role.Level > 0 {
		db = db.Where("user_id = ?", userInfo.UserID)
	}

	// 执行查询
	if err := db.First(&searchDetail).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "未找到对应的查询记录")
		return
	}

	// 查询相关的page操作
	var pageOperations []audit.SQLOperation
	if err := database.DB.Table(middleware.TableSqlAuditLog+" s").
		Select("u.nick_name, s.operation, s.page, s.execution_time, s.affected_rows, s.status, s.created_at").
		Joins("left join "+middleware.TableUsers+" u on u.id = s.user_id").
		Where("s.query_id = ? AND s.operation != 'search'", req.QueryID).
		Order("s.created_at ASC").
		Scan(&pageOperations).Error; err != nil {
		middleware.Logger.Errorf("查询相关操作失败: %v", err)
		// 不中断请求，继续处理
		pageOperations = []audit.SQLOperation{}
	}

	// 构建响应
	response := audit.SQLAuditDetailResponse{
		SearchDetail:   searchDetail,
		PageOperations: pageOperations,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// GetAuditTableName 根据审计类型获取对应的表名
func GetAuditTableName(auditType string) string {
	switch auditType {
	case "es":
		return middleware.TableEsAuditLog
	case "sql":
		return middleware.TableSqlAuditLog
	default:
		return middleware.TableSqlAuditLog // 默认使用SQL审计表
	}
}

// UserStatsHandler 通用的用户统计处理函数(SQL和ES共用)
func UserStatsHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析JSON
	var query audit.UserStatsQuery
	if err := json.Unmarshal(body, &query); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 验证起始时间和结束时间是否提供
	if query.StartTime == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "起始时间不能为空")
		return
	}

	if query.EndTime == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "结束时间不能为空")
		return
	}

	// 检查小时参数
	if query.Hour < -1 || query.Hour > 23 {
		middleware.ResponseError(w, http.StatusBadRequest, "小时参数无效，应为-1到23之间的整数，-1表示所有小时")
		return
	}

	// 解析时间参数
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.StartTime, time.Local)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "开始时间格式无效")
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.EndTime, time.Local)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "结束时间格式无效")
		return
	}

	// 获取审计表名
	tableName := GetAuditTableName(query.AuditType)
	middleware.Logger.Infof("查询时间范围: %s 至 %s, 小时: %d, 审计类型: %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		query.Hour,
		query.AuditType)

	// 查询用户统计
	var stats []audit.UserStatItem

	// 构建基本查询
	db := database.DB.Table(tableName+" l").
		Select("l.user_id, l.user_name, u.nick_name, COUNT(*) as count").
		Joins("LEFT JOIN "+middleware.TableUsers+" u ON l.user_id = u.id").
		Where("l.created_at >= ? AND l.created_at <= ?", startTime, endTime).
		// 只统计search操作
		Where("l.operation = 'search'")

	// 如果指定了小时，添加小时筛选条件
	if query.Hour != -1 {
		db = db.Where("HOUR(l.created_at) = ?", query.Hour)
	}

	// 执行查询
	err = db.Group("l.user_id, l.user_name, u.nick_name").
		Order("count DESC").
		Find(&stats).Error

	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询数据失败")
		return
	}

	// 构建响应
	result := map[string]interface{}{
		"total": len(stats),
		"list":  stats,
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}

// SqlUserStats 查询用户SQL使用统计
func SqlUserStats(w http.ResponseWriter, r *http.Request) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}

	// 解析JSON
	var query audit.UserStatsQuery
	if err := json.Unmarshal(body, &query); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 设置审计类型为SQL
	query.AuditType = "sql"

	// 重新编码请求体
	newBody, err := json.Marshal(query)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "处理请求数据失败")
		return
	}

	// 创建新的请求
	r.Body = io.NopCloser(bytes.NewBuffer(newBody))

	// 调用通用处理函数
	UserStatsHandler(w, r)
}

// UserListHandler 通用的用户查询列表处理函数(SQL和ES共用)
func UserListHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析JSON
	var query audit.UserListQuery
	if err := json.Unmarshal(body, &query); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 验证必填参数
	if query.UserID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	// 检查小时参数
	if query.Hour < 0 || query.Hour > 23 {
		middleware.ResponseError(w, http.StatusBadRequest, "小时参数无效，应为0-23之间的整数")
		return
	}

	// 验证起始时间和结束时间是否提供
	if query.StartTime == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "起始时间不能为空")
		return
	}

	if query.EndTime == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "结束时间不能为空")
		return
	}

	// 解析起始时间和结束时间
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.StartTime, time.Local)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "开始时间格式无效")
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", query.EndTime, time.Local)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "结束时间格式无效")
		return
	}

	// 获取审计表名
	tableName := GetAuditTableName(query.AuditType)
	middleware.Logger.Infof("查询时间范围: %s 至 %s, 小时: %d, 用户ID: %d, 审计类型: %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		query.Hour,
		query.UserID,
		query.AuditType)

	var result map[string]interface{}

	// 根据审计类型执行不同的查询
	switch query.AuditType {
	case "sql":
		// 查询用户的SQL记录并计算操作次数
		type SqlLogWithCount struct {
			sql.SqlAuditLog
			OperationCount int `json:"operation_count" gorm:"column:operation_count"`
		}
		var logs []SqlLogWithCount

		// 构建查询 - 在指定时间范围内查询特定小时的数据
		err = database.DB.Table(tableName+" s").
			Select("s.*, (SELECT COUNT(*) FROM "+tableName+" a WHERE a.query_id = s.query_id AND a.id != s.id) AS operation_count").
			// 用户ID条件
			Where("s.user_id = ?", query.UserID).
			// 时间范围条件
			Where("s.created_at >= ? AND s.created_at <= ?", startTime, endTime).
			// 小时条件：提取created_at的小时部分进行比较
			Where("HOUR(s.created_at) = ?", query.Hour).
			Where("s.operation = 'search'").
			Order("s.created_at DESC").
			Find(&logs).Error

		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询数据失败")
			return
		}

		// 构建响应
		result = map[string]interface{}{
			"total": len(logs),
			"list":  logs,
		}

	case "es":
		// 查询用户的ES记录并计算操作次数
		type EsLogWithCount struct {
			elfk.SysEsLog
			OperationCount int `json:"operation_count" gorm:"column:operation_count"`
		}
		var logs []EsLogWithCount

		// 构建查询 - 在指定时间范围内查询特定小时的数据
		err = database.DB.Table(tableName+" s").
			Select("s.*, (SELECT COUNT(*) FROM "+tableName+" a WHERE a.query_id = s.query_id AND a.id != s.id) AS operation_count").
			// 用户ID条件
			Where("s.user_id = ?", query.UserID).
			// 时间范围条件
			Where("s.created_at >= ? AND s.created_at <= ?", startTime, endTime).
			// 小时条件：提取created_at的小时部分进行比较
			Where("HOUR(s.created_at) = ?", query.Hour).
			// 只查询操作类型为search的记录
			Where("s.operation = 'search'").
			Order("s.created_at DESC").
			Find(&logs).Error

		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询数据失败")
			return
		}

		// 构建响应
		result = map[string]interface{}{
			"total": len(logs),
			"list":  logs,
		}

	default:
		middleware.ResponseError(w, http.StatusBadRequest, "无效的审计类型")
		return
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}

// SqlUserSqlList 查询用户执行的SQL列表
func SqlUserSqlList(w http.ResponseWriter, r *http.Request) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}

	// 解析JSON
	var query audit.UserListQuery
	if err := json.Unmarshal(body, &query); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 设置审计类型为SQL
	query.AuditType = "sql"

	// 重新编码请求体
	newBody, err := json.Marshal(query)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "处理请求数据失败")
		return
	}

	// 创建新的请求
	r.Body = io.NopCloser(bytes.NewBuffer(newBody))

	// 调用通用处理函数
	UserListHandler(w, r)
}
