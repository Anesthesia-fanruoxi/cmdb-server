package audit

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"fmt"
	"time"
)

// SQLAnalysisResult SQL分析结果结构体
type SQLAnalysisResult struct {
	// 总执行次数统计（TOP 10）
	TopExecutors []UserStat `json:"top_executors"`
	// 导出操作统计（TOP 10）
	TopExporters []UserStat `json:"top_exporters"`
	// 翻页最多的查询（TOP 10）
	TopPages []PageStat `json:"top_pages"`
	// 查询操作统计（TOP 10）
	TopSearchers []UserStat `json:"top_searchers"`
	// 时间维度统计
	TimeStats SQLTimeStats `json:"time_stats"`
}

// SQLTimeStats 时间维度统计
type SQLTimeStats struct {
	HourlyDistribution map[int]int64 `json:"hourly_distribution"` // 按小时分布
}

// AnalyzeSQL 分析SQL审计数据
func AnalyzeSQL(startTime, endTime time.Time) (*SQLAnalysisResult, error) {
	result := &SQLAnalysisResult{
		TimeStats: SQLTimeStats{
			HourlyDistribution: make(map[int]int64),
		},
	}

	// 1. 分析总执行次数 - 只统计search操作
	total, err := getSQLUserStats(startTime, endTime, "search")
	if err != nil {
		return nil, fmt.Errorf("查询SQL总执行次数统计失败: %v", err)
	}
	result.TopExecutors = total

	// 2. 分析导出操作
	export, err := getSQLUserStats(startTime, endTime, "export")
	if err != nil {
		return nil, fmt.Errorf("查询SQL导出操作统计失败: %v", err)
	}
	result.TopExporters = export

	// 3. 分析查询操作
	search, err := getSQLUserStats(startTime, endTime, "search")
	if err != nil {
		return nil, fmt.Errorf("查询SQL搜索操作统计失败: %v", err)
	}
	result.TopSearchers = search

	// 4. 分析翻页最多的查询
	page, err := getSqlTopPagedStat(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("查询SQL翻页最多的查询统计失败: %v", err)
	}
	result.TopPages = page

	// 5. 分析时间维度统计 - 只统计search操作
	hourly, err := getSQLHourlyStats(startTime, endTime, "search")
	if err != nil {
		return nil, fmt.Errorf("查询SQL时间分布统计失败: %v", err)
	}
	result.TimeStats.HourlyDistribution = hourly

	return result, nil
}

// getSQLUserStats 获取SQL用户统计数据
func getSQLUserStats(startTime, endTime time.Time, operation string) ([]UserStat, error) {
	var stats []UserStat

	query := database.DB.Table(middleware.TableSqlAuditLog+" l").
		Select(`
			u.nick_name,
			COUNT(*) as execution_count,
			SUM(CASE WHEN l.platform = 'web' THEN 1 ELSE 0 END) as web_count,
			SUM(CASE WHEN l.platform = 'desktop' THEN 1 ELSE 0 END) as desktop_count
		`).
		Joins("LEFT JOIN "+middleware.TableUsers+" u ON l.user_id = u.id").
		Where("l.created_at >= ? AND l.created_at < ?", startTime, endTime)

	if operation != "" {
		query = query.Where("l.operation = ?", operation)
	}

	err := query.Group("l.user_id, l.user_name, u.nick_name").
		Order("execution_count DESC").
		Limit(10).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// getSQLHourlyStats 获取SQL小时统计数据
func getSQLHourlyStats(startTime, endTime time.Time, operation string) (map[int]int64, error) {
	var hourlyStats []struct {
		Hour           int   `gorm:"column:hour_of_day"`
		ExecutionCount int64 `gorm:"column:execution_count"`
	}

	query := database.DB.Table(middleware.TableSqlAuditLog).
		Select(`
			HOUR(created_at) as hour_of_day,
			COUNT(*) as execution_count
		`).
		Where("created_at >= ? AND created_at < ?", startTime, endTime)

	// 如果指定了操作类型，则添加筛选条件
	if operation != "" {
		query = query.Where("operation = ?", operation)
	}

	err := query.
		Group("hour_of_day").
		Order("hour_of_day ASC").
		Scan(&hourlyStats).Error

	if err != nil {
		return nil, err
	}

	hourlyDist := make(map[int]int64)
	for _, stat := range hourlyStats {
		hourlyDist[stat.Hour] = stat.ExecutionCount
	}

	return hourlyDist, nil
}

// getTopPagedStat 获取翻页最多的查询
func getSqlTopPagedStat(startTime, endTime time.Time) ([]PageStat, error) {
	var stats []PageStat

	// 按查询ID统计翻页次数，并从search操作获取平台信息
	err := database.DB.Table(middleware.TableSqlAuditLog+" l").
		Select(`
			u.nick_name,
			l.query_id,
			COUNT(*) AS page_count,
			(SELECT platform FROM `+middleware.TableSqlAuditLog+` WHERE query_id = l.query_id AND operation = 'search' LIMIT 1) AS platform
		`).
		Joins("left join "+middleware.TableUsers+" u on l.user_id = u.id").
		Where("l.operation = 'page' AND l.created_at >= ? AND l.created_at < ?", startTime, endTime).
		Group("l.query_id").
		Having("COUNT(*) > 0").
		Order("page_count DESC").
		Limit(10).
		Find(&stats).Error

	if err != nil {
		return nil, err
	}

	return stats, nil
}
