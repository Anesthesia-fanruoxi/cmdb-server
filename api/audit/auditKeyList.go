package audit

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// KeyAuditListItem 加解密审计查询结果列表项
type KeyAuditListItem struct {
	ID            int64  `json:"id" gorm:"column:id"`
	UserID        int    `json:"user_id" gorm:"column:user_id"`
	UserName      string `json:"user_name" gorm:"column:user_name"`
	NickName      string `json:"nick_name" gorm:"column:nick_name"`
	OperationType string `json:"operation_type" gorm:"column:operation_type"`
	Project       string `json:"project" gorm:"column:project"`
	ProjectName   string `json:"project_name" gorm:"column:project_name"`
	BatchCount    int    `json:"batch_count" gorm:"column:batch_count"`
	Status        string `json:"status" gorm:"column:status"`
	IPAddress     string `json:"ip_address" gorm:"column:ip_address"`
	CreatedAt     string `json:"created_at" gorm:"column:created_at"`
}

// KeyAuditResponse 加解密审计查询响应结构
type KeyAuditResponse struct {
	Total int64              `json:"total"` // 总记录数
	List  []KeyAuditListItem `json:"list"`  // 记录列表
}

// KeyAuditList 加解密审计日志查询接口
func KeyAuditList(w http.ResponseWriter, r *http.Request) {
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
	middleware.Logger.Infof("分页参数: page=%d, offset=%d", page, offset)

	// 构建查询条件
	db := database.DB.
		Table(middleware.TableEncryptDecryptLog + " k").
		Select("k.id, k.user_id, k.user_name, u.nick_name, k.operation_type, k.project, " +
			"COALESCE(p.project_name, k.project) as project_name, " +
			"k.batch_count, k.status, k.ip_address, " +
			"DATE_FORMAT(k.created_at, '%Y-%m-%d %H:%i:%s') as created_at").
		Joins("LEFT JOIN " + middleware.TableUsers + " u ON u.id = k.user_id").
		Joins("LEFT JOIN " + middleware.TableProjectDict + " p ON p.project = k.project").
		Order("k.created_at DESC")

	// 超级管理员可以查看所有审计记录，普通用户只能查看自己的记录
	if role.Level > 0 {
		db = db.Where("k.user_id = ?", userInfo.UserID)
	}

	// 添加查询条件（从URL参数获取）
	if project := query.Get("project_name"); project != "" {
		db = db.Where("k.project_name = ?", project)
	}
	if userName := query.Get("nick_name"); userName != "" {
		db = db.Where("k.nick_name LIKE ?", "%"+userName+"%")
	}
	if operationType := query.Get("operation_type"); operationType != "" {
		db = db.Where("k.operation_type = ?", operationType)
	}
	if status := query.Get("status"); status != "" {
		db = db.Where("k.status = ?", status)
	}

	// 处理时间查询条件
	if startTime := query.Get("start_time"); startTime != "" {
		db = db.Where("k.created_at >= ?", startTime)
	}
	if endTime := query.Get("end_time"); endTime != "" {
		db = db.Where("k.created_at <= ?", endTime)
	}

	// 获取总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录总数失败")
		return
	}
	middleware.Logger.Infof("总记录数: total=%d", total)

	// 分页查询，使用固定的pageSize
	var logs []KeyAuditListItem
	if err := db.
		Limit(pageSize). // 使用固定的pageSize
		Offset(offset).
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询记录失败")
		return
	}
	middleware.Logger.Infof("当前页记录数: count=%d", len(logs))

	// 返回结果
	response := KeyAuditResponse{
		Total: total,
		List:  logs,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
