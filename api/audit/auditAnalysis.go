package audit

import (
	"encoding/json"
	"net/http"
	"time"
)

// AuditStats 审计统计数据
type AuditStats struct {
	// SQL统计
	SQLTotalStats  []UserStat    `json:"sql_total_stats"`  // SQL总执行次数
	SQLExportStats []UserStat    `json:"sql_export_stats"` // SQL导出统计
	SQLSearchStats []UserStat    `json:"sql_search_stats"` // SQL查询统计
	SQLTopPages    []PageStat    `json:"sql_top_pages"`    // SQL翻页统计
	SQLHourlyStats map[int]int64 `json:"sql_hourly_stats"` // SQL按小时统计

	// ES统计
	ESTotalStats    []UserStat    `json:"es_total_stats"`    // ES总执行次数
	ESExportStats   []UserStat    `json:"es_export_stats"`   // ES导出统计
	ESSearchStats   []UserStat    `json:"es_search_stats"`   // ES查询统计
	ESPageStats     []PageStat    `json:"es_page_stats"`     // ES翻页统计
	ESAnalysisStats []UserStat    `json:"es_analysis_stats"` // ES分析统计
	ESHourlyStats   map[int]int64 `json:"es_hourly_stats"`   // ES按小时统计
}

// UserStat 用户统计
type UserStat struct {
	NickName       string `json:"nick_name"`
	ExecutionCount int64  `json:"execution_count"`
	WebCount       int64  `json:"web_count"`     // Web平台执行次数
	DesktopCount   int64  `json:"desktop_count"` // Desktop平台执行次数
}

// QueryPageStat 查询分页统计
type PageStat struct {
	QueryID   string `json:"query_id"`   // 查询ID
	NickName  string `json:"nick_name"`  // 用户昵称
	PageCount int64  `json:"page_count"` // 翻页次数
	Platform  string `json:"platform"`   // 平台来源（从search操作获取）
}

// Response API响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// GetAuditStatsHandler 获取审计统计数据的HTTP处理函数
func GetAuditStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, Response{
			Code: http.StatusMethodNotAllowed,
			Msg:  "方法不允许",
		})
		return
	}

	// 获取查询参数
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		writeJSON(w, Response{
			Code: http.StatusBadRequest,
			Msg:  "开始时间和结束时间不能为空",
		})
		return
	}

	// 解析时间
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, time.Local)
	if err != nil {
		writeJSON(w, Response{
			Code: http.StatusBadRequest,
			Msg:  "开始时间格式错误",
		})
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, time.Local)
	if err != nil {
		writeJSON(w, Response{
			Code: http.StatusBadRequest,
			Msg:  "结束时间格式错误",
		})
		return
	}

	// 检查时间范围
	if endTime.Before(startTime) {
		writeJSON(w, Response{
			Code: http.StatusBadRequest,
			Msg:  "结束时间不能早于开始时间",
		})
		return
	}

	// 获取统计数据
	stats, err := GetAuditStats(startTime, endTime)
	if err != nil {
		writeJSON(w, Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}

	writeJSON(w, Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: stats,
	})
}

// writeJSON 写入JSON响应
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetAuditStats 获取指定时间范围的审计统计数据
func GetAuditStats(startTime, endTime time.Time) (*AuditStats, error) {
	stats := &AuditStats{}

	// 1. 获取SQL统计数据
	sqlResult, err := AnalyzeSQL(startTime, endTime)
	if err != nil {
		return nil, err
	}

	stats.SQLTotalStats = sqlResult.TopExecutors
	stats.SQLExportStats = sqlResult.TopExporters
	stats.SQLSearchStats = sqlResult.TopSearchers
	stats.SQLTopPages = sqlResult.TopPages
	stats.SQLHourlyStats = sqlResult.TimeStats.HourlyDistribution

	// 2. 获取ES统计数据
	esResult, err := AnalyzeES(startTime, endTime)
	if err != nil {
		return nil, err
	}

	stats.ESTotalStats = esResult.TopExecutors
	stats.ESExportStats = esResult.TopExporters
	stats.ESSearchStats = esResult.TopSearchers
	stats.ESPageStats = esResult.TopPages
	stats.ESAnalysisStats = esResult.TopAnalysis
	stats.ESHourlyStats = esResult.TimeStats.HourlyDistribution

	return stats, nil
}
