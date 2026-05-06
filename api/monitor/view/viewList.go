package view

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// ViewList 获取监控视图列表
func ViewList(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 解析分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	category := r.URL.Query().Get("category")

	// 默认值处理
	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			pageSize = 10
		}
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大每页数量
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询视图列表
	viewList, err := monitor.ListViews(r.Context(), offset, pageSize, category)
	if err != nil {
		middleware.Logger.Errorf("获取监控视图列表失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "获取监控视图列表失败")
		return
	}

	// 构造响应数据
	result := map[string]interface{}{
		"total": viewList.Total,
		"items": viewList.Items,
		"page":  page,
		"size":  pageSize,
	}

	// 返回成功
	middleware.ResponseSuccess(w, "获取成功", result)
}
