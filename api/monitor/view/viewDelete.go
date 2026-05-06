package view

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// ViewDelete 删除监控视图
func ViewDelete(w http.ResponseWriter, r *http.Request) {
	// 只接受DELETE请求
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持DELETE请求")
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

	// 先检查视图是否存在
	_, err = monitor.GetViewByID(r.Context(), id)
	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "监控视图不存在")
		return
	}

	// 删除视图
	if err := monitor.DeleteView(r.Context(), id); err != nil {
		middleware.Logger.Errorf("删除监控视图失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "删除监控视图失败")
		return
	}

	// 返回成功
	middleware.ResponseSuccess(w, "删除成功", nil)
}
