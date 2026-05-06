package task

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// saveTaskStatusToRedis 将任务状态保存到Redis
func (tm *TaskManager) saveTaskStatusToRedis(task *Task) error {
	if database.RDB == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	status := &TaskStatusResponse{
		Status:       task.Status,
		Type:         task.Type,
		Progress:     task.Progress,
		ResultKey:    task.ResultKey,
		ErrorMessage: task.ErrorMessage,
		UpdatedAt:    time.Now().In(cnLoc),
		Setup:        task.Setup,
	}

	statusJSON, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("序列化任务状态失败: %v", err)
	}

	statusKey := fmt.Sprintf("%s%d:%s:status", TaskStatusKeyPrefix, task.UserID, task.ID)
	expiration := time.Duration(tm.expireDays) * 24 * time.Hour

	if err := database.RDB.Set(context.Background(), statusKey, statusJSON, expiration).Err(); err != nil {
		middleware.Logger.Errorf("保存任务状态到Redis失败: %v", err)
		return fmt.Errorf("保存任务状态到Redis失败: %v", err)
	}

	return nil
}

// getTaskStatusFromRedis 从Redis获取任务状态
func (tm *TaskManager) getTaskStatusFromRedis(userID int64, taskID string) (*TaskStatusResponse, error) {
	if database.RDB == nil {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	statusKey := fmt.Sprintf("%s%d:%s:status", TaskStatusKeyPrefix, userID, taskID)

	existsCount, err := database.RDB.Exists(context.Background(), statusKey).Result()
	if err != nil {
		return nil, fmt.Errorf("检查Redis键是否存在失败: %v", err)
	}
	if existsCount == 0 {
		return nil, fmt.Errorf("Redis键不存在: %s", statusKey)
	}

	statusJSON, err := database.RDB.Get(context.Background(), statusKey).Result()
	if err != nil {
		middleware.Logger.Errorf("从Redis获取值失败: %v", err)
		return nil, fmt.Errorf("从Redis获取任务状态失败: %v", err)
	}

	var status TaskStatusResponse
	if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
		middleware.Logger.Errorf("解析任务状态JSON失败: %v", err)
		return nil, fmt.Errorf("解析任务状态JSON失败: %v", err)
	}

	return &status, nil
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(userID int64, taskID string) (*Task, error) {
	// 先从内存中查找运行中的任务
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if exists && task.UserID == userID {
		return task, nil
	}

	// 从Redis获取状态信息
	statusKey := fmt.Sprintf("%s%d:%s:status", TaskStatusKeyPrefix, userID, taskID)
	existsCount, err := database.RDB.Exists(context.Background(), statusKey).Result()
	if err != nil {
		return nil, fmt.Errorf("检查Redis键存在失败: %v", err)
	}
	if existsCount == 0 {
		return nil, fmt.Errorf("Redis键不存在: %s", statusKey)
	}

	taskStatus, err := tm.getTaskStatusFromRedis(userID, taskID)
	if err != nil {
		return nil, fmt.Errorf("从Redis获取任务状态失败: %v", err)
	}

	createdAt, _ := time.ParseInLocation("2006-01-02 15:04:05", taskStatus.UpdatedAt.Format("2006-01-02 15:04:05"), cnLoc)
	task = &Task{
		ID:           taskID,
		Type:         taskStatus.Type,
		Status:       taskStatus.Status,
		Progress:     taskStatus.Progress,
		ResultKey:    taskStatus.ResultKey,
		ErrorMessage: taskStatus.ErrorMessage,
		UserID:       userID,
		CreatedAt:    createdAt,
		UpdatedAt:    taskStatus.UpdatedAt,
		ExpiresAt:    createdAt.AddDate(0, 0, tm.expireDays),
		Setup:        taskStatus.Setup,
	}

	return task, nil
}

// GetTasksByFilter 按条件查询任务列表
func (tm *TaskManager) GetTasksByFilter(filterType, filterStatus string, page, pageSize int, userID int64) ([]TaskListItem, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	ctx := context.Background()
	pattern := fmt.Sprintf("%s%d:*:status", TaskStatusKeyPrefix, userID)

	var cursor uint64
	var keys []string
	var err error
	var allKeys []string

	for {
		keys, cursor, err = database.RDB.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			middleware.Logger.Errorf("扫描Redis键失败: %v", err)
			return nil, 0, fmt.Errorf("从Redis获取任务列表失败: %v", err)
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	var filteredItems []TaskListItem
	for _, key := range allKeys {
		keyParts := strings.Split(key, ":")
		if len(keyParts) < 5 {
			continue
		}

		taskID := keyParts[3]

		statusJSON, err := database.RDB.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var status TaskStatusResponse
		if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
			continue
		}

		if filterStatus != "" && status.Status != filterStatus {
			continue
		}

		task, err := tm.GetTask(userID, taskID)
		if err != nil {
			continue
		}

		if filterType != "" && task.Type != filterType {
			continue
		}

		item := TaskListItem{
			ID:        task.ID,
			Type:      task.Type,
			Status:    status.Status,
			CreatedBy: task.CreatedBy,
			UserID:    task.UserID,
			CreatedAt: task.CreatedAt,
		}
		filteredItems = append(filteredItems, item)
	}

	sort.Slice(filteredItems, func(i, j int) bool {
		return filteredItems[i].CreatedAt.After(filteredItems[j].CreatedAt)
	})

	total := len(filteredItems)
	endIndex := offset + pageSize
	if endIndex > total {
		endIndex = total
	}
	startIndex := offset
	if startIndex >= total {
		return []TaskListItem{}, total, nil
	}

	return filteredItems[startIndex:endIndex], total, nil
}

// findTaskByID 从内存或Redis中查找任务
func (tm *TaskManager) findTaskByID(taskID string) (*Task, error) {
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if exists {
		return task, nil
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("%s*:%s:status", TaskStatusKeyPrefix, taskID)

	keys, _, err := database.RDB.Scan(ctx, 0, pattern, 1).Result()
	if err != nil || len(keys) == 0 {
		return nil, fmt.Errorf("任务不存在")
	}

	keyParts := strings.Split(keys[0], ":")
	if len(keyParts) < 3 {
		return nil, fmt.Errorf("任务键格式错误")
	}

	userID, err := strconv.ParseInt(keyParts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析用户ID失败: %v", err)
	}

	return tm.GetTask(userID, taskID)
}

// UpdateTaskProgress 更新任务进度
func (tm *TaskManager) UpdateTaskProgress(taskID string, progress int) error {
	task, err := tm.findTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	task.Progress = progress
	task.UpdatedAt = time.Now().In(cnLoc)

	return tm.saveTaskStatusToRedis(task)
}

// SaveTaskStatus 保存任务状态到Redis（用于执行器更新进度）
func (tm *TaskManager) SaveTaskStatus(task *Task) error {
	task.UpdatedAt = time.Now().In(cnLoc)
	return tm.saveTaskStatusToRedis(task)
}

// UpdateTaskResult 更新任务结果
func (tm *TaskManager) UpdateTaskResult(taskID string, resultData string) error {
	task, err := tm.findTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	resultKey := fmt.Sprintf("%s%d:%s:result", TaskResultKeyPrefix, task.UserID, task.ID)

	if err := database.RDB.Set(context.Background(), resultKey, resultData, time.Duration(tm.expireDays)*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("保存任务结果失败: %v", err)
	}

	task.ResultKey = resultKey
	task.UpdatedAt = time.Now().In(cnLoc)

	return tm.saveTaskStatusToRedis(task)
}
