package task

import (
	"cmdb/config"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TaskDetail 任务详情SSE接口
func TaskDetail(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token := query.Get("token")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	jwtConfig := config.GetConfig().JWT
	userInfo, err := utils.ParseToken(token, jwtConfig.Secret)
	if err != nil {
		middleware.ResponseError(w, http.StatusUnauthorized, "token无效: "+err.Error())
		return
	}

	taskID := query.Get("id")
	if taskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	sendSSEEvent(w, flusher, "connected", map[string]string{"message": "SSE连接成功"})

	var lastDataHash string
	ctx := r.Context()

	lastDataHash = queryAndSendDetail(w, flusher, userInfo.UserID, taskID, lastDataHash, isAdmin)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sseRunningInterval):
			lastDataHash = queryAndSendDetail(w, flusher, userInfo.UserID, taskID, lastDataHash, isAdmin)
		}
	}
}

// queryAndSendDetail 查询并发送任务详情
func queryAndSendDetail(w http.ResponseWriter, flusher http.Flusher, userID int64, taskID, lastHash string, isAdmin bool) string {
	var task *Task
	var err error

	if isAdmin {
		// 管理员：优先从数据库获取任务的真实用户ID
		dbTask, dbErr := GetTaskByTaskID(taskID)
		if dbErr != nil {
			// 数据库查询失败（可能任务刚创建还没写入），尝试用当前用户ID查询
			task, err = GlobalTaskManager.GetTask(userID, taskID)
			if err != nil {
				sendSSEEvent(w, flusher, "error", map[string]string{"message": "任务不存在"})
				return lastHash
			}
		} else {
			// 使用任务的真实用户ID从Redis获取详情
			task, err = GlobalTaskManager.GetTask(dbTask.UserID, taskID)
			if err != nil {
				sendSSEEvent(w, flusher, "error", map[string]string{"message": "任务不存在"})
				return lastHash
			}
		}
	} else {
		// 普通用户直接用自己的ID查询
		task, err = GlobalTaskManager.GetTask(userID, taskID)
		if err != nil {
			sendSSEEvent(w, flusher, "error", map[string]string{"message": "任务不存在"})
			return lastHash
		}

		// 普通用户只能查看自己的任务
		if task.UserID != userID {
			sendSSEEvent(w, flusher, "error", map[string]string{"message": "无权限查看该任务"})
			return lastHash
		}
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.In(loc).Format("2006-01-02 15:04:05")
	}

	result := map[string]interface{}{
		"id":              task.ID,
		"type":            task.Type,
		"type_text":       getTypeText(task.Type),
		"status":          task.Status,
		"status_text":     getStatusText(task.Status),
		"progress":        task.Progress,
		"processed_count": task.ProcessedCount,
		"total_count":     task.TotalCount,
		"result_key":      task.ResultKey,
		"error_message":   task.ErrorMessage,
		"created_at":      formatTime(task.CreatedAt),
		"updated_at":      formatTime(task.UpdatedAt),
		"setup": map[string]interface{}{
			"setup1": map[string]interface{}{"status": task.Setup.Setup1.Status, "start_time": formatTime(task.Setup.Setup1.StartTime), "end_time": formatTime(task.Setup.Setup1.EndTime), "duration": task.Setup.Setup1.Duration, "description": task.Setup.Setup1.Description},
			"setup2": map[string]interface{}{"status": task.Setup.Setup2.Status, "start_time": formatTime(task.Setup.Setup2.StartTime), "end_time": formatTime(task.Setup.Setup2.EndTime), "duration": task.Setup.Setup2.Duration, "description": task.Setup.Setup2.Description},
			"setup3": map[string]interface{}{"status": task.Setup.Setup3.Status, "start_time": formatTime(task.Setup.Setup3.StartTime), "end_time": formatTime(task.Setup.Setup3.EndTime), "duration": task.Setup.Setup3.Duration, "description": task.Setup.Setup3.Description},
			"setup4": map[string]interface{}{"status": task.Setup.Setup4.Status, "start_time": formatTime(task.Setup.Setup4.StartTime), "end_time": formatTime(task.Setup.Setup4.EndTime), "duration": task.Setup.Setup4.Duration, "description": task.Setup.Setup4.Description},
		},
	}

	dataJSON, _ := json.Marshal(result)
	currentHash := fmt.Sprintf("%x", dataJSON)

	if currentHash != lastHash {
		sendSSEEvent(w, flusher, "data", result)
	}
	return currentHash
}

// TaskCancel 取消任务
func TaskCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持PUT请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	var task *Task
	var err error

	// 管理员需要先从数据库获取任务的真实用户ID
	if isAdmin {
		dbTask, dbErr := GetTaskByTaskID(taskID)
		if dbErr != nil {
			middleware.ResponseError(w, http.StatusNotFound, "任务不存在: "+dbErr.Error())
			return
		}
		// 使用任务的真实用户ID从Redis获取详情
		task, err = GlobalTaskManager.GetTask(dbTask.UserID, taskID)
	} else {
		// 普通用户直接用自己的ID查询
		task, err = GlobalTaskManager.GetTask(userInfo.UserID, taskID)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在: "+err.Error())
		return
	}

	// 普通用户只能取消自己的任务
	if !isAdmin && task.UserID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "无权限取消该任务")
		return
	}

	if err := GlobalTaskManager.CancelTask(taskID); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "取消任务失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "任务取消成功", map[string]string{"task_id": taskID})
}

// TaskPreview 获取任务结果预览
func TaskPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Page int    `json:"page"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败")
		return
	}
	defer r.Body.Close()

	if req.ID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	var task *Task
	var err error

	// 管理员需要先从数据库获取任务的真实用户ID
	if isAdmin {
		dbTask, dbErr := GetTaskByTaskID(req.ID)
		if dbErr != nil {
			middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
			return
		}
		// 使用任务的真实用户ID从Redis获取详情
		task, err = GlobalTaskManager.GetTask(dbTask.UserID, req.ID)
	} else {
		// 普通用户直接用自己的ID查询
		task, err = GlobalTaskManager.GetTask(userInfo.UserID, req.ID)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 普通用户只能查看自己的任务
	if !isAdmin && task.UserID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "无权限查看该任务")
		return
	}

	if task.Status != TaskStatusSuccess {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("任务状态为 %s，尚未完成", task.Status))
		return
	}

	if task.ResultKey == "" {
		middleware.ResponseSuccess(w, "任务没有结果数据", nil)
		return
	}

	// 根据类型处理
	if req.Type == "analysis" {
		previewReq, _ := http.NewRequest(http.MethodGet, "", nil)
		q := previewReq.URL.Query()
		q.Add("id", req.ID)
		q.Add("page", strconv.Itoa(req.Page))
		q.Add("page_size", "20")
		q.Add("sort_by", "count")
		previewReq.URL.RawQuery = q.Encode()
		previewReq = previewReq.WithContext(r.Context())
		AnalysisPreview(w, previewReq)
		return
	}

	// ES导出和SQL导出使用统一的表格格式分页
	if req.Type == "es_export" || req.Type == "sql_export" {
		resultData, err := database.RDB.Get(r.Context(), task.ResultKey).Result()
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取任务结果失败")
			return
		}

		var tableData struct {
			Columns []string        `json:"columns"`
			Rows    [][]interface{} `json:"rows"`
		}
		if err := json.Unmarshal([]byte(resultData), &tableData); err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "解析预览数据失败")
			return
		}

		// 分页参数
		pageSize := 20
		cacheTotal := len(tableData.Rows) // 缓存的行数
		totalPages := (cacheTotal + pageSize - 1) / pageSize

		// 从MySQL获取真实总数
		realTotal := cacheTotal // 默认使用缓存总数
		if record, err := GetTaskByTaskID(req.ID); err == nil && record.TotalCount > 0 {
			realTotal = record.TotalCount
		}

		if req.Page > totalPages {
			req.Page = totalPages
		}
		if req.Page < 1 {
			req.Page = 1
		}

		// 计算分页范围
		start := (req.Page - 1) * pageSize
		end := start + pageSize
		if end > cacheTotal {
			end = cacheTotal
		}

		// 获取当前页数据
		pagedRows := tableData.Rows[start:end]

		middleware.ResponseSuccess(w, "获取任务结果成功", map[string]interface{}{
			"task_id":     task.ID,
			"task_type":   task.Type,
			"columns":     tableData.Columns,
			"rows":        pagedRows,
			"page":        req.Page,
			"page_size":   pageSize,
			"cache_total": cacheTotal, // 缓存总数（用于分页）
			"total":       realTotal,  // 真实总数（用于显示）
			"total_pages": totalPages,
		})
		return
	}

	// 默认返回原始结果
	resultData, err := database.RDB.Get(r.Context(), task.ResultKey).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务结果失败")
		return
	}

	var resultObj interface{}
	json.Unmarshal([]byte(resultData), &resultObj)

	middleware.ResponseSuccess(w, "获取任务结果成功", map[string]interface{}{
		"task_id":   task.ID,
		"task_type": task.Type,
		"data":      resultObj,
	})
}

// TaskExport 统一导出接口
func TaskExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败")
		return
	}
	defer r.Body.Close()

	if req.ID == "" || req.Type == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID和类型不能为空")
		return
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	var task *Task
	var err error

	// 管理员需要先从数据库获取任务的真实用户ID
	if isAdmin {
		dbTask, dbErr := GetTaskByTaskID(req.ID)
		if dbErr != nil {
			middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
			return
		}
		// 使用任务的真实用户ID从Redis获取详情
		task, err = GlobalTaskManager.GetTask(dbTask.UserID, req.ID)
	} else {
		// 普通用户直接用自己的ID查询
		task, err = GlobalTaskManager.GetTask(userInfo.UserID, req.ID)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 普通用户只能导出自己的任务
	if !isAdmin && task.UserID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "无权限导出该任务")
		return
	}

	ctx := context.WithValue(r.Context(), "export_params", req)
	r = r.WithContext(ctx)

	switch req.Type {
	case "analysis":
		ExportAnalysis(w, r)
	case "es_export", "sql_export":
		// ES导出和SQL导出使用统一的文件下载逻辑
		// 将POST请求转换为GET请求参数
		newReq, _ := http.NewRequest(http.MethodGet, "", nil)
		q := newReq.URL.Query()
		q.Add("id", req.ID)
		newReq.URL.RawQuery = q.Encode()
		newReq = newReq.WithContext(r.Context())
		TaskDownload(w, newReq)
	default:
		middleware.ResponseSuccess(w, fmt.Sprintf("不支持的任务类型: %s", req.Type), nil)
	}
}

// SSE轮询间隔
const (
	sseNormalInterval  = 2 * time.Second
	sseRunningInterval = 500 * time.Millisecond
)

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Total int                      `json:"total"`
	Tasks []map[string]interface{} `json:"tasks"`
}

// TaskList 任务列表SSE接口
func TaskList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token := query.Get("token")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	jwtConfig := config.GetConfig().JWT
	userInfo, err := utils.ParseToken(token, jwtConfig.Secret)
	if err != nil {
		middleware.ResponseError(w, http.StatusUnauthorized, "token无效: "+err.Error())
		return
	}

	taskType := query.Get("type")
	status := query.Get("status")
	keyword := query.Get("keyword") // 模糊查询关键词

	// 判断是否为管理员（RoleID = 1 为超级管理员）
	isAdmin := userInfo.RoleID == 1

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	sendSSEEvent(w, flusher, "connected", map[string]string{"message": "SSE连接成功"})

	var lastDataHash string
	var hasRunning bool
	ctx := r.Context()

	lastDataHash, hasRunning = queryAndSendTasks(w, flusher, userInfo.UserID, taskType, status, keyword, lastDataHash, isAdmin)

	for {
		interval := sseNormalInterval
		if hasRunning {
			interval = sseRunningInterval
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			lastDataHash, hasRunning = queryAndSendTasks(w, flusher, userInfo.UserID, taskType, status, keyword, lastDataHash, isAdmin)
		}
	}
}

// queryAndSendTasks 查询并发送任务数据（从MySQL查询）
func queryAndSendTasks(w http.ResponseWriter, flusher http.Flusher, userID int64, taskType, status, keyword, lastHash string, isAdmin bool) (string, bool) {
	// 管理员查看所有任务，传 userID 为 0
	queryUserID := userID
	if isAdmin {
		queryUserID = 0
	}

	tasks, err := GetTaskList(queryUserID, taskType, status, keyword)
	if err != nil {
		return lastHash, false
	}

	hasRunning := false
	formattedTasks := make([]map[string]interface{}, 0, len(tasks))
	loc, _ := time.LoadLocation("Asia/Shanghai")
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	ctx := context.Background()

	for _, t := range tasks {
		if t.Status == TaskStatusRunning || t.Status == TaskStatusPending {
			hasRunning = true
		}

		// 从Redis获取实时进度
		progress := 0
		if redisTask, err := GlobalTaskManager.GetTask(t.UserID, t.TaskID); err == nil {
			progress = redisTask.Progress
		}

		// 判断是否超过30天
		isExpired := t.CreatedAt.Before(thirtyDaysAgo)

		taskData := map[string]interface{}{
			"id":          t.TaskID,
			"type":        t.Type,
			"type_text":   getTypeText(t.Type),
			"status":      t.Status,
			"status_text": getStatusText(t.Status),
			"progress":    progress,
			"title":       t.Title,
			"created_at":  t.CreatedAt.In(loc).Format("2006-01-02 15:04:05"),
			"is_expired":  isExpired,
		}

		// 管理员查看时，添加用户昵称字段
		if isAdmin {
			taskData["nick_name"] = t.NickName
		}

		// 检查是否存在临时下载链接（只对成功的任务检查）
		if t.Status == TaskStatusSuccess {
			pattern := fmt.Sprintf("%s*", TaskDownloadCodeKeyPrefix)
			var cursor uint64
			var downloadURL string

			// 扫描所有下载码，查找匹配当前任务的
			for {
				keys, nextCursor, scanErr := database.RDB.Scan(ctx, cursor, pattern, 100).Result()
				if scanErr != nil {
					break
				}

				for _, key := range keys {
					taskID, getErr := database.RDB.Get(ctx, key).Result()
					if getErr != nil {
						continue
					}

					if taskID == t.TaskID {
						// 找到匹配的下载码
						downloadCode := strings.TrimPrefix(key, TaskDownloadCodeKeyPrefix)
						downloadURL = fmt.Sprintf("/api/tasks/download?code=%s", downloadCode)
						break
					}
				}

				if downloadURL != "" || nextCursor == 0 {
					break
				}
				cursor = nextCursor
			}

			// 如果找到临时下载链接，添加到返回数据中
			if downloadURL != "" {
				taskData["download_url"] = downloadURL
			}
		}

		formattedTasks = append(formattedTasks, taskData)
	}

	resp := &TaskListResponse{Total: len(formattedTasks), Tasks: formattedTasks}
	dataJSON, _ := json.Marshal(resp)
	currentHash := fmt.Sprintf("%x", dataJSON)

	if currentHash != lastHash {
		sendSSEEvent(w, flusher, "data", resp)
		return currentHash, hasRunning
	}
	return lastHash, hasRunning
}

// sendSSEEvent 发送SSE事件
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

// getTypeText 获取任务类型文本
func getTypeText(taskType string) string {
	switch taskType {
	case "analysis":
		return "数据分析"
	case "es_export":
		return "ES导出"
	case "sql_export":
		return "SQL导出"
	default:
		return taskType
	}
}

// getStatusText 获取任务状态文本
func getStatusText(status string) string {
	switch status {
	case TaskStatusPending:
		return "等待中"
	case TaskStatusRunning:
		return "执行中"
	case TaskStatusSuccess:
		return "已完成"
	case TaskStatusFailed:
		return "失败"
	case TaskStatusCanceled:
		return "已取消"
	default:
		return status
	}
}
