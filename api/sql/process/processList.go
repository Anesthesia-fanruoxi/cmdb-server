package process

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// FlowListResult 流程列表结果结构体
type FlowListResult struct {
	ID           int64  `json:"id" gorm:"column:id"`
	ProjectID    int64  `json:"projectId" gorm:"column:project_id"`
	ProjectName  string `json:"projectName" gorm:"column:project_name"`
	ApplyID      int64  `json:"applyId" gorm:"column:apply_id"`
	ApplyName    string `json:"applyName" gorm:"column:apply_name"`
	ExecutorID   int64  `json:"executorId" gorm:"column:executor_id"`
	ExecutorName string `json:"executorName" gorm:"column:executor_name"`
}

// SqlProcessList 获取SQL审批流程列表
func SqlProcessList(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 获取筛选参数
	projectIDStr := r.URL.Query().Get("project_id")

	// 计算分页偏移量
	offset := (page - 1) * pageSize

	// 构建基本查询
	query := database.DB.Table(middleware.TableSQLApplyFlow + " s")

	// 添加连表查询
	query = query.Select("s.id, p.id as project_id, p.project_name AS project_name, apply.id as apply_id, apply.nick_name AS apply_name, executor.id as executor_id, executor.nick_name AS executor_name")
	query = query.Joins("LEFT JOIN " + middleware.TableProjectDict + " p ON p.id = s.project_id")
	query = query.Joins("LEFT JOIN " + middleware.TableUsers + " apply ON apply.id = s.apply_id")
	query = query.Joins("LEFT JOIN " + middleware.TableUsers + " executor ON executor.id = s.executor_id")

	// 添加筛选条件
	if projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err == nil {
			query = query.Where("s.project_id = ?", projectID)
		}
	}

	// 执行查询获取总数
	var total int64
	countQuery := database.DB.Table(middleware.TableSQLApplyFlow + " s")
	if projectIDStr != "" {
		projectID, _ := strconv.ParseInt(projectIDStr, 10, 64)
		countQuery = countQuery.Where("s.project_id = ?", projectID)
	}
	countQuery.Count(&total)

	// 查询流程列表
	var flows []FlowListResult
	result := query.Offset(offset).Limit(pageSize).Order("s.id desc").Find(&flows)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询流程列表失败: "+result.Error.Error())
		return
	}

	// 返回响应
	responseData := map[string]interface{}{
		"list":      flows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	middleware.ResponseSuccess(w, "获取流程列表成功", responseData)
}
