package search

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
)

// SearchTBDetailRequest 查询表详情请求结构体
type SearchTBDetailRequest struct {
	Agent  string `json:"agent"`  // 项目标识符，用于查询代理地址
	DBName string `json:"dbName"` // 数据库名称
	TBName string `json:"tbName"` // 表名称
}

// SearchTBDetail 查询表详情
func SqlSearchTBDetail(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"只支持POST请求"}`, http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"读取请求体失败"}`, http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// 解析JSON
	var searchReq SearchTBDetailRequest
	if err := json.Unmarshal(body, &searchReq); err != nil {
		http.Error(w, `{"error":"解析JSON失败"}`, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if searchReq.Agent == "" {
		http.Error(w, `{"error":"agent字段不能为空"}`, http.StatusBadRequest)
		return
	}
	if searchReq.DBName == "" {
		http.Error(w, `{"error":"dbName字段不能为空"}`, http.StatusBadRequest)
		return
	}
	if searchReq.TBName == "" {
		http.Error(w, `{"error":"tableName字段不能为空"}`, http.StatusBadRequest)
		return
	}

	// 检查表名是否应该被过滤
	if ShouldFilterTable(searchReq.TBName) {
		// 返回空结果而不是错误，以免暴露被过滤的表名
		middleware.ResponseSuccess(w, "查询成功", map[string]interface{}{})
		return
	}

	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果不是超级管理员且没有可访问的项目，返回无权限
	if !isSuperAdmin && len(projects) == 0 {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 验证当前请求的项目是否在可访问列表中
	if !isSuperAdmin && !utils.Contains(projects, searchReq.Agent) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 构建请求体
	requestData := map[string]interface{}{
		"type": "tb",
		"op": map[string]interface{}{
			"dbName": searchReq.DBName,
			"tbName": searchReq.TBName,
		},
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/sql/structure"
	decompressedData, err := middleware.CallAgentWithBody(searchReq.Agent, pluginPath, requestData)
	if err != nil {
		http.Error(w, `{"error":"调用Agent失败: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// 解析解密后的JSON数据
	var tableDetail map[string]interface{}
	if err := json.Unmarshal(decompressedData, &tableDetail); err != nil {
		http.Error(w, `{"error":"解析响应失败: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// 插件返回的数据已经包含 {code, message, data} 结构
	// 提取实际的数据部分，避免双层嵌套
	actualData, ok := tableDetail["data"]
	if !ok {
		// 如果没有data字段，直接返回整个结构（兼容旧版本）
		middleware.ResponseSuccess(w, "查询成功", tableDetail)
		return
	}

	// 返回提取后的数据
	middleware.ResponseSuccess(w, "查询成功", actualData)
}
