package view

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"database/sql"
	"net/http"
	"strconv"
)

// ViewDetail 获取监控视图详情
func ViewDetail(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取视图ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "视图ID不能为空")
		return
	}

	// 转换ID为整数
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "视图ID格式错误")
		return
	}

	// 查询视图详情
	view, err := monitor.GetViewByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ResponseError(w, http.StatusNotFound, "监控视图不存在")
		} else {
			middleware.Logger.Errorf("获取监控视图失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "获取监控视图失败")
		}
		return
	}

	// 返回成功
	middleware.ResponseSuccess(w, "获取成功", view)
}
