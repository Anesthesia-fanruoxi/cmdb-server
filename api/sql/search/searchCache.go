package search

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// GenerateQueryID 生成随机查询ID
func GenerateQueryID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳作为备选
		return fmt.Sprintf("qid-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("qid-%s", hex.EncodeToString(b))
}

// GetQueryInfoFromCache 从缓存获取查询元信息
func GetQueryInfoFromCache(ctx context.Context, redisClient *redis.Client, queryID string) (*sql.QueryMainInfo, error) {
	mainKey := middleware.GetSqlSearchMainKey(queryID)
	middleware.Logger.Infof("获取查询元数据，主键: %s", mainKey)

	mainInfoJSON, err := redisClient.Get(ctx, mainKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取查询主信息失败: %v", err)
	}

	// 解析主信息
	var mainInfo sql.QueryMainInfo
	if err := json.Unmarshal([]byte(mainInfoJSON), &mainInfo); err != nil {
		return nil, fmt.Errorf("解析查询主信息失败: %v", err)
	}

	middleware.Logger.Infof("成功获取查询元数据: 数据库=%s, 代理=%s, 总记录数=%d",
		mainInfo.DBName, mainInfo.Agent, mainInfo.TotalCount)

	return &mainInfo, nil
}

// 获取缓存的查询批次数据
func GetQueryBatchFromCache(ctx context.Context, redisClient *redis.Client, queryID string, batchIndex int) (*sql.QueryResult, error) {
	batchKey := middleware.GetSqlSearchBatchKey(fmt.Sprintf("%s-%d", queryID, batchIndex))
	middleware.Logger.Infof("获取批次数据，批次键: %s", batchKey)

	batchDataJSON, err := redisClient.Get(ctx, batchKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取批次数据失败: %v", err)
	}

	// 先解析为通用格式，然后进行格式转换（兼容旧的对象数组格式）
	var rawBatchData map[string]interface{}
	if err := json.Unmarshal([]byte(batchDataJSON), &rawBatchData); err != nil {
		return nil, fmt.Errorf("解析批次数据失败: %v", err)
	}

	// 统一转换批次数据格式：将对象数组转换为二维数组
	if rows, exists := rawBatchData["rows"]; exists {
		if columns, colExists := rawBatchData["columns"]; colExists {
			var columnNames []string
			// 处理columns字段的不同格式
			switch cols := columns.(type) {
			case []interface{}:
				columnNames = make([]string, len(cols))
				for i, col := range cols {
					if colStr, ok := col.(string); ok {
						columnNames[i] = colStr
					}
				}
			case []string:
				columnNames = cols
			}

			if len(columnNames) > 0 {
				// 使用转换函数统一处理rows格式
				normalizedRows := normalizeRowsToArrayFormat(rows, columnNames)
				rawBatchData["rows"] = normalizedRows
			}
		}
	}

	// 重新序列化为标准格式
	normalizedJSON, err := json.Marshal(rawBatchData)
	if err != nil {
		return nil, fmt.Errorf("序列化标准化数据失败: %v", err)
	}

	var batchResult sql.QueryResult
	if err := json.Unmarshal(normalizedJSON, &batchResult); err != nil {
		return nil, fmt.Errorf("解析标准化数据失败: %v", err)
	}

	middleware.Logger.Infof("成功获取批次%d数据，行数: %d", batchIndex, len(batchResult.Rows))
	return &batchResult, nil
}

// GetAllQueryBatchesFromCache 从缓存获取所有批次数据并合并
func GetAllQueryBatchesFromCache(ctx context.Context, redisClient *redis.Client, queryID string, mainInfo *sql.QueryMainInfo) (*sql.QueryResult, error) {
	if mainInfo == nil {
		// 如果未提供mainInfo，尝试获取
		var err error
		mainInfo, err = GetQueryInfoFromCache(ctx, redisClient, queryID)
		if err != nil {
			return nil, fmt.Errorf("获取查询元信息失败: %v", err)
		}
	}

	// 创建结果对象
	result := &sql.QueryResult{
		QueryID: queryID,
		DBName:  mainInfo.DBName,
		Columns: mainInfo.Columns,
		Total:   mainInfo.TotalCount,
		Rows:    [][]interface{}{},
	}

	// 如果批次数为0，返回空结果
	if mainInfo.BatchCount <= 0 {
		middleware.Logger.Warnf("查询没有批次数据: %s", queryID)
		return result, nil
	}

	// 获取所有批次数据
	for i := 1; i <= mainInfo.BatchCount; i++ {
		batchResult, err := GetQueryBatchFromCache(ctx, redisClient, queryID, i)
		if err != nil {
			middleware.Logger.Warnf("获取批次%d数据失败: %v, 继续获取其他批次", i, err)
			continue
		}

		// 将批次数据添加到结果中
		result.Rows = append(result.Rows, batchResult.Rows...)
	}

	middleware.Logger.Infof("成功获取所有批次数据，总行数: %d", len(result.Rows))
	return result, nil
}

// GetQueryResultFromCacheByID 从缓存获取单个查询结果（基于查询ID）
func GetQueryResultFromCacheByID(ctx context.Context, redisClient *redis.Client, queryID string) (*sql.QueryResult, bool) {
	// 1. 首先尝试获取查询元信息
	mainInfo, err := GetQueryInfoFromCache(ctx, redisClient, queryID)
	if err != nil {
		middleware.Logger.Warnf("无法获取查询元信息: %v，缓存未命中", err)
		return nil, false
	}

	// 2. 如果成功获取元信息，则获取所有批次数据
	result, err := GetAllQueryBatchesFromCache(ctx, redisClient, queryID, mainInfo)
	if err != nil {
		middleware.Logger.Warnf("获取批次数据失败: %v，缓存未命中", err)
		return nil, false
	}

	// 3. 如果没有获取到数据，则返回false
	if len(result.Rows) == 0 {
		middleware.Logger.Warnf("获取到的批次数据为空，可能是查询结果为空")
		return result, false
	}

	middleware.Logger.Infof("成功从缓存获取查询结果，行数: %d", len(result.Rows))
	return result, true
}

// 从缓存获取查询信息(查询语句、数据库名、代理名)
func GetQueryInfo(ctx context.Context, redisClient *redis.Client, queryID string) (query string, dbName string, agent string, err error) {
	mainInfo, err := GetQueryInfoFromCache(ctx, redisClient, queryID)
	if err != nil {
		return "", "", "", err
	}

	return mainInfo.Query, mainInfo.DBName, mainInfo.Agent, nil
}

// ExecuteOriginalQueryForExport 根据缓存中的查询元数据，调用导出接口获取完整数据
// 使用专用的导出接口 /api/sql/export，不执行COUNT，直接返回所有数据
func ExecuteOriginalQueryForExport(ctx context.Context, queryID string) (*sql.QueryResult, error) {
	// 1. 获取查询元数据
	mainInfo, err := GetQueryInfoFromCache(ctx, database.RDB, queryID)
	if err != nil {
		return nil, fmt.Errorf("获取查询元数据失败: %v", err)
	}

	if mainInfo.Query == "" || mainInfo.DBName == "" || mainInfo.Agent == "" {
		return nil, fmt.Errorf("查询元数据不完整")
	}

	// 2. 清理SQL（移除LIMIT/OFFSET）
	cleanSQL := cleanSQLForExport(mainInfo.Query)

	// 3. 调用导出接口获取数据
	rows, cols, err := executeQuery(cleanSQL, mainInfo.DBName, mainInfo.Agent)
	if err != nil {
		return nil, fmt.Errorf("执行导出查询失败: %v", err)
	}

	// 4. 构建结果
	result := &sql.QueryResult{
		QueryID: queryID,
		DBName:  mainInfo.DBName,
		Columns: cols,
		Rows:    rows,
		Total:   int64(len(rows)),
	}
	if len(result.Columns) == 0 {
		result.Columns = mainInfo.Columns
	}

	return result, nil
}

// cleanSQLForExport 清理SQL，移除LIMIT和OFFSET子句
func cleanSQLForExport(originalSQL string) string {
	cleanSQL := originalSQL

	// 使用智能方式移除 LIMIT 子句(避免误删表名中的 limit)
	cleanSQL = removeLimitClauseForExport(cleanSQL)

	// 移除 OFFSET 子句
	cleanSQL = removeOffsetClause(cleanSQL)

	return strings.TrimSuffix(strings.TrimSpace(cleanSQL), ";")
}

// removeLimitClauseForExport 智能移除 LIMIT 子句
func removeLimitClauseForExport(query string) string {
	lowerQuery := strings.ToLower(query)

	// 从后往前查找 LIMIT 关键字(使用 LastIndex 找最后一个)
	searchStr := "limit"
	idx := strings.LastIndex(lowerQuery, searchStr)

	if idx == -1 {
		return query
	}

	// 检查是否是独立的关键字
	beforeOK := idx == 0 || isWhitespaceChar(lowerQuery[idx-1])
	afterIdx := idx + len(searchStr)
	afterOK := afterIdx >= len(lowerQuery) || isWhitespaceChar(lowerQuery[afterIdx])

	if beforeOK && afterOK {
		return strings.TrimSpace(query[:idx])
	}

	return query
}

// removeOffsetClause 智能移除 OFFSET 子句
func removeOffsetClause(query string) string {
	lowerQuery := strings.ToLower(query)

	// 从后往前查找 OFFSET 关键字
	searchStr := "offset"
	idx := strings.LastIndex(lowerQuery, searchStr)

	if idx == -1 {
		return query
	}

	// 检查是否是独立的关键字
	beforeOK := idx == 0 || isWhitespaceChar(lowerQuery[idx-1])
	afterIdx := idx + len(searchStr)
	afterOK := afterIdx >= len(lowerQuery) || isWhitespaceChar(lowerQuery[afterIdx])

	if beforeOK && afterOK {
		return strings.TrimSpace(query[:idx])
	}

	return query
}

// executeQuery 执行SQL查询并返回结果（专用于导出接口）
// 使用 /api/sql/export 接口，不执行COUNT，直接返回所有数据
func executeQuery(sqlQuery, dbName, project string) ([][]interface{}, []string, error) {
	// 构建请求数据
	requestData := map[string]interface{}{
		"query":  sqlQuery,
		"dbName": dbName,
	}

	// 调用统一的 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/sql/export"
	decryptedData, err := middleware.CallAgentWithBody(project, pluginPath, requestData)
	if err != nil {
		return nil, nil, fmt.Errorf("调用Agent失败: %w", err)
	}

	// 解析标准 JSON 格式的响应
	var apiResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Results []struct {
				Columns []string        `json:"columns"`
				Rows    [][]interface{} `json:"rows"`
			} `json:"results"`
		} `json:"data"`
	}

	if err := json.Unmarshal(decryptedData, &apiResponse); err != nil {
		return nil, nil, fmt.Errorf("解析查询结果失败: %w", err)
	}

	if apiResponse.Code != 200 {
		return nil, nil, fmt.Errorf("API返回错误: code=%d, message=%s", apiResponse.Code, apiResponse.Message)
	}

	if len(apiResponse.Data.Results) == 0 {
		return [][]interface{}{}, []string{}, nil
	}

	firstResult := apiResponse.Data.Results[0]
	return firstResult.Rows, firstResult.Columns, nil
}
