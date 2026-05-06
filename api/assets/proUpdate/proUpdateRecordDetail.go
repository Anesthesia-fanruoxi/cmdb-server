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

// ProUpdateRecordDetailSSE 获取发版任务详情（SSE实时推送）
// GET /api/assets/proUpdate/record-detail/sse?id=task_id
func ProUpdateRecordDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取任务ID参数
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID参数不能为空")
		return
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用nginx缓冲

	// 获取Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.ResponseError(w, http.StatusInternalServerError, "服务器不支持SSE")
		return
	}

	middleware.Logger.Infof("SSE连接建立 - TaskID: %s", taskID)

	// 使用context监听客户端断开
	ctx := r.Context()
	redisCtx := context.Background()

	// 数据检查定时器（2秒）
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastDataHash string

	// 立即发送一次数据
	if data, hash, err := fetchTaskRecord(redisCtx, taskID); err == nil {
		sendRecordSSEData(w, flusher, data)
		lastDataHash = hash

		// 如果任务已完成，发送完成事件后关闭
		if isTaskFinished(data) {
			sendRecordSSEComplete(w, flusher)
			middleware.Logger.Infof("SSE任务已完成，关闭连接 - TaskID: %s", taskID)
			return
		}
	} else {
		sendRecordSSEError(w, flusher, err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			middleware.Logger.Infof("SSE连接断开 - TaskID: %s", taskID)
			return
		case <-ticker.C:
			// 获取最新数据
			data, hash, err := fetchTaskRecord(redisCtx, taskID)
			if err != nil {
				middleware.Logger.Errorf("SSE获取任务数据失败: %v", err)
				sendRecordSSEError(w, flusher, err.Error())
				continue
			}

			// 只有数据变化时才推送
			if hash != lastDataHash {
				sendRecordSSEData(w, flusher, data)
				lastDataHash = hash

				// 如果任务已完成，发送完成事件后关闭
				if isTaskFinished(data) {
					sendRecordSSEComplete(w, flusher)
					middleware.Logger.Infof("SSE任务已完成，关闭连接 - TaskID: %s", taskID)
					return
				}
			}
		}
	}
}

// fetchTaskRecord 从Redis获取任务记录
func fetchTaskRecord(ctx context.Context, taskID string) (*assets.RedisTaskRecord, string, error) {
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

	// 生成hash用于变化检测
	hash := fmt.Sprintf("%s-%d-%s", taskRecords[0].Status, taskRecords[0].Step, taskRecords[0].Timestamp)

	return &taskRecords[0], hash, nil
}

// isTaskFinished 判断任务是否已完成
func isTaskFinished(record *assets.RedisTaskRecord) bool {
	return record.Status == "success" || record.Status == "failed" || record.Status == "cancelled"
}

// sendRecordSSEData 发送SSE数据
func sendRecordSSEData(w http.ResponseWriter, flusher http.Flusher, data *assets.RedisTaskRecord) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: data\ndata: %s\n\n", jsonData)
	flusher.Flush()
}

// sendRecordSSEError 发送SSE错误
func sendRecordSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errMsg)
	flusher.Flush()
}

// sendRecordSSEComplete 发送SSE完成事件
func sendRecordSSEComplete(w http.ResponseWriter, flusher http.Flusher) {
	fmt.Fprintf(w, "event: complete\ndata: task finished\n\n")
	flusher.Flush()
}
