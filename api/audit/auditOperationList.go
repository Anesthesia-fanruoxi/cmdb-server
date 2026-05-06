package audit

import (
	"cmdb/models/audit"
	"cmdb/pkg/database" // 引入数据库包
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"time"
)

// AuditOperationLogList 获取操作日志列表
func AuditOperationLogList(w http.ResponseWriter, r *http.Request) {
	var query audit.OperationLogQuery // 使用新的查询参数结构体
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var logs []audit.OperationLog
	var total int64

	tx := database.DB.Model(&audit.OperationLog{}) // 添加调试信息

	// 构建查询条件
	if query.UserName != "" {
		tx = tx.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.Module != "" {
		tx = tx.Where("module = ?", query.Module)
	}
	if query.Action != "" {
		tx = tx.Where("action = ?", query.Action)
	}
	if query.Status != nil {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.StartTime != (time.Time{}) {
		tx = tx.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != (time.Time{}) {
		tx = tx.Where("created_at <= ?", query.EndTime)
	}
	// 获取总数
	tx.Count(&total)

	// 设置默认分页参数
	if query.PageSize == 0 {
		query.PageSize = 20 // 默认每页 20 条
	}
	if query.Page == 0 {
		query.Page = 1 // 默认第一页
	}

	// 计算总页数
	totalPages := total / query.PageSize
	if total%query.PageSize != 0 {
		totalPages++
	}

	offset := (query.Page - 1) * query.PageSize
	result := tx.Order("created_at DESC").Offset(int(offset)).Limit(int(query.PageSize)).Find(&logs)

	if result.Error != nil {
		http.Error(w, "查询操作日志失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := audit.OperationResponse{
		Total:      total,
		List:       logs,
		Page:       int(query.Page),     // 转换为 int
		PageSize:   int(query.PageSize), // 转换为 int
		TotalPages: totalPages,
	}
	w.Header().Set("Content-Type", "application/json")
	middleware.ResponseSuccess(w, "请求成功", response)
}
