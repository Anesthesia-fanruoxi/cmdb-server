package audit

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// IpAuditListItem IP审计查询结果列表项
type IpAuditListItem struct {
	ID        int64  `json:"id" gorm:"column:id"`
	RequestIP string `json:"request_ip" gorm:"column:request_ip"`
	QueryIP   string `json:"query_ip" gorm:"column:query_ip"`
	Status    int    `json:"status" gorm:"column:status"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
	Date      string `json:"date" gorm:"column:date"`
}

// IpAuditResponse IP审计查询响应结构
type IpAuditResponse struct {
	Total int64             `json:"total"` // 总记录数
	List  []IpAuditListItem `json:"list"`  // 记录列表
}

// IpAuditList IP位置查询审计日志查询接口
func IpAuditList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
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

	// 从URL参数获取查询条件
	query := r.URL.Query()

	// 解析分页参数
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 计算分页偏移量，固定每页20条
	const pageSize = 20
	offset := (page - 1) * pageSize
	middleware.Logger.Infof("IP审计分页参数: page=%d, offset=%d", page, offset)

	// 构建查询条件
	db := database.DB.
		Table(middleware.TableIPLocationAudit).
		Select("id, request_ip, query_ip, status, " +
			"DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at, date").
		Order("created_at DESC")

	// 添加查询条件（从URL参数获取）
	if requestIP := query.Get("request_ip"); requestIP != "" {
		db = db.Where("request_ip = ?", requestIP)
	}
	if queryIP := query.Get("query_ip"); queryIP != "" {
		db = db.Where("query_ip = ?", queryIP)
	}
	if statusStr := query.Get("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			db = db.Where("status = ?", status)
		}
	}

	// 处理时间查询条件
	if startTime := query.Get("start_time"); startTime != "" {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime := query.Get("end_time"); endTime != "" {
		db = db.Where("created_at <= ?", endTime)
	}

	// 获取总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录总数失败")
		return
	}
	middleware.Logger.Infof("IP审计总记录数: total=%d", total)

	// 分页查询
	var logs []IpAuditListItem
	if err := db.
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录失败")
		return
	}
	middleware.Logger.Infof("IP审计当前页记录数: count=%d", len(logs))

	// 返回结果
	response := IpAuditResponse{
		Total: total,
		List:  logs,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
