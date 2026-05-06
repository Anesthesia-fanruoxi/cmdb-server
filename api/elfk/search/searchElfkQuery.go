package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
)

// QueryELFKLogs 查询ELFK日志
func QueryELFKLogs(params LogQueryParams) (map[string]interface{}, error) {
	// 构建发送到Agent的请求体（只添加非空字段）
	agentReqBody := map[string]interface{}{
		"index": params.IndexPattern,
	}

	// 只添加非空的时间范围字段
	if params.StartTime != "" {
		agentReqBody["start_time"] = params.StartTime
	}
	if params.EndTime != "" {
		agentReqBody["end_time"] = params.EndTime
	}
	if params.TimeField != "" {
		agentReqBody["time_field"] = params.TimeField
	}

	// keyword为必传字段，如果为空则设置为*
	keyword := params.Keyword
	if keyword == "" {
		keyword = "*"
	}
	agentReqBody["keyword"] = keyword

	// 添加排序参数，如果为空则默认降序
	sortOrder := params.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	agentReqBody["sort_order"] = sortOrder

	// 将请求转换为JSON用于日志
	reqJSON, _ := json.Marshal(agentReqBody)
	middleware.Logger.Infof("发送到Agent的查询请求(明文): %s", reqJSON)

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/es-plugs/api/elfk/search"
	decompressedData, err := middleware.CallAgentWithBody(params.Project, pluginPath, agentReqBody)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %w", err)
	}

	middleware.Logger.Infof("数据解密解压成功，解压后长度: %d", len(decompressedData))

	// 解析解密后的JSON数据（Agent返回的格式）
	var agentResponse struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(decompressedData, &agentResponse); err != nil {
		return nil, fmt.Errorf("解析解密后的数据失败: %v", err)
	}

	// 检查Agent返回的状态码
	if agentResponse.Code != 200 {
		return nil, fmt.Errorf("Agent查询失败: %s", agentResponse.Message)
	}

	// 提取查询结果统计信息
	totalHits := 0
	actualHits := 0
	if th, ok := agentResponse.Data["total_hits"].(float64); ok {
		totalHits = int(th)
	}
	if ah, ok := agentResponse.Data["actual_hits"].(float64); ok {
		actualHits = int(ah)
	}

	middleware.Logger.Infof("查询结果：总匹配 %d 条，实际返回 %d 条", totalHits, actualHits)

	// 提取hits数组
	var hitsArray []interface{}
	if hits, ok := agentResponse.Data["hits"].([]interface{}); ok {
		hitsArray = hits
		middleware.Logger.Infof("成功提取hits数组，长度: %d", len(hitsArray))
	} else {
		middleware.Logger.Warnf("无法提取hits数组，类型: %T", agentResponse.Data["hits"])
		hitsArray = []interface{}{}
	}

	// 将Agent格式转换为ES标准格式，保持向后兼容
	esFormat := map[string]interface{}{
		"took":      agentResponse.Data["query_time"],
		"timed_out": agentResponse.Data["timed_out"],
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    totalHits,
				"relation": "eq",
			},
			"max_score": nil,
			"hits":      hitsArray,
		},
	}

	return esFormat, nil
}
