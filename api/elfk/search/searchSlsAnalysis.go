package search

import (
	"cmdb/api/task"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// LogQueryParams 定义
type LogAnalysisParams struct {
	Project      string
	IndexPattern string
	TimeField    string
	StartTime    string
	EndTime      string
	Keyword      string
	From         int
	Size         int
	LogType      string
	Offset       int64 // 添加SLS专用的offset
	Line         int64 // 添加SLS专用的line
}

// QuerySLSAnalysisLogs 专门为分析场景设计的SLS查询函数
// 此函数完全独立实现SLS查询，不依赖其他查询函数
func QuerySLSAnalysisLogs(params LogAnalysisParams) ([]map[string]interface{}, error) {
	// 记录查询参数
	middleware.Logger.Infof("执行SLS专用分析查询: 项目=%s, 索引=%s, 关键词='%s', 偏移量=%d, 行数=%d",
		params.Project, params.IndexPattern, params.Keyword, params.Offset, params.Line)

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

	// 构建查询条件
	query := params.Keyword
	if query == "" {
		query = "*" // 如果没有关键词，查询所有日志
	}

	// 解析日志库名
	logstore := params.IndexPattern
	if strings.HasSuffix(logstore, "*") {
		logstore = logstore[:len(logstore)-1]
	}

	// 查询日志数据
	getLogsV2Request := &sls20201230.GetLogsV2Request{
		From:    tea.Int32(int32(startTime.Unix())),
		To:      tea.Int32(int32(endTime.Unix())),
		Query:   tea.String(query),
		Line:    tea.Int64(params.Line),
		Offset:  tea.Int64(params.Offset),
		Reverse: tea.Bool(true),
	}

	// 执行查询
	runtime := &util.RuntimeOptions{}
	headers := &sls20201230.GetLogsV2Headers{}
	response, err := client.GetLogsV2WithOptions(
		tea.String(params.Project),
		tea.String(logstore),
		getLogsV2Request,
		headers,
		runtime,
	)
	if err != nil {
		return nil, fmt.Errorf("执行SLS查询请求失败: %v", err)
	}

	// 解析响应
	logMaps := make([]map[string]interface{}, 0)

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
				for _, item := range dataArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						logMaps = append(logMaps, itemMap)
					}
				}
				middleware.Logger.Infof("SLS分析查询返回日志: 获取%d条记录, 偏移量%d",
					len(dataArray), params.Offset)
			}
		}
	}

	// 统计实际返回数量
	middleware.Logger.Infof("SLS分析查询完成: 偏移量=%d, 请求行数=%d, 实际返回=%d条记录",
		params.Offset, params.Line, len(logMaps))

	return logMaps, nil
}

// analyzeSlsData 分析SLS数据
func analyzeSlsData(req AnalysisRequest, t *task.Task) (*AnalysisResult, error) {
	// 记录日志
	middleware.Logger.Infof("开始分析SLS数据, QID: %s, 字段: %s, 数据量: %d", req.Qid, req.Field, req.Count)

	// 设置总数
	t.TotalCount = req.Count
	t.ProcessedCount = 0

	// 获取查询信息
	project, _, indexPattern, timeField, _, startTime, endTime, keyword, err := GetQueryInfoByID(context.Background(), req.Qid)
	if err != nil {
		middleware.Logger.Warnf("获取查询信息失败: %v", err)
		return nil, fmt.Errorf("获取查询信息失败: %v", err)
	}

	// 计算需要执行的并行查询数量
	batchSize := 100                                        // SLS每次查询100条记录
	totalBatches := (req.Count + batchSize - 1) / batchSize // 向上取整
	maxConcurrency := 100                                   // 最大并行数

	middleware.Logger.Infof("SLS分析计划: 总批次=%d, 每批大小=%d, 最大并行=%d",
		totalBatches, batchSize, maxConcurrency)

	// 创建结果通道和错误通道
	resultChan := make(chan []map[string]interface{}, totalBatches)
	errorChan := make(chan error, totalBatches)

	// 使用WaitGroup同步并行任务
	var wg sync.WaitGroup

	// 创建信号量控制并行度
	semaphore := make(chan struct{}, maxConcurrency)

	// 启动Worker进行并行查询
	for i := 0; i < totalBatches; i++ {
		wg.Add(1)
		go func(batchIndex int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 计算当前批次的偏移量和大小
			offset := batchIndex * batchSize
			currentBatchSize := batchSize
			if offset+batchSize > req.Count {
				currentBatchSize = req.Count - offset
			}

			// 构建查询参数
			queryParams := LogAnalysisParams{
				LogType:      "sls",
				Project:      project,
				IndexPattern: indexPattern,
				TimeField:    timeField,
				StartTime:    startTime,
				EndTime:      endTime,
				Keyword:      keyword,
				Offset:       int64(offset),           // 添加SLS偏移量
				Line:         int64(currentBatchSize), // 添加SLS返回数量
			}

			middleware.Logger.Infof("执行SLS查询 #%d: 偏移=%d, 大小=%d",
				batchIndex+1, offset, currentBatchSize)

			// 执行专用查询
			logMaps, err := QuerySLSAnalysisLogs(queryParams)
			if err != nil {
				errorChan <- fmt.Errorf("SLS批次 #%d 查询失败: %v", batchIndex+1, err)
				return
			}

			middleware.Logger.Infof("SLS批次 #%d 查询完成，获取 %d 条记录",
				batchIndex+1, len(logMaps))

			// 发送结果到通道
			resultChan <- logMaps
		}(i)
	}

	// 等待所有查询完成，并关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 收集所有错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// 如果存在错误，返回第一个错误
	if len(errors) > 0 {
		middleware.Logger.Errorf("SLS查询存在 %d 个错误，第一个错误: %v", len(errors), errors[0])
		return nil, errors[0]
	}

	// 收集所有结果
	var allResults []map[string]interface{}
	for logs := range resultChan {
		allResults = append(allResults, logs...)
	}

	middleware.Logger.Infof("SLS查询完成，共获取 %d 条记录", len(allResults))

	// 分析数据
	fieldValues := make(map[string]int)
	processedCount := 0

	for _, record := range allResults {
		// 从记录中获取字段值
		value, err := getFieldValueFromSource(record, req.Field)
		if err != nil {
			continue
		}

		// 如果配置了分隔符，进一步提取内容
		if req.StartDelimiter != "" || req.EndDelimiter != "" {
			extracted, err := extractWithDelimiter(value, req.StartDelimiter, req.EndDelimiter)
			if err != nil {
				continue
			}
			value = extracted
		}

		// 统计值的出现次数
		fieldValues[value]++
		processedCount++

		// 更新任务进度（每100条更新一次）
		if processedCount%100 == 0 {
			t.ProcessedCount = processedCount
			task.GlobalTaskManager.SaveTaskStatus(t)
		}

		// 如果已达到请求的数据量限制，结束处理
		if processedCount >= req.Count {
			break
		}
	}

	// 最终更新进度
	t.ProcessedCount = processedCount
	task.GlobalTaskManager.SaveTaskStatus(t)

	middleware.Logger.Infof("SLS数据分析完成，共分析 %d 条记录，统计 %d 个不同的值",
		processedCount, len(fieldValues))

	// 构建分析结果
	result := buildAnalysisResult(fieldValues)

	return result, nil
}
