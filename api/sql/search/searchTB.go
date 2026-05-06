package search

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
)

// SearchTBRequest 查询表请求结构体
type SearchTBRequest struct {
	Agent  string `json:"agent"`  // 项目标识符，用于查询代理地址
	DBName string `json:"dbName"` // 数据库名称
}

// SearchTB 查询表列表
func SqlSearchTB(w http.ResponseWriter, r *http.Request) {
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
	var searchReq SearchTBRequest
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
			"tbName": nil,
		},
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/sql/structure"
	decompressedData, err := middleware.CallAgentWithBody(searchReq.Agent, pluginPath, requestData)
	if err != nil {
		http.Error(w, `{"error":"调用Agent失败: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// 解析外层响应结构
	var response struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(decompressedData, &response); err != nil {
		http.Error(w, `{"error":"解析响应失败: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// 检查响应状态
	if response.Code != 200 {
		http.Error(w, `{"error":"Agent返回错误: `+response.Msg+`"}`, http.StatusInternalServerError)
		return
	}

	tables := response.Data

	// 过滤表
	var filteredTables []string
	for _, tableName := range tables {
		// 过滤条件示例：排除系统表和备份表等
		if shouldIncludeTable(tableName) {
			filteredTables = append(filteredTables, tableName)
		}
	}

	// 返回过滤后的结果
	middleware.ResponseSuccess(w, "查询成功", map[string]interface{}{
		"tables": filteredTables,
	})
}

// shouldIncludeTable 判断表是否应该包含在结果中
func shouldIncludeTable(tableName string) bool {
	// 如果表应该被过滤，则不包含在结果中
	return !ShouldFilterTable(tableName)
}
