package view

import (
	"cmdb/models/monitor"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// CreateViewRequest 创建视图请求参数
type CreateViewRequest struct {
	ViewName        string `json:"view_name" validate:"required,max=100"`
	Query           string `json:"query" validate:"required"`
	Description     string `json:"description" validate:"max=255"`
	Category        string `json:"category" validate:"max=64"`
	ChartType       string `json:"chart_type" validate:"max=50"`
	RefreshInterval int    `json:"refresh_interval" validate:"gte=0"`
	TimeRange       string `json:"time_range" validate:"max=50"`
}

// ViewCreate 创建监控视图
func ViewCreate(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 获取当前用户ID
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok || userInfo.UserID <= 0 {
		middleware.ResponseError(w, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 解析请求参数
	var req CreateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	// 验证请求参数 (使用validator已被弃用，这里简单校验)
	if req.ViewName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "视图名称不能为空")
		return
	}
	if req.Query == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "查询语句不能为空")
		return
	}

	// 设置默认值
	if req.ChartType == "" {
		req.ChartType = "line"
	}
	if req.RefreshInterval <= 0 {
		req.RefreshInterval = 60 // 默认60秒
	}
	if req.TimeRange == "" {
		req.TimeRange = "1h" // 默认1小时
	}

	// 创建视图模型
	view := &monitor.MonitorView{
		ViewName:    req.ViewName,
		Query:       req.Query,
		Description: req.Description,
		ChartType:   req.ChartType,
		CreatedBy:   userInfo.UserID,
	}

	// 保存到数据库
	if err := monitor.CreateView(r.Context(), view); err != nil {
		middleware.Logger.Errorf("创建监控视图失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "创建监控视图失败")
		return
	}

	// 返回成功
	middleware.ResponseSuccess(w, "创建监控视图成功", nil)
}
