package search

import (
	"cmdb/api/task"
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// analyzeElfkData 分析ELFK数据
func analyzeElfkData(req AnalysisRequest, t *task.Task) (*AnalysisResult, error) {
	// 记录日志
	middleware.Logger.Infof("开始分析ELFK数据, QID: %s, 字段: %s, 数据量: %d", req.Qid, req.Field, req.Count)

	// 设置总数
	t.TotalCount = req.Count
	t.ProcessedCount = 0

	// 获取查询信息
	project, _, indexPattern, timeField, timeFormat, startTime, endTime, keyword, err := GetQueryInfoByID(context.Background(), req.Qid)
	if err != nil {
		middleware.Logger.Warnf("获取查询信息失败: %v", err)
		// 尝试从Redis缓存中获取部分信息
		if cachedResult, found := GetESResultFromCacheByID(context.Background(), database.RDB, req.Qid); found {
			project = cachedResult.Project
			indexPattern = cachedResult.IndexName
		} else {
			return nil, fmt.Errorf("获取查询信息失败: %v", err)
		}
	}

	// 创建Redis临时key用于流式统计
	tempKey := fmt.Sprintf("analysis:%s:%d", req.Qid, time.Now().Unix())
	ctx := context.Background()

	// 设置key过期时间（防止程序崩溃导致key残留）
	database.RDB.Expire(ctx, tempKey, 1*time.Hour)

	// 确保函数结束时清理临时key
	defer func() {
		database.RDB.Del(ctx, tempKey)
		middleware.Logger.Infof("清理分析临时key: %s", tempKey)
	}()

	var processedHits int // 实际处理的记录数

	// 根据请求的数据量决定查询方式
	if req.Count <= 100 {
		// 数据量小于等于2000，直接从Redis缓存获取
		middleware.Logger.Infof("数据量 %d <= 2000, 使用Redis缓存", req.Count)
		cachedResult, found := GetESResultFromCacheByID(context.Background(), database.RDB, req.Qid)
		if !found {
			return nil, fmt.Errorf("Redis缓存未命中, QID: %s", req.Qid)
		}

		hits := cachedResult.Hits
		// 限制处理数量
		if len(hits) > req.Count {
			hits = hits[:req.Count]
		}
		processedHits = len(hits)
		middleware.Logger.Infof("从Redis获取到 %d 条记录", processedHits)

		// 更新任务进度
		t.ProcessedCount = processedHits
		task.GlobalTaskManager.SaveTaskStatus(t)

		// 转换类型并使用Redis Pipeline批量处理数据
		interfaceHits := make([]interface{}, len(hits))
		for i, hit := range hits {
			interfaceHits[i] = hit
		}
		err := processBatchToRedis(interfaceHits, req, tempKey, ctx)
		if err != nil {
			return nil, fmt.Errorf("处理缓存数据失败: %v", err)
		}
	} else {
		// 数据量大于2000，使用ES滚动查询
		middleware.Logger.Infof("数据量 %d > 2000, 使用ES滚动查询", req.Count)

		// 构建滚动查询参数
		// 构建需要的字段列表，包含时间字段和分析字段
		includeFields := []string{timeField, req.Field} // 时间字段和分析字段

		scrollReq := &elfk.ScrollRequest{
			Project:        project,
			IndexPattern:   indexPattern,
			TimeField:      timeField,
			TimeFormat:     timeFormat, // 使用从视图配置中获取的时间格式
			StartTime:      startTime,
			EndTime:        endTime,
			Keyword:        keyword,
			Size:           2000,          // 每批2000条记录
			UseFieldFilter: true,          // 启用字段过滤
			IncludeFields:  includeFields, // 包含时间字段和分析字段
		}

		totalCount := 0
		batchCount := 0
		maxRetries := 3

		// 初始化滚动查询
		scrollID, hits, err := initScrollSearch(scrollReq)
		if err != nil {
			return nil, fmt.Errorf("初始化滚动查询失败: %v", err)
		}

		// 确保在函数结束时清理滚动查询上下文
		defer func() {
			if scrollID != "" {
				err := clearScroll(scrollID, project)
				if err != nil {
					return
				}
			}
		}()

		// 处理首批数据
		batchSize := len(hits)
		batchCount++

		// 处理批次数据的函数 - 使用Redis流式处理
		processBatch := func(hits []interface{}, batchNum int) error {
			batchHits := len(hits)
			middleware.Logger.Infof("开始处理第 %d 批数据，批次大小: %d", batchNum, batchHits)

			err := processBatchToRedis(hits, req, tempKey, ctx)
			if err != nil {
				return err
			}

			// 更新计数
			for range hits {
				totalCount++
				// 如果已达到请求的数据量限制，提前结束
				if totalCount >= req.Count {
					return fmt.Errorf("已达到请求的数据量限制")
				}
			}

			middleware.Logger.Infof("第 %d 批处理完成，累计: %d/%d 条", batchNum, totalCount, req.Count)

			// 更新任务进度
			t.ProcessedCount = totalCount
			task.GlobalTaskManager.SaveTaskStatus(t)

			return nil
		}

		// 处理首批数据
		if err := processBatch(hits, batchCount); err != nil {
			if err.Error() != "已达到请求的数据量限制" {
				middleware.Logger.Warnf("处理批次数据失败: %v", err)
			} else {
				middleware.Logger.Infof("已达到请求的数据量限制: %d", req.Count)
				// 提前结束时也从Redis获取结果
				result, err := getAnalysisResultFromRedis(tempKey, ctx)
				if err != nil {
					return nil, fmt.Errorf("获取分析结果失败: %v", err)
				}
				return result, nil
			}
		}

		// 处理后续批次
		for {
			if batchSize == 0 {
				break
			}

			var scrollErr error
			var newHits []interface{}
			var newScrollID string

			// 添加重试机制
			for retry := 0; retry < maxRetries; retry++ {
				newHits, newScrollID, scrollErr = continueScroll(scrollID, project)
				if scrollErr == nil {
					break
				}

				if retry < maxRetries-1 {
					middleware.Logger.Warnf("滚动查询失败(尝试 %d/%d): %v", retry+1, maxRetries, scrollErr)
					time.Sleep(time.Second * time.Duration(retry+1))
				}
			}

			if scrollErr != nil {
				middleware.Logger.Errorf("滚动查询失败: %v", scrollErr)
				break
			}

			if newScrollID != "" {
				scrollID = newScrollID
			}

			hits = newHits
			batchSize = len(hits)
			batchCount++

			if batchSize > 0 {
				if err := processBatch(hits, batchCount); err != nil {
					if err.Error() != "已达到请求的数据量限制" {
						middleware.Logger.Warnf("处理批次数据失败: %v", err)
					} else {
						middleware.Logger.Infof("已达到请求的数据量限制: %d", req.Count)
						break
					}
				}

				// 显式释放内存，帮助GC
				hits = nil
				newHits = nil
			}

			if batchCount%10 == 0 {
				middleware.Logger.Infof("已处理 %d 批数据，累计: %d 条记录", batchCount, totalCount)
			}
		}

		processedHits = totalCount
		middleware.Logger.Infof("滚动查询完成，共处理 %d 条记录", processedHits)
	}

	// 从Redis获取最终统计结果
	result, err := getAnalysisResultFromRedis(tempKey, ctx)
	if err != nil {
		return nil, fmt.Errorf("获取分析结果失败: %v", err)
	}

	middleware.Logger.Infof("分析完成，共统计 %d 个不同的值", len(result.Values))
	return result, nil
}

// getFieldValueFromSource 从记录中获取字段值
func getFieldValueFromSource(source map[string]interface{}, field string) (string, error) {
	// 处理嵌套字段，如 content.message
	parts := strings.Split(field, ".")
	var current interface{} = source

	// 遍历字段路径
	for i, part := range parts {
		// 检查当前节点是否为map
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("字段路径 %s 的第 %d 部分不是一个对象", field, i)
		}

		// 获取下一级字段
		value, exists := currentMap[part]
		if !exists {
			return "", fmt.Errorf("字段 %s 不存在", part)
		}

		// 更新当前节点
		current = value
	}

	// 转换为字符串
	switch v := current.(type) {
	case string:
		return v, nil
	case float64:
		// 判断是否为整数，避免科学计数法
		if v == float64(int64(v)) {
			// 是整数，使用%.0f格式化，避免科学计数法
			return fmt.Sprintf("%.0f", v), nil
		}
		// 是小数，保留完整精度
		return fmt.Sprintf("%f", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case nil:
		return "", nil
	default:
		// 尝试将复杂类型序列化为JSON
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("无法将值转换为字符串: %v", err)
		}
		return string(bytes), nil
	}
}

// extractWithDelimiter 使用分隔符从字符串中提取内容
func extractWithDelimiter(content, startDelimiter, endDelimiter string) (string, error) {
	// 如果开始分隔符为空，则从头开始
	startIndex := 0
	if startDelimiter != "" {
		idx := strings.Index(content, startDelimiter)
		if idx == -1 {
			return "", fmt.Errorf("未找到开始分隔符: %s", startDelimiter)
		}
		startIndex = idx + len(startDelimiter)
	}

	// 如果结束分隔符为空，则提取到字符串末尾
	if endDelimiter == "" {
		return content[startIndex:], nil
	}

	endIndex := strings.Index(content[startIndex:], endDelimiter)
	if endIndex == -1 {
		// 如果未找到结束分隔符，也提取到字符串末尾
		return content[startIndex:], nil
	}

	return content[startIndex : startIndex+endIndex], nil
}

// processBatchToRedis 将批次数据处理并写入Redis
func processBatchToRedis(hits []interface{}, req AnalysisRequest, tempKey string, ctx context.Context) error {
	if len(hits) == 0 {
		return nil
	}

	// 使用Pipeline批量操作Redis
	pipe := database.RDB.Pipeline()
	processedCount := 0

	for i, hit := range hits {
		// 将接口类型转换为map
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			if i < 5 { // 只记录前5条错误
				middleware.Logger.Warnf("记录格式错误，无法转换为map: %v", hit)
			}
			continue
		}

		// 获取_source字段
		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			if i < 5 { // 只记录前5条错误
				middleware.Logger.Warnf("记录缺少_source字段或格式错误")
			}
			continue
		}

		// 获取字段值
		value, err := getFieldValueFromSource(source, req.Field)
		if err != nil {
			if i < 5 { // 只记录前5条错误
				middleware.Logger.Warnf("获取字段 %s 值失败: %v", req.Field, err)
			}
			continue
		}

		// 如果配置了分隔符，进一步提取内容
		if req.StartDelimiter != "" || req.EndDelimiter != "" {
			extracted, err := extractWithDelimiter(value, req.StartDelimiter, req.EndDelimiter)
			if err != nil {
				if i < 5 { // 只记录前5条错误
					middleware.Logger.Warnf("使用分隔符提取值失败: %v", err)
				}
				continue
			}
			value = extracted
		}

		// 将值添加到Redis Hash中进行计数
		pipe.HIncrBy(ctx, tempKey, value, 1)
		processedCount++
	}

	// 执行Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("Redis Pipeline执行失败: %v", err)
	}

	middleware.Logger.Infof("批次处理完成，成功处理 %d/%d 条记录", processedCount, len(hits))
	return nil
}

// getAnalysisResultFromRedis 从Redis获取分析结果
func getAnalysisResultFromRedis(tempKey string, ctx context.Context) (*AnalysisResult, error) {
	// 从Redis Hash获取所有字段和值
	result := database.RDB.HGetAll(ctx, tempKey).Val()
	if len(result) == 0 {
		return &AnalysisResult{Values: make(map[string]int)}, nil
	}

	// 转换为分析结果格式
	values := make(map[string]int)
	for key, value := range result {
		if count, err := strconv.Atoi(value); err == nil {
			values[key] = count
		} else {
			middleware.Logger.Warnf("转换计数值失败: %s = %s", key, value)
		}
	}

	return &AnalysisResult{Values: values}, nil
}

// buildAnalysisResult 构建分析结果（保留兼容性）
func buildAnalysisResult(valueMap map[string]int) *AnalysisResult {
	return &AnalysisResult{
		Values: valueMap,
	}
}
