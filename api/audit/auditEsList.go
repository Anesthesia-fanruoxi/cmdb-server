package audit

import (
	"bytes"
	"cmdb/models/audit"
	"cmdb/models/elfk"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// ESAuditRequest ES审计查询请求结构
type ESAuditRequest struct {
	UserName    string `json:"user_name"`    // 用户名
	Project     string `json:"project"`      // 项目名称
	QueryId     string `json:"query_id"`     // 查询id
	StartTime   string `json:"start_time"`   // 开始时间
	EndTime     string `json:"end_time"`     // 结束时间
	CurrentPage int    `json:"current_page"` // 当前页码
}

// ESAuditListItem 添加一个新的自定义结构体来匹配查询结果
type ESAuditListItem struct {
	QueryID        string    `json:"query_id" gorm:"column:query_id"`
	NickName       string    `json:"nick_name" gorm:"column:nick_name"`
	ProjectName    string    `json:"project_name" gorm:"column:project_name"`
	ViewName       string    `json:"view_name" gorm:"column:view_name"`
	SearchTime     time.Time `json:"search_time" gorm:"column:search_time"`
	Keyword        string    `json:"keyword" gorm:"column:q_keyword"`
	ResponseCode   int       `json:"response_code" gorm:"column:response_code"`
	QueryTimeMs    int       `json:"query_time_ms" gorm:"column:query_time_ms"`
	DocCount       int       `json:"doc_count" gorm:"column:doc_count"`
	Operation      string    `json:"operation" gorm:"column:operation"`
	OperationCount int       `json:"operation_count" gorm:"column:operation_count"`
	ClientIP       string    `json:"client_ip" gorm:"column:client_ip"`
	City           string    `json:"city" gorm:"column:city"`
	Country        string    `json:"country" gorm:"column:country"`
	Region         string    `json:"region" gorm:"column:region"`
	District       string    `json:"district" gorm:"column:district"`
}

// ESAuditResponse ES审计查询响应结构
type ESAuditResponse struct {
	Total int64             `json:"total"` // 总记录数
	List  []ESAuditListItem `json:"list"`  // 记录列表
}

// ESAuditDetailRequest 审计详情请求结构
type ESAuditDetailRequest struct {
	QueryID string `json:"query_id"` // 查询ID
}

// EsOperation 操作记录
type EsOperation struct {
	NickName    string    `json:"nick_name"`     // 昵称
	Operation   string    `json:"operation"`     // 操作
	Page        int       `json:"page"`          // 页码
	QueryTimeMs int64     `json:"query_time_ms"` // 查询耗时
	DocCount    int       `json:"doc_count"`     // 文档数量
	CreatedAt   time.Time `json:"created_at"`    // 操作时间
}

// ESAuditDetailResponse 审计详情响应结构
type ESAuditDetailResponse struct {
	SearchDetail   elfk.SysEsLog `json:"search_detail"`   // 查询操作详情
	PageOperations []EsOperation `json:"page_operations"` // 相关翻页操作列表
}

// AuditSearchList ES审计日志查询接口
func AuditSearchList(w http.ResponseWriter, r *http.Request) {
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
	var req ESAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.CurrentPage <= 0 {
		req.CurrentPage = 1 // 默认第1页
	}

	// 计算分页偏移量，固定每页20条
	const pageSize = 20
	offset := (req.CurrentPage - 1) * pageSize
	middleware.Logger.Infof("分页参数: current_page=%d, offset=%d", req.CurrentPage, offset)

	// 构建查询条件
	db := database.DB.
		Table(middleware.TableEsAuditLog + " s").
		Select("s.query_id, s.city,s.country,s.client_ip,s.region,s.district,u.nick_name, p.project_name, s.view_name, s.created_at AS search_time, " +
			"s.q_keyword, s.response_code, s.query_time_ms, s.doc_count, s.operation, " +
			"COUNT(a.id) AS operation_count").
		Joins("LEFT JOIN " + middleware.TableEsAuditLog + " a ON s.query_id = a.query_id AND a.id != s.id").
		Joins("LEFT JOIN " + middleware.TableUsers + " u ON u.id = s.user_id").
		Joins("LEFT JOIN " + middleware.TableProjectDict + " p ON p.project = s.project").
		Where("s.operation = 'search'").
		Group("s.id, s.query_id, s.user_name, s.project, s.view_name, s.created_at, " +
			"s.q_keyword, s.response_code, s.query_time_ms, s.doc_count").
		Order("s.created_at DESC")

	// 超级管理员可以查看所有审计记录，普通用户只能查看其部门有权限的项目审计
	if role.Level > 0 {
		// 获取用户部门可访问的项目列表
		var projects []string
		if err := database.DB.Model(&system.DeptProject{}).
			Where("dept_id = ?", userInfo.DeptID).
			Pluck("project", &projects).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取部门项目权限失败")
			return
		}

		// 如果部门没有任何项目权限，直接返回空结果
		if len(projects) == 0 {
			response := ESAuditResponse{
				Total: 0,
				List:  []ESAuditListItem{},
			}
			middleware.ResponseSuccess(w, "查询成功", response)
			return
		}

		// 按部门可访问项目进行过滤
		db = db.Where("p.project IN ?", projects)
	}

	// 添加查询条件
	if req.Project != "" {
		db = db.Where("p.project = ?", req.Project)
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

	// 分页查询，使用固定的pageSize
	var logs []ESAuditListItem
	if err := db.
		Limit(pageSize). // 使用固定的pageSize
		Offset(offset).
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录失败")
		return
	}

	// 返回结果
	response := ESAuditResponse{
		Total: total,
		List:  logs,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// EsUserStats 查询用户ES使用统计
func EsUserStats(w http.ResponseWriter, r *http.Request) {
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

	// 设置审计类型为ES
	query.AuditType = "es"

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

// EsUserList 查询用户执行的ES查询列表
func EsUserList(w http.ResponseWriter, r *http.Request) {
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

	// 设置审计类型为ES
	query.AuditType = "es"

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

// EsAuditHourlyDetails 查询按小时分组的ES审计详情
func EsAuditHourlyDetails(w http.ResponseWriter, r *http.Request) {
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
	var query audit.HourlyAuditQuery
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

	middleware.Logger.Infof("查询时间范围: %s 至 %s, 小时: %d", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"), query.Hour)

	// 构建查询
	db := database.DB.Table(middleware.TableEsAuditLog)

	// 添加时间范围条件
	db = db.Where("created_at >= ? AND created_at <= ?", startTime, endTime)

	// 如果指定了小时，添加小时条件
	if query.Hour != -1 {
		db = db.Where("HOUR(created_at) = ?", query.Hour)
	}

	// 查询所有数据
	var logs []elfk.SysEsLog

	if err := db.Order("created_at DESC").
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询数据失败")
		return
	}

	// 构建响应
	result := map[string]interface{}{
		"total": len(logs),
		"list":  logs,
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}

// AuditSearchDetail 获取ES审计详情接口
func AuditSearchDetail(w http.ResponseWriter, r *http.Request) {
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
	var req ESAuditDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 验证参数
	if req.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "查询ID不能为空")
		return
	}

	// 查询search类型的记录 - 查询所有字段的详细信息
	var searchDetail elfk.SysEsLog
	db := database.DB.Table(middleware.TableEsAuditLog).
		// 查询所有字段
		Where("query_id = ? AND operation = 'search'", req.QueryID)

	// 超级管理员可以查看所有审计记录，普通用户只能查看其部门有权限的项目审计
	if role.Level > 0 {
		// 获取用户部门可访问的项目列表
		var projects []string
		if err := database.DB.Model(&system.DeptProject{}).
			Where("dept_id = ?", userInfo.DeptID).
			Pluck("project", &projects).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取部门项目权限失败")
			return
		}

		// 如果部门没有任何项目权限，直接返回空结果
		if len(projects) == 0 {
			middleware.ResponseError(w, http.StatusForbidden, "无权查看该记录")
			return
		}

		// 按部门可访问项目进行过滤
		db = db.Where("project IN ?", projects)
	}

	// 执行查询
	if err := db.First(&searchDetail).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "未找到对应的查询记录")
		return
	}

	// 查询相关的page操作 - 只查询需要的字段
	var Operations []EsOperation
	if err := database.DB.Table(middleware.TableEsAuditLog+" l").
		Select("u.nick_name, l.operation, l.page, l.query_time_ms, l.doc_count, l.created_at").
		Joins("left join "+middleware.TableUsers+" u on u.id = l.user_id").
		Where("l.query_id = ? AND l.operation != 'search'", req.QueryID).
		Order("created_at ASC").
		Scan(&Operations).Error; err != nil {
		middleware.Logger.Errorf("查询页面操作失败: %v", err)
		// 不中断请求，继续处理
		Operations = []EsOperation{}
	}

	// 构建响应
	response := ESAuditDetailResponse{
		SearchDetail:   searchDetail,
		PageOperations: Operations,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
