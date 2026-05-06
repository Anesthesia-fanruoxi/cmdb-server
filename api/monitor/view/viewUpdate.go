package view

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"database/sql"
	"encoding/json"
	"net/http"
)

// UpdateViewRequest 更新视图请求参数
type UpdateViewRequest struct {
	ID              int64  `json:"id" validate:"required,gt=0"`
	ViewName        string `json:"view_name" validate:"required,max=100"`
	Query           string `json:"query" validate:"required"`
	Description     string `json:"description" validate:"max=255"`
	Category        string `json:"category" validate:"max=64"`
	ChartType       string `json:"chart_type" validate:"max=50"`
	RefreshInterval int    `json:"refresh_interval" validate:"gte=0"`
	TimeRange       string `json:"time_range" validate:"max=50"`
}

// ViewUpdate 更新监控视图
func ViewUpdate(w http.ResponseWriter, r *http.Request) {
	// 只接受PUT请求
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持PUT请求")
		return
	}

	// 解析请求参数
	var req UpdateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	// 简单参数校验
	if req.ID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "视图ID必须大于0")
		return
	}

	if req.ViewName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "视图名称不能为空")
		return
	}

	if req.Query == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "查询语句不能为空")
		return
	}

	// 检查视图是否存在
	existingView, err := monitor.GetViewByID(r.Context(), req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ResponseError(w, http.StatusNotFound, "监控视图不存在")
		} else {
			middleware.Logger.Errorf("获取监控视图失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "获取监控视图失败")
		}
		return
	}

	// 设置更新字段
	existingView.ViewName = req.ViewName
	existingView.Query = req.Query
	existingView.Description = req.Description
	existingView.ChartType = req.ChartType

	// 更新数据库
	if err := monitor.UpdateView(r.Context(), existingView); err != nil {
		middleware.Logger.Errorf("更新监控视图失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "更新监控视图失败")
		return
	}

	// 返回成功
	middleware.ResponseSuccess(w, "更新成功", nil)
}
