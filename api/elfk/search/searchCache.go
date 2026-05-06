package search

import (
	"cmdb/pkg/middleware"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ESSearchResult ES搜索结果结构体
type ESSearchResult struct {
	Project    string                   `json:"project"`     // 项目标识
	LogType    string                   `json:"log_type"`    // 来源
	IndexName  string                   `json:"index_name"`  // 索引名称
	QueryTime  int64                    `json:"query_time"`  // 查询耗时(毫秒)
	TotalHits  int64                    `json:"total_hits"`  // 总命中数(ES报告的真实总命中数)
	ActualHits int                      `json:"actual_hits"` // 实际存储的数据量(用于分页计算)
	Hits       []map[string]interface{} `json:"hits"`        // 查询结果
	Page       int                      `json:"page"`        // 当前页码
	Size       int                      `json:"size"`        // 每页大小
	Pages      int                      `json:"pages"`       // 总页数
	QueryID    string                   `json:"query_id"`    // 查询唯一标识符
}

// 缓存配置
const (
	// 默认缓存过期时间（1小时）
	DefaultCacheExpiration = 60 * 60
	// 最大缓存结果数
	MaxCacheResults = 10000
	// 默认页大小
	DefaultPageSize = 20
)

// GenerateESQueryID 生成随机查询ID
func GenerateESQueryID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳作为备选
		return fmt.Sprintf("es-qid-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("es-qid-%s", hex.EncodeToString(b))
}

// GetESResultFromCacheByID 从缓存获取ES查询结果（基于查询ID）
func GetESResultFromCacheByID(ctx context.Context, redisClient *redis.Client, queryID string) (*ESSearchResult, bool) {
	if redisClient == nil {
		middleware.Logger.Error("Redis客户端未初始化")
		return nil, false
	}

	cacheKey := middleware.GetESSearchKey(queryID)

	// 从 Redis 获取缓存数据
	cachedData, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		// 缓存未命中或发生错误
		middleware.Logger.Errorf("ES缓存获取失败: %s - %v", cacheKey, err)
		return nil, false
	}

	// 解析缓存数据
	var result ESSearchResult
	if err := json.Unmarshal(cachedData, &result); err != nil {
		// 解析错误，视为缓存未命中
		middleware.Logger.Error("ES缓存解析失败: %s - %v", cacheKey, err)
		return nil, false
	}

	return &result, true
}

// SaveESResultToCache 将ES查询结果存入缓存
func SaveESResultToCache(ctx context.Context, redisClient *redis.Client, result *ESSearchResult, expiration time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	// 如果没有查询ID，生成一个
	if result.QueryID == "" {
		result.QueryID = GenerateESQueryID()
	}

	// 限制缓存结果数量为10000条
	if len(result.Hits) > MaxCacheResults {
		result.Hits = result.Hits[:MaxCacheResults]
	}

	// 设置实际数据量
	result.ActualHits = len(result.Hits)

	cacheKey := middleware.GetESSearchKey(result.QueryID)
	// 序列化结果
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化ES查询结果失败: %w", err)
	}

	// 如果未指定过期时间，使用默认值
	if expiration == 0 {
		expiration = time.Duration(DefaultCacheExpiration) * time.Second
	}

	// 存入 Redis
	return redisClient.Set(ctx, cacheKey, resultBytes, expiration).Err()
}

// GetESPagedResult 根据查询ID获取指定页的缓存结果
func GetESPagedResult(ctx context.Context, redisClient *redis.Client, queryID string, page, size int) (*ESSearchResult, error) {
	// 验证参数
	if queryID == "" {
		return nil, fmt.Errorf("查询ID不能为空")
	}

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = DefaultPageSize
	}

	// 从缓存获取原始结果
	cachedResult, found := GetESResultFromCacheByID(ctx, redisClient, queryID)
	if !found {
		return nil, fmt.Errorf("缓存未命中，请先执行查询")
	}

	// 应用分页
	pagedResult := applyESPagination(cachedResult, page, size)

	return pagedResult, nil
}

// applyESPagination 对ES查询结果应用分页
func applyESPagination(result *ESSearchResult, page, size int) *ESSearchResult {
	if result == nil {
		return nil
	}

	if result.LogType == "elfk" {
		// elfk: 缓存里是全量数据，按实际数据量分页
		actualHits := result.ActualHits
		if actualHits == 0 {
			actualHits = len(result.Hits)
		}
		totalPages := (actualHits + size - 1) / size
		if totalPages <= 0 {
			totalPages = 1
		}

		startIdx := (page - 1) * size
		endIdx := startIdx + size
		if startIdx >= actualHits {
			return &ESSearchResult{
				Project:    result.Project,
				LogType:    result.LogType,
				IndexName:  result.IndexName,
				QueryTime:  result.QueryTime,
				TotalHits:  result.TotalHits,
				ActualHits: actualHits,
				QueryID:    result.QueryID,
				Page:       page,
				Size:       size,
				Pages:      totalPages,
				Hits:       []map[string]interface{}{},
			}
		}
		if endIdx > actualHits {
			endIdx = actualHits
		}
		pagedHits := make([]map[string]interface{}, endIdx-startIdx)
		copy(pagedHits, result.Hits[startIdx:endIdx])

		return &ESSearchResult{
			Project:    result.Project,
			LogType:    result.LogType,
			IndexName:  result.IndexName,
			QueryTime:  result.QueryTime,
			TotalHits:  result.TotalHits,
			ActualHits: actualHits,
			QueryID:    result.QueryID,
			Page:       page,
			Size:       size,
			Pages:      totalPages,
			Hits:       pagedHits,
		}
	} else {
		// sls: 缓存里只有一页数据，直接返回
		totalPages := (int(result.TotalHits) + size - 1) / size
		if totalPages <= 0 {
			totalPages = 1
		}
		return &ESSearchResult{
			Project:    result.Project,
			LogType:    result.LogType,
			IndexName:  result.IndexName,
			QueryTime:  result.QueryTime,
			TotalHits:  result.TotalHits,
			ActualHits: result.ActualHits,
			QueryID:    result.QueryID,
			Page:       page,
			Size:       size,
			Pages:      totalPages,
			Hits:       result.Hits, // 直接返回缓存内容
		}
	}
}
