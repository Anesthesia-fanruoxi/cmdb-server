package view

import (
	"cmdb/models/elfk"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"time"
)

// UpdateViewRequest 更新视图请求结构
type UpdateViewRequest struct {
	ID           uint   `json:"id"`            // 视图ID，必填
	Name         string `json:"name"`          // 视图名称，可选
	Category     string `json:"category"`      //所属分类
	IndexPattern string `json:"index_pattern"` // 索引模式，可选
	TimeField    string `json:"time_field"`    // 时间字段，可选
	TimeFormat   string `json:"time_format"`   // 时间格式，可选
	Description  string `json:"description"`   // 视图描述，可选
}

// UpdateView 更新视图
func UpdateView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 声明变量
	var err error
	var projects []string
	var isSuperAdmin bool

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 解析请求数据
	var req UpdateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 验证必填字段
	if req.ID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "视图ID不能为空")
		return
	}

	// 查询原有视图
	var view elfk.View
	if err := database.DB.Where("id = ?", req.ID).First(&view).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "视图不存在")
		return
	}

	// 获取项目权限并验证
	projects, isSuperAdmin, err = utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 验证是否有权限访问该项目
	if !isSuperAdmin && !utils.Contains(projects, view.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 更新字段（只更新非空字段）
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.IndexPattern != "" {
		updates["index_pattern"] = req.IndexPattern
	}
	if req.TimeField != "" {
		updates["time_field"] = req.TimeField
	}
	if req.TimeFormat != "" {
		updates["time_format"] = req.TimeFormat
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 添加更新时间
	updates["update_time"] = time.Now()

	// 更新视图
	if err := database.DB.Model(&view).Updates(updates).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新视图失败")
		return
	}

	// 重新查询更新后的视图
	if err := database.DB.First(&view, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的视图失败")
		return
	}
	err = UpdateViewRedisCache(view.ID)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "视图缓存写入redis失败")
		return
	}
	middleware.ResponseSuccess(w, "更新视图成功", view)
}
