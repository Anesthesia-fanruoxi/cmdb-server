package export

import (
	"cmdb/config"
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SSE轮询间隔
const (
	sseNormalInterval    = 2 * time.Second        // 正常轮询间隔
	sseExportingInterval = 500 * time.Millisecond // 导出中轮询间隔
)

// ListExportsResponse 列出SQL审批响应
type ListExportsResponse struct {
	TotalCount int64               `json:"total_count"`
	Export     []sql.SqlExportItem `json:"export"`
	Progress   []*ExportProgress   `json:"progress,omitempty"`
}

// SqlExportList SSE实时推送导出列表
func SqlExportList(w http.ResponseWriter, r *http.Request) {
	// 从URL获取token
	query := r.URL.Query()
	token := query.Get("token")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	// 验证token
	jwtConfig := config.GetConfig().JWT
	userInfo, err := utils.ParseToken(token, jwtConfig.Secret)
	if err != nil {
		middleware.ResponseError(w, http.StatusUnauthorized, "token无效: "+err.Error())
		return
	}

	// 获取查询参数
	var status *int
	if query.Get("status") != "" {
		if s, err := strconv.Atoi(query.Get("status")); err == nil {
			status = &s
		}
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	middleware.Logger.Infof("SSE连接成功，用户: %s", userInfo.UserName)
	sendSSEEvent(w, flusher, "connected", map[string]string{"message": "SSE连接成功"})

	var lastDataHash string
	var hasExporting bool
	ctx := r.Context()

	// 立即发送第一次数据
	lastDataHash, hasExporting = queryAndSendExportData(w, flusher, userInfo.UserID, status, lastDataHash)

	for {
		// 根据是否有导出中任务决定轮询间隔
		interval := sseNormalInterval
		if hasExporting {
			interval = sseExportingInterval
		}

		select {
		case <-ctx.Done():
			middleware.Logger.Infof("SSE客户端断开，用户: %s", userInfo.UserName)
			return
		case <-time.After(interval):
			lastDataHash, hasExporting = queryAndSendExportData(w, flusher, userInfo.UserID, status, lastDataHash)
		}
	}
}

// queryAndSendExportData 查询数据并发送
func queryAndSendExportData(w http.ResponseWriter, flusher http.Flusher, userID int64, status *int, lastHash string) (string, bool) {
	db := database.DB.Table(middleware.TableSQLExport + " AS se")
	db = db.Joins("LEFT JOIN " + sql.TableProjectDict + " AS pd ON se.project = pd.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS submitter ON se.submitter_id = submitter.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS applier ON se.apply_id = applier.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS reviewer ON se.reviewer_id = reviewer.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS executor ON se.executor_id = executor.id")

	db = db.Select(`
		se.id, 
		COALESCE(submitter.nick_name, se.submitter_name) AS submitter_name,
		COALESCE(applier.nick_name, se.apply_name) AS apply_name,
		se.apply_status,
		COALESCE(reviewer.nick_name, se.reviewer_name) AS reviewer_name,
		COALESCE(executor.nick_name, se.executor_name) AS executor_name,
		se.created_at, se.status,
		COALESCE(pd.project_name, se.project) AS project_name,
		COALESCE(pd.project, se.project) AS project
	`)

	db = db.Where("se.submitter_id = ? OR se.apply_id = ? OR se.reviewer_id = ? OR se.executor_id = ?",
		userID, userID, userID, userID)

	if status != nil {
		db = db.Where("se.status = ?", *status)
	}

	// 最多返回200条，不分页
	var results []sql.ExportListResult
	db.Order("se.created_at DESC").Limit(200).Find(&results)

	items := make([]sql.SqlExportItem, 0, len(results))
	var progressList []*ExportProgress
	hasExporting := false

	for _, result := range results {
		items = append(items, sql.SqlExportItem{
			ID:              result.ID,
			Project:         result.Project,
			ProjectName:     result.ProjectName,
			SubmitterName:   result.SubmitterName,
			ApplyName:       result.ApplyName,
			ReviewerName:    result.ReviewerName,
			ExecutorName:    result.ExecutorName,
			CreatedAt:       sql.FormatTimeField(result.CreatedAt),
			Status:          result.Status,
			StatusText:      sql.GetStatusText(result.Status),
			CurrentOperator: sql.CalculateCurrentOperatorFromListResult(&result),
		})

		// 检查是否有导出中的任务
		if result.Status == sql.StatusExporting || result.Status == sql.StatusExecuting {
			hasExporting = true
			if progress := GetExportProgress(result.ID); progress != nil {
				progressList = append(progressList, progress)
			}
		}
	}

	resp := &ListExportsResponse{
		TotalCount: int64(len(items)),
		Export:     items,
		Progress:   progressList,
	}

	dataJSON, _ := json.Marshal(resp)
	currentHash := fmt.Sprintf("%x", dataJSON)

	// 只有数据变化时才发送
	if currentHash != lastHash {
		sendSSEEvent(w, flusher, "data", resp)
		return currentHash, hasExporting
	}
	return lastHash, hasExporting
}

// sendSSEEvent 发送SSE事件
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
