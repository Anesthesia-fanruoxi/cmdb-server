package search

import (
	"bytes"
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// normalizeRowsToArrayFormat 将任意格式的rows转换为二维数组
func normalizeRowsToArrayFormat(rows interface{}, columns []string) [][]interface{} {
	if rows == nil || len(columns) == 0 {
		return [][]interface{}{}
	}

	switch r := rows.(type) {
	case [][]interface{}:
		// 已经是二维数组，直接返回
		return r
	case []map[string]interface{}:
		// 对象数组转二维数组
		if len(r) == 0 {
			return [][]interface{}{}
		}
		result := make([][]interface{}, len(r))
		for i, rowMap := range r {
			row := make([]interface{}, len(columns))
			for j, col := range columns {
				if val, exists := rowMap[col]; exists {
					row[j] = val
				}
			}
			result[i] = row
		}
		return result
	case []interface{}:
		// 处理可能的混合格式
		if len(r) == 0 {
			return [][]interface{}{}
		}
		result := make([][]interface{}, len(r))
		for i, item := range r {
			switch rowItem := item.(type) {
			case []interface{}:
				result[i] = rowItem
			case map[string]interface{}:
				row := make([]interface{}, len(columns))
				for j, col := range columns {
					if val, exists := rowItem[col]; exists {
						row[j] = val
					}
				}
				result[i] = row
			default:
				result[i] = []interface{}{}
			}
		}
		return result
	default:
		return [][]interface{}{}
	}
}

// SqlSearchData 处理SQL查询请求
func SqlSearchData(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 记录开始时间，用于计算执行时间
	startTime := time.Now()

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析JSON
	var searchReq sql.SearchRequest
	if err := json.Unmarshal(body, &searchReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 设置默认分页参数
	if searchReq.Page <= 0 {
		searchReq.Page = 1
	}

	// 验证必填字段
	if searchReq.Agent == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "agent字段不能为空")
		return
	}
	if searchReq.Query == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query字段不能为空")
		return
	}
	if searchReq.DBName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "dbName字段不能为空")
		return
	}

	// 设置默认 platform 为 web
	if searchReq.Platform == "" {
		searchReq.Platform = "web"
	}

	// 如果前端传入了 query_id，使用它；否则生成新的
	queryID := searchReq.QueryID
	if queryID == "" {
		queryID = GenerateQueryID()
	}

	ctx := context.Background()

	// 检查查询是否已被取消
	if IsQueryCancelled(ctx, queryID) {
		middleware.ResponseError(w, http.StatusBadRequest, "查询已被取消")
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

	// 构建请求体 - 包含 query、dbName 和 queryID（供插件取消使用）
	requestData := map[string]string{
		"query":    searchReq.Query,
		"dbName":   searchReq.DBName,
		"query_id": queryID,
	}

	// 调用统一的 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/sql/search"
	decryptedData, err := middleware.CallAgentWithBody(searchReq.Agent, pluginPath, requestData)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "调用Agent失败: "+err.Error())
		return
	}

	// 检查查询是否已被取消
	if IsQueryCancelled(ctx, queryID) {
		middleware.Logger.Infof("[查询取消] 查询 %s 在等待响应时被取消", queryID)
		middleware.ResponseError(w, http.StatusBadRequest, "查询已被取消")
		return
	}

	// 先解析外层响应结构（兼容两种格式）
	var agentResponse struct {
		Code    int         `json:"code"`
		Msg     string      `json:"msg"`     // 旧格式
		Message string      `json:"message"` // 新格式
		Data    interface{} `json:"data"`    // 可能是map也可能是string
	}

	// 使用Decoder处理大数字
	decoder := json.NewDecoder(bytes.NewReader(decryptedData))
	decoder.UseNumber() // 使用Number类型处理数字

	if err := decoder.Decode(&agentResponse); err != nil {
		errMsg := "解析解密后的数据失败: " + err.Error()
		middleware.ResponseError(w, http.StatusInternalServerError, errMsg)
		return
	}

	// 检查响应状态：code != 200 或 message == "error"
	if agentResponse.Code != 200 {
		errorMsg := agentResponse.Msg
		if errorMsg == "" {
			errorMsg = agentResponse.Message
		}
		middleware.Logger.Errorf("【代理拒绝】Code=%d, Msg=%s", agentResponse.Code, errorMsg)
		middleware.ResponseError(w, agentResponse.Code, errorMsg)
		return
	}

	// 检查message字段是否为"error"
	if agentResponse.Message == "error" {
		// data字段包含错误信息
		errorMsg := ""
		if dataStr, ok := agentResponse.Data.(string); ok {
			errorMsg = dataStr
		} else {
			errorMsg = "未知错误"
		}
		middleware.Logger.Errorf("【代理拒绝】Message=error, Data=%s", errorMsg)
		middleware.ResponseError(w, http.StatusBadRequest, errorMsg)
		return
	}

	middleware.Logger.Infof("【代理响应成功】Code=%d, Message=%s", agentResponse.Code, agentResponse.Message)

	// 提取实际数据（只有在message不是error时才处理）
	var decompressedResult map[string]interface{}
	if dataMap, ok := agentResponse.Data.(map[string]interface{}); ok {
		decompressedResult = dataMap
	} else {
		// 如果Data不是map，说明格式不对
		middleware.ResponseError(w, http.StatusInternalServerError, "代理返回的数据格式错误")
		return
	}
	// 统一转换代理返回的数据格式：将对象数组转换为二维数组
	if results, ok := decompressedResult["results"].([]interface{}); ok {
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if rows, exists := resultMap["rows"]; exists {
					if columns, colExists := resultMap["columns"]; colExists {
						if colSlice, ok := columns.([]interface{}); ok {
							// 转换columns为字符串数组
							columnNames := make([]string, len(colSlice))
							for i, col := range colSlice {
								if colStr, ok := col.(string); ok {
									columnNames[i] = colStr
								}
							}
							// 使用转换函数统一处理rows格式
							normalizedRows := normalizeRowsToArrayFormat(rows, columnNames)
							resultMap["rows"] = normalizedRows
						}
					}
				}
			}
		}
	}

	// 将解密后的结果转换为 SearchResult 结构 - 保留json.Number类型
	resultBytes, err := json.Marshal(decompressedResult)
	if err != nil {
		errMsg := "结果序列化失败: " + err.Error()
		middleware.ResponseError(w, http.StatusInternalServerError, errMsg)
		return
	}

	// 使用Decoder解析结果，保留json.Number类型
	decoder = json.NewDecoder(bytes.NewReader(resultBytes))
	decoder.UseNumber() // 再次使用Number类型处理数字

	var searchResult sql.SearchResult
	if err := decoder.Decode(&searchResult); err != nil {
		errMsg := "结果转换失败: " + err.Error()
		middleware.Logger.Errorf("结果转换错误: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, errMsg)
		return
	}

	// 为每个结果集生成独立的唯一ID
	// 如果前端传入了 queryID，第一个结果集使用它，其余的在其基础上添加后缀
	for i := range searchResult.Results {
		var resultQueryID string
		if i == 0 && searchReq.QueryID != "" {
			resultQueryID = queryID // 使用前端传入的或之前生成的
		} else {
			resultQueryID = GenerateQueryID()
		}
		searchResult.Results[i].QueryID = resultQueryID

		// 使用新的分页缓存结构异步保存结果
		if database.RDB != nil {
			go func(id string, result *sql.QueryResult, idx int) {
				ctx := context.Background()
				// 提取当前结果集对应的SQL语句
				// 如果是多SQL查询，尝试从原始查询中提取对应的SQL语句
				querySql := searchReq.Query
				if strings.Contains(querySql, ";") {
					// 分割多条SQL语句
					sqlStatements := strings.Split(querySql, ";")
					// 确保索引在有效范围内
					if idx < len(sqlStatements) && idx >= 0 {
						querySql = strings.TrimSpace(sqlStatements[idx])
						if querySql != "" {
							querySql += ";" // 添加分号
						}
					}
				}

				// 保存主查询和批次数据
				if err := SaveQueryToCache(ctx, database.RDB, id, result, querySql, searchReq.DBName, searchReq.Agent); err != nil {
					middleware.Logger.Errorf("保存查询结果到缓存失败: %s - %v", id, err)
				}
			}(resultQueryID, searchResult.Results[i], i)
		}
	}

	// 声明分页结果变量
	var pagedResult *sql.SearchResult

	// 对首次查询结果进行分页处理，只返回第一页数据（20条）
	// 创建分页后的结果集
	pagedSearchResult := sql.SearchResult{
		DBName:  searchResult.DBName,
		Results: make([]*sql.QueryResult, len(searchResult.Results)),
		Took:    searchResult.Took,
		Total:   searchResult.Total,
		Page:    searchReq.Page,
	}

	// 对每个结果集进行分页截取
	for i, result := range searchResult.Results {
		// 计算第一页的数据范围
		startIdx := 0
		endIdx := PageCacheSize
		if endIdx > len(result.Rows) {
			endIdx = len(result.Rows)
		}

		// 创建分页后的结果
		pagedQueryResult := &sql.QueryResult{
			QueryID: result.QueryID,
			DBName:  result.DBName,
			Columns: result.Columns,
			Took:    result.Took,
			Total:   result.Total,
			Rows:    result.Rows[startIdx:endIdx], // 只取第一页数据
		}

		// 计算总页数
		totalPages := (int(result.Total) + PageCacheSize - 1) / PageCacheSize
		pagedQueryResult.Page = searchReq.Page
		pagedQueryResult.Pages = totalPages

		pagedSearchResult.Results[i] = pagedQueryResult
	}

	// 计算总页数（取所有结果集中的最大页数）
	maxPages := 0
	for _, result := range pagedSearchResult.Results {
		if result.Pages > maxPages {
			maxPages = result.Pages
		}
	}
	pagedSearchResult.Pages = maxPages

	pagedResult = &pagedSearchResult

	// 计算执行时间
	executionTime := time.Since(startTime).Milliseconds()

	// 使用goroutine并行处理每个结果集的审计记录
	// 这样不会阻塞主流程，每个结果集有独立的审计处理
	for i, result := range searchResult.Results {
		if result.QueryID != "" {
			// 创建局部变量，避免闭包问题
			currentResult := result
			currentIndex := i

			// 使用goroutine并行处理审计记录
			go func(res *sql.QueryResult, idx int) {
				// 使用每个结果集自己的执行时间
				resultExecutionTime := executionTime
				if res.Took > 0 {
					resultExecutionTime = res.Took
				}

				// 提取当前结果集对应的SQL语句
				querySql := searchReq.Query
				if strings.Contains(querySql, ";") {
					// 分割多条SQL语句
					sqlStatements := strings.Split(querySql, ";")
					// 确保索引在有效范围内
					if idx < len(sqlStatements) && idx >= 0 {
						querySql = strings.TrimSpace(sqlStatements[idx])
						if querySql != "" {
							querySql += ";" // 添加分号
						}
					}
				}

				// 计算影响行数 - 根据SQL类型和情况选择合适的方法
				var affectedRows int
				sqlType := GetSqlType(querySql)

				// 根据SQL类型选择合适的影响行数计算方法
				if sqlType == "SELECT" {
					// 对于SELECT查询，优先使用Total字段（COUNT查询结果）
					if res.Total > 0 {
						affectedRows = int(res.Total)
					} else {
						// 如果Total为0，使用实际行数
						affectedRows = len(res.Rows)
					}
				} else {
					// 对于非SELECT查询（INSERT/UPDATE/DELETE等），使用实际行数
					affectedRows = len(res.Rows)
				}

				// 记录审计日志 - 使用提取的SQL和计算好的影响行数
				RecordSqlAudit(r.Clone(r.Context()), searchReq.Agent, searchReq.DBName,
					querySql, res.QueryID, resultExecutionTime,
					"success", "", OperationSearch, idx+1, searchReq.Platform, affectedRows)
			}(currentResult, currentIndex)
		}
	}

	// 返回解密后的结果（已分页）
	middleware.ResponseSuccess(w, "查询成功", pagedResult)
}

// SaveQueryToCache 保存查询结果到缓存（主键+批次键结构）
// 智能缓存策略：
// 1. 小数据集（≤1000行）：全部缓存到Redis，快速访问
// 2. 中等数据集（1000-5000行）：缓存前3个批次（1500行），其余动态加载
// 3. 大数据集（>5000行）：只缓存第1个批次（500行），其余动态加载
func SaveQueryToCache(ctx context.Context, rdb *redis.Client, queryID string, result *sql.QueryResult, query string, dbName string, agent string) error {
	if rdb == nil || result == nil {
		return fmt.Errorf("Redis客户端或查询结果为空")
	}

	totalRows := int(result.Total)
	actualRows := len(result.Rows) // 实际返回的行数

	// 计算批次数量（基于总记录数）
	batchCount := (totalRows + BatchSize - 1) / BatchSize
	if batchCount == 0 {
		batchCount = 1
	}

	// 计算总页数
	totalPages := (totalRows + PageCacheSize - 1) / PageCacheSize

	// 智能决定缓存策略
	var cacheBatchCount int
	var cacheStrategy string

	if totalRows <= 1000 {
		// 小数据集：全部缓存
		cacheBatchCount = batchCount
		cacheStrategy = "full"
	} else if totalRows <= 5000 {
		// 中等数据集：缓存前3个批次
		cacheBatchCount = 3
		if cacheBatchCount > batchCount {
			cacheBatchCount = batchCount
		}
		cacheStrategy = "partial"
	} else {
		// 大数据集：只缓存第1个批次
		cacheBatchCount = 1
		cacheStrategy = "minimal"
	}

	// 创建主信息
	mainInfo := sql.QueryMainInfo{
		TotalCount:    result.Total,
		PageSize:      PageCacheSize,
		TotalPages:    totalPages,
		BatchCount:    batchCount,
		BatchSize:     BatchSize,
		Query:         query,
		DBName:        dbName,
		Agent:         agent,
		Columns:       result.Columns,
		CacheStrategy: cacheStrategy, // 记录缓存策略
	}

	middleware.Logger.Infof("【SaveQueryToCache】查询ID=%s, 总行数=%d, 实际行数=%d, 批次数=%d, 缓存批次=%d, 策略=%s",
		queryID, totalRows, actualRows, batchCount, cacheBatchCount, cacheStrategy)

	// 保存主信息
	mainKey := middleware.GetSqlSearchMainKey(queryID)
	mainInfoJSON, err := json.Marshal(mainInfo)
	if err != nil {
		return fmt.Errorf("序列化主信息失败: %v", err)
	}

	if err := rdb.Set(ctx, mainKey, mainInfoJSON, QueryCacheTTL).Err(); err != nil {
		return fmt.Errorf("保存主信息到Redis失败: %v", err)
	}

	// 保存批次数据
	// 策略1：如果实际返回的数据量不大（≤5000行），全部缓存，避免后续重复查询数据库
	// 策略2：如果实际返回的数据量很大（>5000行），按策略缓存
	savedBatches := 0
	actualCacheBatchCount := cacheBatchCount

	if actualRows <= 5000 {
		// 实际数据不多，全部缓存
		actualCacheBatchCount = (actualRows + BatchSize - 1) / BatchSize
		middleware.Logger.Infof("实际数据量≤5000行，调整为全部缓存策略，缓存批次数: %d", actualCacheBatchCount)
	}

	for i := 0; i < actualCacheBatchCount && i < batchCount; i++ {
		start := i * BatchSize
		end := start + BatchSize

		// 检查是否超出实际数据范围
		if start >= actualRows {
			middleware.Logger.Infof("批次%d超出实际数据范围，跳过缓存", i+1)
			break
		}

		if end > actualRows {
			end = actualRows
		}

		// 当前批次数据
		batchRows := result.Rows[start:end]
		if len(batchRows) == 0 {
			middleware.Logger.Infof("批次%d数据为空，跳过缓存", i+1)
			break
		}

		// 创建批次结果
		batchResult := &sql.QueryResult{
			QueryID: queryID + "-" + strconv.Itoa(i+1),
			DBName:  result.DBName,
			Columns: result.Columns,
			Took:    result.Took,
			Total:   result.Total,
			Rows:    batchRows,
		}

		batchKey := middleware.GetSqlSearchBatchKey(queryID + "-" + strconv.Itoa(i+1))
		batchResultJSON, err := json.Marshal(batchResult)
		if err != nil {
			middleware.Logger.Errorf("序列化批次%d失败: %v", i+1, err)
			continue
		}

		if err := rdb.Set(ctx, batchKey, batchResultJSON, QueryCacheTTL).Err(); err != nil {
			middleware.Logger.Errorf("保存批次%d到Redis失败: %v", i+1, err)
			continue
		}

		savedBatches++
		if i < 3 || i == actualCacheBatchCount-1 {
			// 只记录前3个批次和最后一个批次的日志，避免日志过多
			middleware.Logger.Infof("批次%d已缓存，行数: %d", i+1, len(batchRows))
		}
	}

	middleware.Logger.Infof("【SaveQueryToCache】完成，实际数据%d行，缓存%d个批次（共%d批次）",
		actualRows, savedBatches, batchCount)
	return nil
}

// GetQueryResultByPage 根据页码获取查询结果
func GetQueryResultByPage(ctx context.Context, rdb *redis.Client, queryID string, page int, isCount bool) (*sql.QueryResult, error) {
	if rdb == nil {
		return nil, fmt.Errorf("Redis客户端为空")
	}

	// 首先获取主信息
	mainKey := middleware.GetSqlSearchMainKey(queryID)
	mainInfoJSON, err := rdb.Get(ctx, mainKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("查询主信息不存在")
		}
		return nil, fmt.Errorf("获取查询主信息失败: %v", err)
	}

	// 解析主信息
	var mainInfo sql.QueryMainInfo
	if err := json.Unmarshal([]byte(mainInfoJSON), &mainInfo); err != nil {
		return nil, fmt.Errorf("解析查询主信息失败: %v", err)
	}

	// 计算需要哪个批次
	// 每页20条，第1页是0-19，第2页是20-39，...
	startIndex := (page - 1) * PageCacheSize
	batchIndex := (startIndex / BatchSize) + 1 // 批次从1开始计数

	// 检查批次是否超出范围
	if batchIndex > mainInfo.BatchCount {
		return nil, fmt.Errorf("请求的页码超出范围: 批次%d > 总批次%d", batchIndex, mainInfo.BatchCount)
	}

	// 获取批次数据
	batchKey := middleware.GetSqlSearchBatchKey(queryID + "-" + strconv.Itoa(batchIndex))
	batchResultJSON, err := rdb.Get(ctx, batchKey).Result()
	if err != nil {
		if err == redis.Nil {
			// 批次不存在，尝试重新加载
			batchResult, err := loadMissingBatch(ctx, rdb, queryID, mainInfo, batchIndex, isCount)
			if err != nil {
				return nil, err
			}

			// 计算在批次内的偏移
			offsetInBatch := startIndex % BatchSize
			endInBatch := offsetInBatch + PageCacheSize

			// 检查是否有足够的数据
			if offsetInBatch >= len(batchResult.Rows) {
				// 返回空结果集
				return &sql.QueryResult{
					QueryID: queryID,
					DBName:  mainInfo.DBName,
					Columns: mainInfo.Columns,
					Total:   mainInfo.TotalCount,
					Page:    page,
					Pages:   (int(mainInfo.TotalCount) + PageCacheSize - 1) / PageCacheSize,
					Rows:    make([][]interface{}, 0),
				}, nil
			}

			// 截取需要的行数据
			if endInBatch > len(batchResult.Rows) {
				endInBatch = len(batchResult.Rows)
			}
			pageRows := batchResult.Rows[offsetInBatch:endInBatch]

			// 创建分页结果
			pageResult := &sql.QueryResult{
				QueryID: queryID,
				DBName:  batchResult.DBName,
				Columns: batchResult.Columns,
				Took:    batchResult.Took,
				Total:   mainInfo.TotalCount,
				Page:    page,
				Pages:   (int(mainInfo.TotalCount) + PageCacheSize - 1) / PageCacheSize,
				Rows:    pageRows,
			}

			middleware.Logger.Infof("从动态加载的批次%d获取第%d页数据成功，行数: %d", batchIndex, page, len(pageRows))
			return pageResult, nil
		}
		return nil, fmt.Errorf("获取批次%d数据失败: %v", batchIndex, err)
	}

	// 解析批次结果
	var batchResult sql.QueryResult
	if err := json.Unmarshal([]byte(batchResultJSON), &batchResult); err != nil {
		return nil, fmt.Errorf("解析批次%d数据失败: %v", batchIndex, err)
	}

	// 计算在批次内的偏移
	offsetInBatch := startIndex % BatchSize
	endInBatch := offsetInBatch + PageCacheSize

	// 检查是否有足够的数据
	if offsetInBatch >= len(batchResult.Rows) {
		// 返回空结果集
		return &sql.QueryResult{
			QueryID: queryID,
			DBName:  mainInfo.DBName,
			Columns: mainInfo.Columns,
			Total:   mainInfo.TotalCount,
			Page:    page,
			Pages:   (int(mainInfo.TotalCount) + PageCacheSize - 1) / PageCacheSize,
			Rows:    make([][]interface{}, 0),
		}, nil
	}

	// 截取需要的行数据
	if endInBatch > len(batchResult.Rows) {
		endInBatch = len(batchResult.Rows)
	}
	pageRows := batchResult.Rows[offsetInBatch:endInBatch]

	// 创建分页结果
	pageResult := &sql.QueryResult{
		QueryID: queryID,
		DBName:  batchResult.DBName,
		Columns: batchResult.Columns,
		Took:    batchResult.Took,
		Total:   mainInfo.TotalCount,
		Page:    page,
		Pages:   (int(mainInfo.TotalCount) + PageCacheSize - 1) / PageCacheSize,
		Rows:    pageRows,
	}

	middleware.Logger.Infof("从批次%d获取第%d页数据成功，行数: %d", batchIndex, page, len(pageRows))
	return pageResult, nil
}

// GetSqlType 判断SQL语句的类型
func GetSqlType(sqlStr string) string {
	// 去除空白字符并转为大写
	trimmedSQL := strings.TrimSpace(sqlStr)
	if len(trimmedSQL) == 0 {
		return "UNKNOWN"
	}

	// 获取SQL语句的前几个字符，用于判断类型
	upperSQL := strings.ToUpper(trimmedSQL)

	// 判断SQL类型
	if strings.HasPrefix(upperSQL, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(upperSQL, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(upperSQL, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(upperSQL, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(upperSQL, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(upperSQL, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(upperSQL, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(upperSQL, "TRUNCATE") {
		return "TRUNCATE"
	}

	return "OTHER"
}

// loadMissingBatch 加载缺失的批次数据
func loadMissingBatch(ctx context.Context, rdb *redis.Client, queryID string, mainInfo sql.QueryMainInfo, batchIndex int, isCount bool) (*sql.QueryResult, error) {
	// 直接从mainInfo中获取查询信息，不再需要从Redis中获取info键
	query := mainInfo.Query
	dbName := mainInfo.DBName
	agent := mainInfo.Agent

	middleware.Logger.Infof("【loadMissingBatch】原始查询: %s", query)

	// 计算偏移量
	offset := (batchIndex - 1) * BatchSize

	// 构造带有偏移量的查询
	offsetQuery := AddOffsetLimitToQuery(query, offset, BatchSize)
	middleware.Logger.Infof("【loadMissingBatch】添加OFFSET/LIMIT后的查询: %s", offsetQuery)

	// 执行查询
	middleware.Logger.Infof("加载批次%d的数据，偏移量: %d, 限制: %d", batchIndex, offset, BatchSize)

	// 构建请求
	requestData := map[string]interface{}{
		"query":    offsetQuery,
		"dbName":   dbName,
		"is_count": isCount, // 传递是否需要count的标志
	}

	middleware.Logger.Infof("【loadMissingBatch】发送给Agent的查询: query=%s, dbName=%s, isCount=%v", offsetQuery, dbName, isCount)

	// 调用统一的 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/sql/search"
	decompressedData, err := middleware.CallAgentWithBody(agent, pluginPath, requestData)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %v", err)
	}

	// 先解析外层响应结构
	var agentResponse struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(decompressedData))
	decoder.UseNumber()

	if err := decoder.Decode(&agentResponse); err != nil {
		return nil, fmt.Errorf("解析解密后的数据失败: %v", err)
	}

	// 检查响应状态
	if agentResponse.Code != 200 {
		return nil, fmt.Errorf("Agent返回错误: %s", agentResponse.Msg)
	}

	// 提取实际数据
	decompressedResult := agentResponse.Data

	// 统一转换代理返回的数据格式：将对象数组转换为二维数组
	if results, ok := decompressedResult["results"].([]interface{}); ok {
		for _, result := range results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if rows, exists := resultMap["rows"]; exists {
					if columns, colExists := resultMap["columns"]; colExists {
						if colSlice, ok := columns.([]interface{}); ok {
							// 转换columns为字符串数组
							columnNames := make([]string, len(colSlice))
							for i, col := range colSlice {
								if colStr, ok := col.(string); ok {
									columnNames[i] = colStr
								}
							}
							// 使用转换函数统一处理rows格式
							normalizedRows := normalizeRowsToArrayFormat(rows, columnNames)
							resultMap["rows"] = normalizedRows
						}
					}
				}
			}
		}
	}

	// 重新序列化为标准格式
	resultBytes, err := json.Marshal(decompressedResult)
	if err != nil {
		return nil, fmt.Errorf("序列化标准化数据失败: %v", err)
	}

	// 使用转换后的数据解析SearchResult
	decoder = json.NewDecoder(bytes.NewReader(resultBytes))
	decoder.UseNumber()

	var searchResult sql.SearchResult
	if err := decoder.Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %v", err)
	}

	if len(searchResult.Results) == 0 {
		return nil, fmt.Errorf("搜索结果为空")
	}

	// 使用加载的结果创建批次
	batchResult := searchResult.Results[0]
	batchResult.QueryID = queryID + "-" + strconv.Itoa(batchIndex)
	batchResult.Total = mainInfo.TotalCount // 使用主信息中的总数

	// 保存批次到缓存
	batchKey := middleware.GetSqlSearchBatchKey(queryID + "-" + strconv.Itoa(batchIndex))
	batchResultJSON, err := json.Marshal(batchResult)
	if err != nil {
		middleware.Logger.Errorf("序列化批次%d结果失败: %v", batchIndex, err)
	} else {
		if err := rdb.Set(ctx, batchKey, batchResultJSON, QueryCacheTTL).Err(); err != nil {
			middleware.Logger.Errorf("保存批次%d到Redis失败: %v", batchIndex, err)
		} else {
			middleware.Logger.Infof("动态加载的批次%d已保存到Redis，键: %s，记录数: %d",
				batchIndex, batchKey, len(batchResult.Rows))
		}
	}

	return batchResult, nil
}

// SqlSearchPage 处理分页查询请求
func SqlSearchPage(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 记录开始时间，用于计算执行时间
	startTime := time.Now()

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析JSON
	var pageReq sql.PageRequest
	if err := json.Unmarshal(body, &pageReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 验证必填字段
	if pageReq.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query_id字段不能为空")
		return
	}

	// 设置默认分页参数
	if pageReq.Page <= 0 {
		pageReq.Page = 1
	}

	// 从缓存获取结果
	if database.RDB == nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "Redis客户端未初始化")
		return
	}

	ctx := r.Context()

	// 动态加载分页不需要执行count，总数已在首次查询时缓存
	isCount := false

	// 只使用新的分页结构获取查询结果
	pagedResult, err := GetQueryResultByPage(ctx, database.RDB, pageReq.QueryID, pageReq.Page, isCount)
	if err != nil {
		// 新结构获取失败，直接返回错误
		errMsg := "获取分页数据失败: " + err.Error()
		middleware.ResponseError(w, http.StatusNotFound, errMsg)
		return
	}

	// 创建与 SqlSearchData 一致的返回结构
	searchResult := &sql.SearchResult{
		DBName:  pagedResult.DBName,
		Results: []*sql.QueryResult{pagedResult},
		Took:    pagedResult.Took,
		Total:   pagedResult.Total,
		Page:    pageReq.Page,
		Pages:   pagedResult.Pages,
	}

	// 计算执行时间
	executionTime := time.Since(startTime).Milliseconds()

	// 记录审计日志 - 成功，直接设置影响行数
	affectedRows := len(pagedResult.Rows)
	RecordSqlAudit(r, pagedResult.DBName, pagedResult.DBName, "分页查询: "+pageReq.QueryID, pageReq.QueryID, executionTime, "success", "", OperationPage, pageReq.Page, "", affectedRows)

	// 返回分页结果
	middleware.ResponseSuccess(w, "查询成功", searchResult)
}
