package proUpdate

import (
	"cmdb/models/assets"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ProUpdateMobileSSE 手机端任务详情SSE（直接用任务ID访问）
// GET /api/assets/proUpdate/mobile/sse?id=task_id
func ProUpdateMobileSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取任务ID
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	ctx := context.Background()

	// 检查任务是否存在
	recordJSON, err := GetTaskRecord(ctx, taskID)
	if err != nil || recordJSON == "" {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.ResponseError(w, http.StatusInternalServerError, "服务器不支持SSE")
		return
	}

	middleware.Logger.Infof("手机端SSE连接建立 - TaskID: %s", taskID)

	reqCtx := r.Context()
	redisCtx := context.Background()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastDataHash string

	// 立即发送一次数据
	if data, hash, err := fetchMobileTaskRecord(redisCtx, taskID); err == nil {
		sendMobileSSEData(w, flusher, data)
		lastDataHash = hash

		// 如果任务已完成，发送完成事件
		if isMobileTaskFinished(data) {
			sendMobileSSEComplete(w, flusher)
			middleware.Logger.Infof("手机端SSE任务完成 - TaskID: %s", taskID)
			return
		}
	} else {
		sendMobileSSEError(w, flusher, err.Error())
	}

	for {
		select {
		case <-reqCtx.Done():
			middleware.Logger.Infof("手机端SSE连接断开 - TaskID: %s", taskID)
			return
		case <-ticker.C:
			data, hash, err := fetchMobileTaskRecord(redisCtx, taskID)
			if err != nil {
				middleware.Logger.Errorf("手机端SSE获取数据失败: %v", err)
				sendMobileSSEError(w, flusher, err.Error())
				continue
			}

			if hash != lastDataHash {
				sendMobileSSEData(w, flusher, data)
				lastDataHash = hash

				if isMobileTaskFinished(data) {
					sendMobileSSEComplete(w, flusher)
					middleware.Logger.Infof("手机端SSE任务完成 - TaskID: %s", taskID)
					return
				}
			}
		}
	}
}

// fetchMobileTaskRecord 获取手机端任务记录（精简版，不包含敏感信息）
func fetchMobileTaskRecord(ctx context.Context, taskID string) (*assets.MobileTaskResponse, string, error) {
	recordJSON, err := GetTaskRecord(ctx, taskID)
	if err != nil {
		return nil, "", fmt.Errorf("查询任务记录失败: %v", err)
	}

	if recordJSON == "" {
		return nil, "", fmt.Errorf("未找到任务记录")
	}

	var taskRecords []assets.RedisTaskRecord
	if err := json.Unmarshal([]byte(recordJSON), &taskRecords); err != nil {
		return nil, "", fmt.Errorf("解析任务记录失败: %v", err)
	}

	if len(taskRecords) == 0 {
		return nil, "", fmt.Errorf("任务记录为空")
	}

	redisRecord := &taskRecords[0]

	// 根据 type 获取步骤模板
	stepTemplates := GetStepsByType(redisRecord.Type)
	totalSteps := len(stepTemplates)

	// 构建步骤状态映射
	stepStatusMap := make(map[int]*assets.RedisStepInfo)
	for i := range redisRecord.StepDetail {
		stepStatusMap[redisRecord.StepDetail[i].Step] = &redisRecord.StepDetail[i]
	}

	// 构建精简版步骤列表
	steps := make([]assets.MobileStepInfo, 0, totalSteps)
	for _, template := range stepTemplates {
		step := assets.MobileStepInfo{
			Step:       template.Step,
			StepName:   template.StepName,
			StepStatus: "pending",
		}

		if stepInfo, exists := stepStatusMap[template.Step]; exists {
			step.StepStatus = stepInfo.StepStatus
			step.StepStartedAt = stepInfo.StepStartedAt
			step.StepFinishedAt = stepInfo.StepFinishedAt
			step.Duration = stepInfo.Duration
		}

		steps = append(steps, step)
	}

	response := &assets.MobileTaskResponse{
		ProjectName:  redisRecord.ProjectName,
		ErrorMessage: redisRecord.ErrorMessage,
		StartedAt:    redisRecord.StartedAt,
		FinishedAt:   redisRecord.FinishedAt,
		Status:       redisRecord.Status,
		CurrentStep:  redisRecord.Step,
		TotalSteps:   totalSteps,
		Type:         redisRecord.Type,
		Steps:        steps,
	}

	hash := fmt.Sprintf("%s-%d-%s", redisRecord.Status, redisRecord.Step, redisRecord.Timestamp)
	return response, hash, nil
}

// isMobileTaskFinished 判断任务是否已完成
func isMobileTaskFinished(record *assets.MobileTaskResponse) bool {
	return record.Status == "success" || record.Status == "failed" || record.Status == "cancelled"
}

// sendMobileSSEData 发送SSE数据
func sendMobileSSEData(w http.ResponseWriter, flusher http.Flusher, data *assets.MobileTaskResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: data\ndata: %s\n\n", jsonData)
	flusher.Flush()
}

// sendMobileSSEError 发送SSE错误
func sendMobileSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errMsg)
	flusher.Flush()
}

// sendMobileSSEComplete 发送SSE完成事件
func sendMobileSSEComplete(w http.ResponseWriter, flusher http.Flusher) {
	fmt.Fprintf(w, "event: complete\ndata: task finished\n\n")
	flusher.Flush()
}
