package search

import (
	"cmdb/models/sql"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// SearchDB 查询数据库列表
func SqlSearchDB(w http.ResponseWriter, r *http.Request) {
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
	var searchReq sql.SearchDBRequest
	if err := json.Unmarshal(body, &searchReq); err != nil {
		http.Error(w, `{"error":"解析JSON失败"}`, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if searchReq.Agent == "" {
		http.Error(w, `{"error":"agent字段不能为空"}`, http.StatusBadRequest)
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
		"type": "db",
		"op": map[string]interface{}{
			"dbName": nil,
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
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Databases []string    `json:"databases"`
			Metadata  interface{} `json:"metadata"`
		} `json:"data"`
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

	databases := response.Data.Databases

	// 过滤掉information_schema数据库
	filteredDatabases := make([]string, 0, len(databases))

	// 过滤系统数据库
	for _, db := range databases {
		// 检查是否是系统数据库
		isSysDB := false
		for _, sysDB := range SystemDatabases {
			if strings.EqualFold(db, sysDB) {
				isSysDB = true
				break
			}
		}

		// 如果不是系统数据库，则添加到允许列表
		if !isSysDB {
			filteredDatabases = append(filteredDatabases, db)
		}
	}

	filteredMetadata := filterDBMetadata(response.Data.Metadata, filteredDatabases)

	// 返回过滤后的结果，包含元数据
	middleware.ResponseSuccess(w, "查询成功", map[string]interface{}{
		"databases": filteredDatabases,
		"metadata":  filteredMetadata,
	})
}

func filterDBMetadata(metadata interface{}, allowedDatabases []string) interface{} {
	if metadata == nil {
		return nil
	}

	allowed := make(map[string]struct{}, len(allowedDatabases))
	for _, db := range allowedDatabases {
		allowed[db] = struct{}{}
	}

	metaMap, ok := metadata.(map[string]interface{})
	if !ok {
		return metadata
	}

	dbsRaw, ok := metaMap["databases"].([]interface{})
	if !ok {
		return metadata
	}

	newDBs := make([]interface{}, 0, len(dbsRaw))
	for _, dbItem := range dbsRaw {
		dbMap, ok := dbItem.(map[string]interface{})
		if !ok {
			continue
		}

		dbName, _ := dbMap["db_name"].(string)
		if dbName == "" {
			continue
		}
		if _, exists := allowed[dbName]; !exists {
			continue
		}

		if tablesRaw, ok := dbMap["tables"].([]interface{}); ok {
			newTables := make([]interface{}, 0, len(tablesRaw))
			for _, tableItem := range tablesRaw {
				tableMap, ok := tableItem.(map[string]interface{})
				if !ok {
					continue
				}
				tableName, _ := tableMap["name"].(string)
				if tableName != "" && ShouldFilterTable(tableName) {
					continue
				}
				newTables = append(newTables, tableMap)
			}
			dbMap["tables"] = newTables
		}

		newDBs = append(newDBs, dbMap)
	}

	metaMap["databases"] = newDBs
	return metaMap
}
