package audit

import (
	"cmdb/models/audit"
	"cmdb/pkg/database" // 引入数据库包
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"time"
)

// AuditLoginLogList 获取操作日志列表
func AuditLoginLogList(w http.ResponseWriter, r *http.Request) {
	var query audit.LoginLogQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}

	var logs []audit.LoginLog
	var total int64

	// 关联用户表查询，获取nickname
	tx := database.DB.Table(middleware.TableLoginLog).
		Select(middleware.TableLoginLog + ".*, " + middleware.TableUsers + ".nick_name as user_name").
		Joins("LEFT JOIN " + middleware.TableUsers + " ON " + middleware.TableLoginLog + ".user_id = " + middleware.TableUsers + ".id")

	// 构建查询条件
	if query.UserName != "" {
		// 通过nickname模糊查询
		tx = tx.Where(middleware.TableUsers+".nick_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.Status != nil {
		tx = tx.Where(middleware.TableLoginLog+".status = ?", query.Status)
	}
	if query.StartTime != (time.Time{}) {
		tx = tx.Where(middleware.TableLoginLog+".created_at >= ?", query.StartTime)
	}
	if query.EndTime != (time.Time{}) {
		tx = tx.Where(middleware.TableLoginLog+".created_at <= ?", query.EndTime)
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

	// 执行查询
	offset := (query.Page - 1) * query.PageSize
	result := tx.Order(middleware.TableLoginLog + ".created_at DESC").
		Offset(int(offset)).
		Limit(int(query.PageSize)).
		Scan(&logs)

	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询操作日志失败: "+result.Error.Error())
		return
	}

	// 构建响应
	response := audit.LoginResponse{
		Total:      total,
		List:       logs, // 直接使用查询结果
		Page:       int(query.Page),
		PageSize:   int(query.PageSize),
		TotalPages: totalPages,
	}

	middleware.ResponseSuccess(w, "请求成功", response)
}
