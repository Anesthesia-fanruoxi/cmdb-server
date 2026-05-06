package apply

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
	"strings"
	"time"
)

// SSE轮询间隔
const ssePollingInterval = 3 * time.Second

// 安全格式化时间字段
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return middleware.FormatTimeToChina(*t)
}

// SqlApplyItem 精简后的SQL审批项
type SqlApplyItem struct {
	ID              int64  `json:"id"`
	SubmitterName   string `json:"submitter_name"`
	ApplyName       string `json:"apply_name"`
	ExecutorName    string `json:"executor_name"`
	ExecutionTime   string `json:"execution_time"`
	Project         string `json:"project"`
	Description     string `json:"description"` // 申请说明/注释
	Status          int    `json:"status"`
	CurrentOperator string `json:"current_operator"`
	CreatedAt       string `json:"created_at"` // 创建时间
}

// ListApplysResponse 列出SQL审批响应
type ListApplysResponse struct {
	TotalCount int64          `json:"total_count"`
	Apply      []SqlApplyItem `json:"apply"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// SqlApplyList SSE实时推送审批列表
func SqlApplyList(w http.ResponseWriter, r *http.Request) {
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

	// 提交人过滤参数
	submitterName := query.Get("submitter_name")

	var applyStatus *int
	if query.Get("apply_status") != "" {
		if s, err := strconv.Atoi(query.Get("apply_status")); err == nil {
			applyStatus = &s
		}
	}

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
	ctx := r.Context()

	// 立即发送第一次数据
	lastDataHash = queryAndSendData(w, flusher, userInfo.UserID, applyStatus, status, submitterName, lastDataHash)

	ticker := time.NewTicker(ssePollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			middleware.Logger.Infof("SSE客户端断开，用户: %s", userInfo.UserName)
			return
		case <-ticker.C:
			lastDataHash = queryAndSendData(w, flusher, userInfo.UserID, applyStatus, status, submitterName, lastDataHash)
		}
	}
}

// queryAndSendData 查询数据并发送
func queryAndSendData(w http.ResponseWriter, flusher http.Flusher, userID int64, applyStatus, status *int, submitterName, lastHash string) string {
	// 构建WHERE条件
	var whereConditions []string
	var args []interface{}

	if applyStatus != nil {
		whereConditions = append(whereConditions, "apply_status = ?")
		args = append(args, *applyStatus)
	}
	if status != nil {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, *status)
	}
	if submitterName != "" {
		whereConditions = append(whereConditions, "submitter_name LIKE ?")
		args = append(args, "%"+submitterName+"%")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " AND " + strings.Join(whereConditions, " AND ")
	}

	// 构建UNION查询，每个子查询都可以使用索引
	// 按创建时间倒序，限制100条
	unionSQL := fmt.Sprintf(`
		SELECT * FROM (
			(SELECT * FROM sql_apply WHERE submitter_id = ?%s)
			UNION
			(SELECT * FROM sql_apply WHERE apply_id = ?%s)  
			UNION
			(SELECT * FROM sql_apply WHERE executor_id = ?%s)
		) AS combined
		ORDER BY created_at DESC
		LIMIT 100
	`, whereClause, whereClause, whereClause)

	// 准备参数：每个子查询都需要userID + 额外的条件参数
	var queryArgs []interface{}
	for i := 0; i < 3; i++ {
		queryArgs = append(queryArgs, userID)
		queryArgs = append(queryArgs, args...)
	}

	// 执行查询
	var applies []*sql.SQLApply
	if err := database.DB.Raw(unionSQL, queryArgs...).Scan(&applies).Error; err != nil {
		middleware.Logger.Errorf("查询审批记录失败: %v", err)
		sendSSEEvent(w, flusher, "error", map[string]string{"message": "获取审批记录失败"})
		return lastHash
	}

	formattedApplies := make([]SqlApplyItem, 0, len(applies))
	for _, apply := range applies {
		formattedApplies = append(formattedApplies, SqlApplyItem{
			ID:              apply.ID,
			SubmitterName:   apply.SubmitterName,
			ApplyName:       apply.ApplyName,
			ExecutorName:    apply.ExecutorName,
			ExecutionTime:   formatTime(apply.ExecutionTime),
			Project:         apply.Project,
			Description:     apply.Description,
			Status:          apply.Status,
			CurrentOperator: sql.CalculateApplyCurrentOperator(apply),
			CreatedAt:       middleware.FormatTimeToChina(apply.CreatedAt),
		})
	}

	resp := &ListApplysResponse{
		TotalCount: int64(len(applies)),
		Apply:      formattedApplies,
		Page:       1,
		PageSize:   100,
		TotalPages: 1,
	}

	dataJSON, _ := json.Marshal(resp)
	currentHash := fmt.Sprintf("%x", dataJSON)

	if currentHash != lastHash {
		sendSSEEvent(w, flusher, "data", resp)
		return currentHash
	}
	return lastHash
}

// sendSSEEvent 发送SSE事件
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
