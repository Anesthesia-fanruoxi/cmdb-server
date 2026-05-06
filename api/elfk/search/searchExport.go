/*
ES数据导出模块 (searchExport.go)

该模块提供了Elasticsearch数据的导出功能，主要包含以下功能：
1. 滚动查询功能 - 支持查询ES中超过10000条的数据（绕过默认限制）
2. 数据扁平化处理 - 将嵌套的JSON结构展平为单层结构
3. Excel导出功能 - 将数据导出为Excel文件

API说明：
1. 滚动查询测试接口 - /api/elfk/search/scroll/test
   请求方法: POST
   请求参数:
   {
     "project": "项目标识",
     "index_pattern": "索引名称",
     "start_time": "开始时间",
     "end_time": "结束时间",
     "time_field": "时间字段",
     "keyword": "搜索关键词",
     "size": 批次大小
   }
   返回:
   {
     "code": 0,
     "message": "查询成功",
     "data": {
       "total_count": 总记录数,
       "query_time": 查询耗时(毫秒)
     }
   }

2. 数据导出接口 - /api/elfk/search/export
   请求方法: POST
   请求参数: 与滚动查询接口相同
   返回:
   {
     "code": 0,
     "message": "导出成功",
     "data": {
       "file_path": "文件路径",
       "file_name": "文件名",
       "export_time": 导出耗时(毫秒)
     }
   }

注意事项:
1. 对于大数据量查询，建议适当增加批次大小(size)以提高性能
2. 导出文件存储在服务器的tmp目录下，需要单独实现文件下载接口
3. 所有操作会记录审计日志
*/

package search

import (
	"bytes"
	"cmdb/api/task"
	"cmdb/models/elfk"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"time"
)

// 滚动查询功能 - 返回总数据条数
func ScrollQuery(req *elfk.ScrollRequest) (int, error) {
	totalCount := 0
	batchCount := 0
	maxRetries := 3

	scrollID, hits, err := initScrollSearch(req)
	if err != nil {
		return 0, fmt.Errorf("初始化滚动查询失败: %v", err)
	}

	defer func() {
		if scrollID != "" {
			clearScroll(scrollID, req.Project)
		}
	}()

	batchSize := len(hits)
	totalCount += batchSize
	batchCount++

	retryCount := 0

	for {
		if batchSize == 0 {
			break
		}

		var scrollErr error
		var newHits []interface{}
		var newScrollID string

		for retry := 0; retry < maxRetries; retry++ {
			newHits, newScrollID, scrollErr = continueScroll(scrollID, req.Project)
			if scrollErr == nil {
				retryCount = 0
				break
			}

			if retry < maxRetries-1 {
				middleware.Logger.Warnf("滚动查询失败(尝试 %d/%d): %v", retry+1, maxRetries, scrollErr)
				time.Sleep(time.Second * time.Duration(retry+1))
			}
		}

		if scrollErr != nil {
			retryCount++

			if retryCount >= maxRetries {
				middleware.Logger.Errorf("连续 %d 次查询失败，停止查询: %v", retryCount, scrollErr)
				return totalCount, fmt.Errorf("部分查询失败: %v", scrollErr)
			}

			continue
		}

		if newScrollID != "" {
			scrollID = newScrollID
		}

		hits = newHits
		batchSize = len(hits)
		totalCount += batchSize
		batchCount++

		if batchCount%50 == 0 {
			middleware.Logger.Infof("已处理 %d 批数据，累计: %d 条记录", batchCount, totalCount)
		}
	}

	return totalCount, nil
}

// initScrollSearch 初始化滚动查询，返回scrollID和首批结果
func initScrollSearch(req *elfk.ScrollRequest) (string, []interface{}, error) {
	scrollID, hits, _, err := initScrollSearchWithTotal(req)
	return scrollID, hits, err
}

// initScrollSearchWithTotal 初始化滚动查询，返回scrollID、首批结果和总命中数
func initScrollSearchWithTotal(req *elfk.ScrollRequest) (string, []interface{}, int, error) {
	// 验证参数
	if req.Project == "" {
		return "", nil, 0, fmt.Errorf("项目标识不能为空")
	}
	if req.IndexPattern == "" {
		return "", nil, 0, fmt.Errorf("索引名称不能为空")
	}
	if req.Size <= 0 {
		req.Size = 1000 // 设置默认批次大小
	}

	// 设置默认时间字段
	if req.TimeField == "" {
		req.TimeField = "@timestamp"
	}
	if req.TimeFormat == "" {
		req.TimeFormat = "iso8601"
	}

	// 构建初始化滚动查询的请求体（与普通查询格式一致，由Agent构建ES查询）
	reqMap := map[string]interface{}{
		"action":      "init",
		"index":       req.IndexPattern,
		"size":        req.Size,
		"scroll_time": "1m",
	}

	// 只添加非空的时间范围字段
	if req.StartTime != "" {
		reqMap["start_time"] = req.StartTime
	}
	if req.EndTime != "" {
		reqMap["end_time"] = req.EndTime
	}
	if req.TimeField != "" {
		reqMap["time_field"] = req.TimeField
	}

	// keyword为必传字段，如果为空则设置为*
	keyword := req.Keyword
	if keyword == "" {
		keyword = "*"
	}
	reqMap["keyword"] = keyword

	// 如果有自定义查询，添加query字段（必须非空）
	if req.Query != nil && len(req.Query) > 0 {
		reqMap["query"] = req.Query
		middleware.Logger.Infof("使用自定义查询: %v", req.Query)
	}

	// 如果有字段过滤
	if req.UseFieldFilter && len(req.IncludeFields) > 0 {
		reqMap["_source"] = req.IncludeFields
		middleware.Logger.Infof("启用字段过滤，包含字段: %v", req.IncludeFields)
	}

	queryBody, _ := json.Marshal(reqMap)
	middleware.Logger.Infof("发送给Agent的滚动查询请求(明文): %s", string(queryBody))

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/es-plugs/api/elfk/scroll"
	decompressedData, err := middleware.CallAgentWithBody(req.Project, pluginPath, reqMap)
	if err != nil {
		return "", nil, 0, fmt.Errorf("调用Agent失败: %w", err)
	}

	// 解析解密后的JSON数据（Agent新格式）
	var agentResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ScrollID   string        `json:"scroll_id"`
			QueryTime  int           `json:"query_time"`
			TotalHits  int           `json:"total_hits"`
			ActualHits int           `json:"actual_hits"`
			Hits       []interface{} `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(decompressedData, &agentResponse); err != nil {
		middleware.Logger.Errorf("解析解密后的数据失败: %v", err)
		return "", nil, 0, fmt.Errorf("解析解密后的数据失败: %v", err)
	}

	// 检查Agent返回的状态码
	if agentResponse.Code != 200 {
		return "", nil, 0, fmt.Errorf("Agent滚动查询失败: %s", agentResponse.Message)
	}

	middleware.Logger.Infof("初始化滚动查询成功: scroll_id=%s, total_hits=%d, actual_hits=%d",
		agentResponse.Data.ScrollID, agentResponse.Data.TotalHits, agentResponse.Data.ActualHits)

	return agentResponse.Data.ScrollID, agentResponse.Data.Hits, agentResponse.Data.TotalHits, nil
}

// IsScrollIDValid 检查滚动ID是否有效
func IsScrollIDValid(scrollID string) bool {
	// 简单检查：非空且长度合理
	return scrollID != "" && len(scrollID) > 10
}

// continueScroll 继续滚动查询，返回下一批结果和新的scrollID
func continueScroll(scrollID string, project string) ([]interface{}, string, error) {
	if scrollID == "" {
		return nil, "", fmt.Errorf("scroll_id不能为空")
	}

	if !IsScrollIDValid(scrollID) {
		return nil, "", fmt.Errorf("无效的滚动ID")
	}

	reqMap := map[string]interface{}{
		"action":      "continue",
		"scroll_id":   scrollID,
		"scroll_time": "1m",
	}

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/es-plugs/api/elfk/scroll"
	decompressedData, err := middleware.CallAgentWithBody(project, pluginPath, reqMap)
	if err != nil {
		return nil, "", fmt.Errorf("调用Agent失败: %w", err)
	}

	// 解析解密后的JSON数据（Agent新格式）
	var agentResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ScrollID   string        `json:"scroll_id"`
			QueryTime  int           `json:"query_time"`
			TotalHits  int           `json:"total_hits"`
			ActualHits int           `json:"actual_hits"`
			Hits       []interface{} `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(decompressedData, &agentResponse); err != nil {
		return nil, "", fmt.Errorf("解析解密后的数据失败: %v", err)
	}

	// 检查Agent返回的状态码
	if agentResponse.Code != 200 {
		return nil, "", fmt.Errorf("Agent滚动查询失败: %s", agentResponse.Message)
	}

	middleware.Logger.Infof("继续滚动查询: scroll_id=%s, actual_hits=%d",
		agentResponse.Data.ScrollID, agentResponse.Data.ActualHits)

	return agentResponse.Data.Hits, agentResponse.Data.ScrollID, nil
}

// clearScroll 清理滚动查询上下文，释放资源
func clearScroll(scrollID string, project string) error {
	if scrollID == "" {
		return nil
	}

	reqMap := map[string]interface{}{
		"action":    "clear",
		"scroll_id": scrollID,
	}

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/es-plugs/api/elfk/scroll"
	_, err := middleware.CallAgentWithBody(project, pluginPath, reqMap)
	if err != nil {
		return fmt.Errorf("调用Agent失败: %w", err)
	}

	return nil
}

// 在完成滚动查询功能后，将实现导出到Excel的功能

// EsScrollData 处理测试滚动查询的HTTP请求
func EsScrollData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req elfk.ScrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.IndexPattern == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "索引名称不能为空")
		return
	}
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行滚动查询测试
	totalCount, err := ScrollQuery(&req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("滚动查询测试失败: %v", err))
		return
	}

	// 计算查询耗时
	queryTime := time.Since(startTime).Milliseconds()

	// 构建返回结果
	result := map[string]interface{}{
		"index_pattern": req.IndexPattern,
		"project":       req.Project,
		"keyword":       req.Keyword,
		"total_count":   totalCount,
		"query_time_ms": queryTime,
		"batch_size":    req.Size,
		"time_field":    req.TimeField,
		"start_time":    req.StartTime,
		"end_time":      req.EndTime,
		"message":       fmt.Sprintf("滚动查询测试成功，共查询到 %d 条记录", totalCount),
	}

	// 记录审计日志（查询操作）
	RecordESSearchAudit(
		r,
		req.QueryId,
		req.Project,
		req.ViewName,
		req.IndexPattern,
		req.TimeField,
		req.StartTime,
		req.EndTime,
		req.Keyword,
		queryTime,
		totalCount,
		"查询测试",
		1,
		http.StatusOK,
		nil,
	)

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}

// EsExportDate 处理导出请求 - 创建异步导出任务
// prefix为字段前缀，用于构建嵌套字段的完整路径
func flattenMap(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range data {
		// 构建当前字段的完整路径
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		// 根据值的类型进行处理
		switch value := v.(type) {
		case map[string]interface{}:
			// 递归处理嵌套的map
			nestedMap := flattenMap(value, key)
			for nk, nv := range nestedMap {
				result[nk] = nv
			}
		case []interface{}:
			// 检查数组元素是否为map
			allMaps := true
			for _, item := range value {
				if _, ok := item.(map[string]interface{}); !ok {
					allMaps = false
					break
				}
			}

			if allMaps && len(value) > 0 {
				// 数组中元素为map，扁平化处理每个map元素（添加索引）
				for i, item := range value {
					if mapItem, ok := item.(map[string]interface{}); ok {
						indexed := fmt.Sprintf("%s[%d]", key, i)
						nestedMap := flattenMap(mapItem, indexed)
						for nk, nv := range nestedMap {
							result[nk] = nv
						}
					}
				}
			} else {
				// 普通数组，直接保存
				result[key] = value
			}
		default:
			// 基本类型，直接保存
			result[key] = value
		}
	}

	return result
}

// EsExportDate 处理导出请求 - 创建异步导出任务
func EsExportDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求数据
	var req elfk.ScrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.IndexPattern == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "索引名称不能为空")
		return
	}
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 默认启用字段过滤
	if !req.UseFieldFilter {
		req.UseFieldFilter = true
	}
	if len(req.IncludeFields) == 0 {
		req.IncludeFields = []string{"@timestamp", "message", "level", "content"}
	}

	// 构建任务参数
	taskParams, err := json.Marshal(map[string]interface{}{
		"project":        req.Project,
		"view_name":      req.ViewName,
		"index_pattern":  req.IndexPattern,
		"time_field":     req.TimeField,
		"start_time":     req.StartTime,
		"end_time":       req.EndTime,
		"keyword":        req.Keyword,
		"size":           req.Size,
		"include_fields": req.IncludeFields,
	})
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化参数失败")
		return
	}

	// 创建异步导出任务
	exportTask, err := task.GlobalTaskManager.CreateTask(
		"es_export",
		json.RawMessage(taskParams),
		userInfo.UserName,
		userInfo.UserID,
	)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("创建导出任务失败: %v", err))
		return
	}

	// 记录审计日志
	RecordESSearchAudit(r, req.QueryId, req.Project, req.ViewName, req.IndexPattern,
		req.TimeField, req.StartTime, req.EndTime, req.Keyword, 0, 0, OperationExport, 1, http.StatusOK, nil)

	middleware.ResponseSuccess(w, "导出任务已创建", map[string]interface{}{
		"task_id": exportTask.ID,
		"message": "导出任务已创建，请在任务中心查看进度",
	})
}

// ExportToExcelWithProgress 带进度回调的导出函数，返回Excel数据、总数和前100条预览数据
func ExportToExcelWithProgress(req *elfk.ScrollRequest, progressCallback func(processed, total int)) ([]byte, int, []map[string]interface{}, error) {
	totalExportedCount := 0
	batchCount := 0
	maxRetries := 3
	previewLimit := 100
	previewData := make([]map[string]interface{}, 0, previewLimit)

	scrollID, hits, totalHits, err := initScrollSearchWithTotal(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("初始化滚动查询失败: %v", err)
	}

	// 记录总数用于进度计算
	middleware.Logger.Infof("ES查询总命中数: %d 条记录", totalHits)

	defer func() {
		if scrollID != "" {
			clearScroll(scrollID, req.Project)
		}
	}()

	batchSize := len(hits)
	batchCount++

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	headers := []string{}
	headerMap := make(map[string]int)

	// 提取表头
	if batchSize > 0 {
		for _, hit := range hits {
			hitMap, ok := hit.(map[string]interface{})
			if !ok {
				continue
			}
			source, ok := hitMap["_source"].(map[string]interface{})
			if !ok {
				continue
			}
			flattenedData := flattenMap(source, "")
			for key := range flattenedData {
				if _, exists := headerMap[key]; !exists {
					headerMap[key] = len(headers)
					headers = append(headers, key)
				}
			}
		}
	}

	// 写入表头
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, header)
	}

	rowIndex := 2

	// 处理批次数据
	processAndWriteBatch := func(hits []interface{}) {
		for _, hit := range hits {
			hitMap, ok := hit.(map[string]interface{})
			if !ok {
				continue
			}
			source, ok := hitMap["_source"].(map[string]interface{})
			if !ok {
				continue
			}
			flattenedData := flattenMap(source, "")

			// 收集前100条作为预览数据
			if len(previewData) < previewLimit {
				previewData = append(previewData, flattenedData)
			}

			for colIndex, header := range headers {
				cellName, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex)
				if value, exists := flattenedData[header]; exists {
					switch v := value.(type) {
					case []interface{}, map[string]interface{}:
						jsonBytes, _ := json.Marshal(v)
						f.SetCellValue(sheetName, cellName, string(jsonBytes))
					default:
						f.SetCellValue(sheetName, cellName, v)
					}
				}
			}
			rowIndex++
			totalExportedCount++
		}
		// 回调更新进度
		if progressCallback != nil {
			progressCallback(totalExportedCount, totalHits)
		}
	}

	processAndWriteBatch(hits)

	// 继续滚动查询
	for batchSize > 0 {
		var scrollErr error
		for retry := 0; retry < maxRetries; retry++ {
			hits, scrollID, scrollErr = continueScroll(scrollID, req.Project)
			if scrollErr == nil {
				break
			}
			if retry < maxRetries-1 {
				time.Sleep(time.Second * time.Duration(retry+1))
			}
		}

		if scrollErr != nil {
			break
		}

		batchSize = len(hits)
		batchCount++

		if batchSize > 0 {
			processAndWriteBatch(hits)
		}
	}

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return nil, totalExportedCount, nil, fmt.Errorf("写入Excel失败: %v", err)
	}

	return buffer.Bytes(), totalExportedCount, previewData, nil
}
