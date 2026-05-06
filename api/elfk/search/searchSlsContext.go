package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// SearchSLSContext 处理SLS上下文查询
func SearchSLSContext(w http.ResponseWriter, bodyData []byte, startTime time.Time) {
	// 解析JSON请求
	var contextReq map[string]interface{}
	if err := json.Unmarshal(bodyData, &contextReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}

	// 获取必要参数
	project, ok := contextReq["project"].(string)
	if !ok || project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 获取SLS特有参数
	packID, ok := contextReq["pack_id"].(string)
	if !ok || packID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "pack_id参数不能为空")
		return
	}

	packMeta, ok := contextReq["pack_meta"].(string)
	if !ok || packMeta == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "pack_meta参数不能为空")
		return
	}

	logstore, ok := contextReq["index_pattern"].(string)
	if !ok || logstore == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "logstore参数不能为空")
		return
	}

	// 获取前后行数，默认各10行
	backLines, ok := contextReq["back_lines"].(float64)
	if !ok {
		backLines = 10
	}

	forwardLines, ok := contextReq["forward_lines"].(float64)
	if !ok {
		forwardLines = 10
	}

	middleware.Logger.Infof("SLS上下文查询参数: project=%s, logstore=%s, pack_id=%s, back_lines=%v, forward_lines=%v",
		project, logstore, packID, backLines, forwardLines)

	// 从全局SLS客户端管理器获取客户端
	client, err := GlobalSLSManager.GetClient(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取SLS客户端失败: %v", err))
		return
	}

	// 构建上下文查询请求
	getContextLogsRequest := &sls20201230.GetContextLogsRequest{
		PackId:       tea.String(packID),
		PackMeta:     tea.String(packMeta),
		BackLines:    tea.Int64(int64(backLines)),
		ForwardLines: tea.Int64(int64(forwardLines)),
	}

	// 准备运行时选项和请求头
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 发起上下文查询请求
	response, err := client.GetContextLogsWithOptions(
		tea.String(project),
		tea.String(logstore),
		getContextLogsRequest,
		headers,
		runtime,
	)
	if err != nil {
		middleware.Logger.Errorf("SLS上下文查询失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("SLS上下文查询失败: %v", err))
		return
	}

	// 检查响应
	if response == nil || response.Body == nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "SLS返回空响应")
		return
	}

	// 解析响应体
	bodyBytes, err := json.Marshal(response.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("序列化SLS响应失败: %v", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("解析SLS响应失败: %v", err))
		return
	}

	// 处理SLS返回的数据，转换为统一的上下文查询结果格式
	contextResult := processSlsContextResult(result)

	// 添加处理时间到结果中
	contextResult["took_ms"] = time.Since(startTime).Milliseconds()

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", contextResult)
}

// processSlsContextResult 将SLS返回的结果转换为统一的上下文查询结果格式
func processSlsContextResult(slsResult map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 从结果中提取日志
	var logs []interface{}
	if logsArray, ok := slsResult["logs"].([]interface{}); ok {
		logs = logsArray
	} else {
		logs = []interface{}{}
	}

	// 获取中心日志位置
	centerIndex := -1
	for i, log := range logs {
		if logMap, ok := log.(map[string]interface{}); ok {
			if indexNum, ok := logMap["__index_number__"].(string); ok && indexNum == "0" {
				centerIndex = i
				break
			}
		}
	}

	// 如果找不到中心日志，取中间位置
	if centerIndex == -1 && len(logs) > 0 {
		centerIndex = len(logs) / 2
	}

	// 分离前、中、后日志
	var beforeLogs []interface{}
	var centerLog interface{}
	var afterLogs []interface{}

	if centerIndex >= 0 {
		if centerIndex > 0 {
			beforeLogs = logs[:centerIndex]
		}
		if centerIndex < len(logs) {
			centerLog = logs[centerIndex]
		}
		if centerIndex+1 < len(logs) {
			afterLogs = logs[centerIndex+1:]
		}
	}

	// 设置结果
	result["before"] = convertSlsLogsToSearchHits(beforeLogs)
	result["before_total"] = len(beforeLogs)

	if centerLog != nil {
		centerHit := convertSlsLogToSearchHit(centerLog)
		result["center"] = centerHit
	} else {
		result["center"] = nil
	}

	result["after"] = convertSlsLogsToSearchHits(afterLogs)
	result["after_total"] = len(afterLogs)

	// 计算总数
	total := len(beforeLogs) + len(afterLogs)
	if centerLog != nil {
		total++
	}
	result["total"] = total

	return result
}

// convertSlsLogsToSearchHits 将SLS日志数组转换为SearchHit数组
func convertSlsLogsToSearchHits(slsLogs []interface{}) []SearchHit {
	hits := make([]SearchHit, 0, len(slsLogs))

	for _, log := range slsLogs {
		if logMap, ok := log.(map[string]interface{}); ok {
			hit := convertSlsLogToSearchHit(logMap)
			if hit != nil {
				hits = append(hits, *hit)
			}
		}
	}

	return hits
}

// convertSlsLogToSearchHit 将单个SLS日志转换为SearchHit
func convertSlsLogToSearchHit(slsLog interface{}) *SearchHit {
	logMap, ok := slsLog.(map[string]interface{})
	if !ok {
		return nil
	}

	// 尝试获取__time__作为ID
	var id string
	if timeVal, ok := logMap["__time__"]; ok {
		id = fmt.Sprintf("%v", timeVal)
	} else if packID, ok := logMap["__tag__:__pack_id__"]; ok {
		// 使用pack_id和索引号组合作为ID
		indexNum := logMap["__index_number__"]
		id = fmt.Sprintf("%v_%v", packID, indexNum)
	} else {
		// 生成随机ID
		id = fmt.Sprintf("sls-%d", time.Now().UnixNano())
	}

	// 构造SearchHit
	hit := &SearchHit{
		ID:     id,
		Index:  "sls", // 对于SLS，没有明确的索引概念，使用常量
		Score:  1.0,   // SLS没有评分概念，使用常量
		Source: logMap,
	}

	return hit
}
