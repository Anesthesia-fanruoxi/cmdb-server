package search

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ESSearchDataRequest ES数据搜索请求结构体
type ESSearchDataRequest struct {
	ViewName     string `json:"view_name"`     // 视图名称
	IndexPattern string `json:"index_pattern"` // 索引名称
	Project      string `json:"project"`       // 项目标识
	StartTime    string `json:"start_time"`    // 开始时间
	EndTime      string `json:"end_time"`      // 结束时间
	TimeField    string `json:"time_field"`    // 时间字段，如 "@timestamp"
	TimeFormat   string `json:"time_format"`   // 时间格式，如 "iso8601"
	Keyword      string `json:"keyword"`       // 搜索关键词
	Size         int    `json:"size"`          // 获取的数据量，最大10000，默认10000
	LogType      string `json:"log_type"`      // 新增字段：elfk 或 sls
	SortOrder    string `json:"sort_order"`    // 时间排序：asc(升序) 或 desc(降序)，默认desc
}

// LogQueryParams 日志查询参数
type LogQueryParams struct {
	Project      string `json:"project"`
	ViewID       int    `json:"view_id"`
	ViewName     string `json:"view_name"`
	IndexPattern string `json:"index_pattern"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	TimeField    string `json:"time_field"`
	TimeFormat   string `json:"time_format"`
	Keyword      string `json:"keyword"`
	Size         int    `json:"size"`       // 返回结果数量
	LogType      string `json:"log_type"`   // 新增字段：elfk 或 sls
	SortOrder    string `json:"sort_order"` // 时间排序：asc(升序) 或 desc(降序)，默认desc
}

// ESSearchDataList ES数据搜索
func ESSearchDataList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 记录开始时间，用于计算执行时间
	startTime := time.Now()

	// 解析请求数据
	var req ESSearchDataRequest
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
	if req.StartTime == "" || req.EndTime == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "时间范围不能为空")
		return
	}
	if req.TimeField == "" {
		req.TimeField = "@timestamp" // 默认使用 @timestamp 作为时间字段
	}
	if req.TimeFormat == "" {
		req.TimeFormat = "iso8601" // 默认使用 ISO 格式
	}
	if req.Size <= 0 || req.Size > MaxCacheResults {
		req.Size = MaxCacheResults // 默认最大返回条数
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc" // 默认降序排列
	}

	// 声明变量并获取项目权限
	var projects []string
	var isSuperAdmin bool
	var err error
	projects, isSuperAdmin, err = utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果不是超级管理员且没有权限访问该项目，返回无权限错误
	if !isSuperAdmin && !utils.Contains(projects, req.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 转换查询参数
	queryParams := LogQueryParams{
		Project:      req.Project,
		ViewName:     req.ViewName,
		IndexPattern: req.IndexPattern,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		TimeField:    req.TimeField,
		TimeFormat:   req.TimeFormat,
		Keyword:      req.Keyword,
		Size:         2000, // 固定值2000
		LogType:      req.LogType,
		SortOrder:    req.SortOrder,
	}

	// 根据LogType选择查询方法
	var result map[string]interface{}

	if req.LogType == "sls" {
		// 查询SLS日志
		logResult, queryErr := QuerySLSLogs(queryParams)
		if queryErr != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("SLS日志查询失败: %v", queryErr))
			return
		}

		// 将SLS查询结果转换为标准格式
		result, err = convertSLSResultToStandard(logResult)
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("转换SLS结果失败: %v", err))
			return
		}
	} else {
		// 默认使用ELFK查询
		result, err = QueryELFKLogs(queryParams)
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("ELFK日志查询失败: %v", err))
			return
		}
	}

	// 处理搜索结果并缓存
	esResult, err := processESSearchResult(result, req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("处理搜索结果失败: %v", err))
		return
	}

	// 计算查询耗时
	queryTime := time.Since(startTime).Milliseconds()
	esResult.QueryTime = queryTime

	// 修复点：确保在使用前获取context
	ctx := r.Context()

	// 将完整结果缓存到Redis
	if database.RDB != nil {
		err := SaveESResultToCache(ctx, database.RDB, esResult, 0)
		if err != nil {
			middleware.Logger.Errorf("缓存ES结果失败: %v", err)
		}
	} else {
		middleware.Logger.Error("警告：Redis客户端未初始化")
	}

	// 获取第一页结果（默认100条）
	firstPageResult := applyESPagination(esResult, 1, DefaultPageSize)

	// 记录审计日志（查询操作）
	RecordESSearchAudit(
		r,
		esResult.QueryID,
		req.Project,
		req.ViewName,
		req.IndexPattern,
		req.TimeField,
		req.StartTime,
		req.EndTime,
		req.Keyword,
		queryTime,
		int(esResult.TotalHits),
		OperationQuery,
		1,
		http.StatusOK,
		nil,
	)

	// 返回第一页结果及查询ID
	middleware.ResponseSuccess(w, "查询成功", firstPageResult)
}

// processESSearchResult 处理ES搜索结果，转换为我们需要的格式并准备缓存
func processESSearchResult(result map[string]interface{}, req ESSearchDataRequest) (*ESSearchResult, error) {
	esResult := &ESSearchResult{
		Project:   req.Project,
		IndexName: req.IndexPattern,
		QueryID:   GenerateESQueryID(),
		Size:      DefaultPageSize,
		Page:      1,
		Hits:      []map[string]interface{}{}, // 确保初始化
		LogType:   req.LogType,                // 新增，确保 LogType 传递
	}

	// 新版ES格式处理
	if hitsObj, ok := result["hits"].(map[string]interface{}); ok {
		// 提取total
		if total, ok := hitsObj["total"].(map[string]interface{}); ok {
			if val, ok := total["value"]; ok {
				// 处理不同类型的total值
				switch v := val.(type) {
				case float64:
					esResult.TotalHits = int64(v)
				case int64:
					esResult.TotalHits = v
				case int:
					esResult.TotalHits = int64(v)
				case string:
					parsedVal, _ := strconv.ParseInt(v, 10, 64)
					esResult.TotalHits = parsedVal
				}
				middleware.Logger.Infof("从hitsObj[total][value]提取的总数: %d", esResult.TotalHits)
			}
		}

		// 关键修复：正确处理hits数组，保留完整的hit对象（包含_index、_id、_score、_source等）
		if hits, ok := hitsObj["hits"].([]interface{}); ok {
			// 添加调试日志，输出实际返回的hits数量
			middleware.Logger.Infof("ES搜索返回的hits数量: %d", len(hits))

			for _, hit := range hits {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					// 保留完整的hit对象，包含所有元数据
					esResult.Hits = append(esResult.Hits, hitMap)
				}
			}

			// 添加调试日志，输出处理后的hits数量
			middleware.Logger.Infof("处理后存入Redis的hits数量: %d", len(esResult.Hits))
		}
	}

	// 旧版格式兼容处理 - 添加这部分回来
	if esResult.TotalHits == 0 {
		if total, ok := result["total"]; ok {
			switch v := total.(type) {
			case float64:
				esResult.TotalHits = int64(v)
			case int64:
				esResult.TotalHits = v
			case int:
				esResult.TotalHits = int64(v)
			case string:
				parsedVal, _ := strconv.ParseInt(v, 10, 64)
				esResult.TotalHits = parsedVal
			}
			middleware.Logger.Infof("从result[total]提取的总数: %d", esResult.TotalHits)
		}
		if hits, ok := result["hits"].([]interface{}); ok {
			for _, hit := range hits {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					esResult.Hits = append(esResult.Hits, hitMap)
				}
			}
		}
	}

	// 计算页数
	if esResult.Size > 0 {
		esResult.Pages = (int(esResult.TotalHits) + esResult.Size - 1) / esResult.Size
	}

	return esResult, nil
}
