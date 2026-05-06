package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// QuerySLSLogs 查询SLS日志
func QuerySLSLogs(params LogQueryParams) (interface{}, error) {
	// 使用SLSManager获取对应项目的客户端
	client, err := GlobalSLSManager.GetClient(params.Project)
	if err != nil {
		return nil, fmt.Errorf("获取SLS客户端失败: %v", err)
	}

	// 解析时间
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", params.StartTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %v", err)
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", params.EndTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("解析结束时间失败: %v", err)
	}

	// 构建查询条件 - SLS不需要解析关键词，直接使用
	query := params.Keyword
	if query == "" {
		query = "*" // 如果没有关键词，查询所有日志
	}

	// 解析项目与日志库，通常格式为"app*"，需要提取前缀作为日志库名
	logstore := params.IndexPattern
	if strings.HasSuffix(logstore, "*") {
		logstore = logstore[:len(logstore)-1]
	}

	// 获取符合条件的总数
	project := params.Project // 直接使用传入的项目标识
	totalHits, err := CountSLSLogs(client, project, logstore, startTime.Unix(), endTime.Unix(), query)
	if err != nil {
		middleware.Logger.Warnf("获取SLS总数失败: %v", err)
		totalHits = 0
	}

	// 根据排序参数确定是否逆序查询
	reverse := true // 默认降序（逆序）
	if params.SortOrder == "asc" {
		reverse = false // 升序（正序）
	}

	// 查询日志数据
	getLogsV2Request := &sls20201230.GetLogsV2Request{
		From:    tea.Int32(int32(startTime.Unix())),
		To:      tea.Int32(int32(endTime.Unix())),
		Query:   tea.String(query + "|with_pack_meta"),
		Line:    tea.Int64(20),
		Offset:  tea.Int64(0),
		Reverse: tea.Bool(reverse),
	}

	// 执行查询
	runtime := &util.RuntimeOptions{}
	// 使用正确的GetLogsV2Headers类型
	headers := &sls20201230.GetLogsV2Headers{}
	response, err := client.GetLogsV2WithOptions(
		tea.String(project),
		tea.String(logstore),
		getLogsV2Request,
		headers,
		runtime,
	)
	if err != nil {
		return nil, fmt.Errorf("查询SLS日志失败: %v", err)
	}

	// 构建结果
	result := map[string]interface{}{
		"count": totalHits,
		"logs":  []interface{}{},
	}

	// 解析返回的日志数据
	if response != nil && response.Body != nil {
		bodyBytes, err := json.Marshal(response.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化SLS响应失败: %v", err)
		}

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			return nil, fmt.Errorf("解析SLS响应失败: %v", err)
		}

		if dataVal, ok := bodyMap["data"]; ok && dataVal != nil {
			if dataArray, ok := dataVal.([]interface{}); ok {
				result["logs"] = dataArray
			}
		}
	}

	middleware.Logger.Infof("SLS查询返回日志: 总数:%d", totalHits)
	return result, nil
}

// convertSLSResultToStandard 将SLS查询结果转换为标准格式
func convertSLSResultToStandard(slsResult interface{}) (map[string]interface{}, error) {
	// 将SLS结果转换为标准格式
	result := make(map[string]interface{})

	// 处理SLS查询结果
	slsMap, ok := slsResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("SLS结果格式错误")
	}

	// 构造与ES结果兼容的格式
	hits := make(map[string]interface{})
	hitsArray := make([]interface{}, 0)

	// 提取日志数据
	if logs, ok := slsMap["logs"].([]interface{}); ok {
		for _, log := range logs {
			// 将每条日志作为一个hit
			hit := map[string]interface{}{
				"_source": log,
			}
			hitsArray = append(hitsArray, hit)
		}
	}

	// 获取总数
	count := int64(0)
	if countValue, ok := slsMap["count"]; ok {
		switch v := countValue.(type) {
		case float64:
			count = int64(v)
		case int64:
			count = v
		case int:
			count = int64(v)
		case string:
			parsedCount, _ := strconv.ParseInt(v, 10, 64)
			count = parsedCount
		}
	}

	// 构造hits结构
	hits["hits"] = hitsArray
	hits["total"] = map[string]interface{}{
		"value": count,
	}

	// 记录转换后的总数
	middleware.Logger.Infof("SLS转换后总数: %d", count)

	// 设置最终结果
	result["hits"] = hits

	// 同时添加一个顶层total字段，兼容旧版格式处理
	result["total"] = float64(count)

	return result, nil
}
