package view

import (
	"bytes"
	"cmdb/api/elfk/search"
	"cmdb/models/elfk"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// responseRecorder 自定义 http.ResponseWriter，用于捕获响应
type responseRecorder struct {
	status int
	body   *bytes.Buffer
	header http.Header
}

// newResponseRecorder 创建一个新的 ResponseRecorder
func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		status: http.StatusOK,
		body:   bytes.NewBuffer(nil),
		header: make(http.Header),
	}
}

// Header 实现 http.ResponseWriter 接口
func (r *responseRecorder) Header() http.Header {
	return r.header
}

// Write 实现 http.ResponseWriter 接口
func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

// WriteHeader 实现 http.ResponseWriter 接口
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

// buildViewResponse 构建标准的视图响应格式
func buildViewResponse(view elfk.ViewResponse, fields, indices interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":            view.ID,
		"name":          view.Name,
		"project":       view.Project,
		"category":      view.Category,
		"log_type":      view.LogType,
		"index_pattern": view.IndexPattern,
		"time_field":    view.TimeField,
		"time_format":   view.TimeFormat,
		"description":   view.Description,
		"update_time":   view.UpdateTime,
		"all_field":     fields,
		"indices":       indices,
	}
}

// ViewDetail 获取视图详情
func ViewDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

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

	// 获取视图ID
	viewID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的视图ID")
		return
	}

	// 先尝试从Redis获取视图信息
	viewInfo, err := GetViewFromRedis(r.Context(), uint(viewID))
	if err == nil {
		// 如果成功获取Redis缓存，直接返回
		middleware.ResponseSuccess(w, "获取视图详情成功（从缓存）", viewInfo)
		return
	}

	// 查询视图
	var view elfk.View
	if err := database.DB.Where("id = ?", viewID).First(&view).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "视图不存在")
		return
	}

	// 获取项目权限并验证
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 验证是否有权限访问该项目
	if !isSuperAdmin && !utils.Contains(projects, view.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 构建视图详情响应
	viewResponse := elfk.ViewResponse{
		ID:           view.ID,
		Name:         view.Name,
		Project:      view.Project,
		Category:     view.Category,
		LogType:      view.LogType,
		IndexPattern: view.IndexPattern,
		TimeField:    view.TimeField,
		TimeFormat:   view.TimeFormat,
		Description:  view.Description,
		UpdateTime:   view.UpdateTime,
	}

	// 创建一个临时请求，用于调用EsInfo接口获取最新字段信息
	indexBasePattern := view.IndexPattern
	// 处理可能的通配符和排除索引
	if indexBasePattern == "" {
		indexBasePattern = "*"
	}

	// 使用数据库中查询到的LogType字段
	logType := view.LogType
	if logType == "" {
		// 如果数据库中LogType为空，默认使用"elfk"
		logType = "elfk"
	}

	// 创建临时请求获取索引映射信息，添加log_type参数
	tempUrl := fmt.Sprintf("/api/elfk/search/indices?project=%s&index=%s&log_type=%s",
		view.Project, indexBasePattern, logType)
	tempReq, err := http.NewRequest(http.MethodGet, tempUrl, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建映射请求失败")
		return
	}

	// 复制原始请求的上下文，确保用户信息被正确传递
	tempReq = tempReq.WithContext(r.Context())

	// 从原始请求中复制一些头信息
	for key, values := range r.Header {
		for _, value := range values {
			tempReq.Header.Add(key, value)
		}
	}

	// 创建ResponseWriter的包装器，用于捕获EsInfo的响应
	recorder := newResponseRecorder()

	search.EsInfo(recorder, tempReq)

	// 检查EsInfo的响应状态
	if recorder.status != http.StatusOK {
		middleware.Logger.Errorf("调用EsInfo接口失败: %d - %s", recorder.status, recorder.body.String())
		// 如果获取字段失败，仍然返回基本视图信息，保持原有格式
		middleware.ResponseSuccess(w, "获取视图详情成功（字段信息获取失败）", buildViewResponse(viewResponse, nil, nil))
		return
	}

	// 解析EsInfo返回的结果
	var esInfoResult map[string]interface{}
	if err := json.Unmarshal(recorder.body.Bytes(), &esInfoResult); err != nil {
		middleware.Logger.Errorf("解析EsInfo响应失败: %v", err)
		middleware.ResponseSuccess(w, "获取视图详情成功（字段信息解析失败）", buildViewResponse(viewResponse, nil, nil))
		return
	}

	// 从EsInfo返回结果中提取字段信息
	data, ok := esInfoResult["data"].(map[string]interface{})
	if !ok {
		middleware.Logger.Errorf("EsInfo响应格式错误: 缺少data字段")
		middleware.ResponseSuccess(w, "获取视图详情成功（字段信息格式错误）", buildViewResponse(viewResponse, nil, nil))
		return
	}

	// 保持原有视图结构，仅用ES的字段信息进行补充
	// 注意：原来的结构是直接返回视图信息，现在需要确保与之前的返回格式一致
	middleware.ResponseSuccess(w, "获取视图详情成功", buildViewResponse(viewResponse, data["fields"], data["indices"]))
}
