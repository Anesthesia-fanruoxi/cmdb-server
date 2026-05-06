package task

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 任务状态常量
const (
	TaskStatusPending  = "pending"
	TaskStatusRunning  = "running"
	TaskStatusSuccess  = "success"
	TaskStatusFailed   = "failed"
	TaskStatusCanceled = "canceled"
)

// Redis键名前缀
const (
	TaskStatusKeyPrefix       = "cmdb:task:"
	TaskResultKeyPrefix       = "cmdb:task:"
	TaskDownloadCodeKeyPrefix = "cmdb:task:download:" // 临时下载码前缀
)

// StepStatus 步骤状态结构
type StepStatus struct {
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int64     `json:"duration"`
	Description string    `json:"description"`
}

// SetupSteps 通用任务步骤结构
type SetupSteps struct {
	Setup1 StepStatus `json:"setup1"`
	Setup2 StepStatus `json:"setup2"`
	Setup3 StepStatus `json:"setup3"`
	Setup4 StepStatus `json:"setup4"`
}

// Task 任务结构
type Task struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	Params         json.RawMessage `json:"params"`
	Progress       int             `json:"progress"`
	ProcessedCount int             `json:"processed_count"` // 已处理数据条数
	TotalCount     int             `json:"total_count"`     // 总数据条数
	ResultKey      string          `json:"result_key"`
	ErrorMessage   string          `json:"error_message"`
	CreatedBy      string          `json:"created_by"`
	UserID         int64           `json:"user_id"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	ExpiresAt      time.Time       `json:"expires_at"`
	Setup          SetupSteps      `json:"setup"`
	ctx            context.Context
	cancel         context.CancelFunc
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	Status       string     `json:"status"`
	Type         string     `json:"type"`
	Progress     int        `json:"progress"`
	ResultKey    string     `json:"result_key"`
	ErrorMessage string     `json:"error_message"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Setup        SetupSteps `json:"setup"`
}

// TaskListItem 任务列表项
type TaskListItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedBy string    `json:"created_by"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Execute(ctx context.Context, task *Task) error
}

// TaskManager 任务管理器
type TaskManager struct {
	executors  map[string]TaskExecutor
	tasks      map[string]*Task
	workerPool chan struct{}
	mu         sync.RWMutex
	maxWorkers int
	expireDays int
}

// 全局任务管理器和时区
var (
	GlobalTaskManager *TaskManager
	cnLoc             = time.FixedZone("CST", 8*3600)
)

// NewTaskManager 创建新的任务管理器
func NewTaskManager(maxWorkers, expireDays int) *TaskManager {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}
	if expireDays <= 0 {
		expireDays = 30 // 默认30天过期
	}

	return &TaskManager{
		executors:  make(map[string]TaskExecutor),
		tasks:      make(map[string]*Task),
		workerPool: make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
		expireDays: expireDays,
	}
}

// RegisterExecutor 注册任务执行器
func (tm *TaskManager) RegisterExecutor(taskType string, executor TaskExecutor) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.executors[taskType] = executor
}

// CreateTask 创建新任务
func (tm *TaskManager) CreateTask(taskType string, params json.RawMessage, createdBy string, userID int64) (*Task, error) {
	tm.mu.RLock()
	executor, exists := tm.executors[taskType]
	tm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未注册的任务类型: %s", taskType)
	}

	if params == nil || len(params) == 0 {
		params = json.RawMessage("{}")
	}

	taskID := uuid.New().String()
	now := time.Now().In(cnLoc)

	task := &Task{
		ID:        taskID,
		Type:      taskType,
		Status:    TaskStatusPending,
		Params:    params,
		Progress:  0,
		CreatedBy: createdBy,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.AddDate(0, 0, tm.expireDays),
		Setup: SetupSteps{
			Setup1: StepStatus{Status: "waiting", Description: "准备步骤"},
			Setup2: StepStatus{Status: "waiting", Description: "执行步骤"},
			Setup3: StepStatus{Status: "waiting", Description: "验证步骤"},
			Setup4: StepStatus{Status: "waiting", Description: "清理步骤"},
		},
	}

	if err := tm.saveTaskStatusToRedis(task); err != nil {
		return nil, fmt.Errorf("保存任务状态失败: %v", err)
	}

	// 保存到MySQL（仅限需要展示的类型）
	if isDisplayableType(taskType) {
		title := getTaskTitle(taskType, params)

		if err := CreateTaskRecord(taskID, taskType, createdBy, userID, title); err != nil {
			middleware.Logger.Warnf("保存任务到MySQL失败: %v", err)
		}
	}

	// 保存参数到Redis
	if len(params) > 0 {
		paramsKey := fmt.Sprintf("%s%d:%s:params", TaskStatusKeyPrefix, userID, taskID)
		expiration := time.Duration(tm.expireDays) * 24 * time.Hour
		database.RDB.Set(context.Background(), paramsKey, string(params), expiration)
	}

	go tm.scheduleTask(task, executor)

	middleware.Logger.Infof("创建任务成功: ID=%s, 类型=%s", taskID, taskType)
	return task, nil
}

// scheduleTask 调度任务执行
func (tm *TaskManager) scheduleTask(task *Task, executor TaskExecutor) {
	taskID := task.ID
	userID := task.UserID

	tm.workerPool <- struct{}{}
	defer func() { <-tm.workerPool }()

	// 从Redis获取最新状态
	statusKey := fmt.Sprintf("%s%d:%s:status", TaskStatusKeyPrefix, userID, taskID)
	if exists, _ := database.RDB.Exists(context.Background(), statusKey).Result(); exists == 0 {
		middleware.Logger.Errorf("Redis键不存在: %s", statusKey)
		return
	}

	statusJSON, err := database.RDB.Get(context.Background(), statusKey).Result()
	if err != nil {
		middleware.Logger.Errorf("从Redis获取任务状态失败: %v", err)
		return
	}

	var status TaskStatusResponse
	if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
		middleware.Logger.Errorf("解析任务状态失败: %v", err)
		return
	}

	if status.Status != TaskStatusPending {
		return
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	task.ctx = ctx
	task.cancel = cancel

	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now().In(cnLoc)
	tm.saveTaskStatusToRedis(task)

	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	// 获取参数
	if task.Params == nil || len(task.Params) == 0 {
		paramsKey := fmt.Sprintf("%s%d:%s:params", TaskStatusKeyPrefix, userID, taskID)
		if paramsJSON, err := database.RDB.Get(context.Background(), paramsKey).Result(); err == nil {
			task.Params = json.RawMessage(paramsJSON)
		}
	}

	// 执行任务
	err = executor.Execute(ctx, task)

	tm.mu.Lock()
	delete(tm.tasks, taskID)
	tm.mu.Unlock()

	if err != nil {
		task.Status = TaskStatusFailed
		task.ErrorMessage = err.Error()
		if isDisplayableType(task.Type) {
			UpdateTaskFailed(taskID, err.Error())
		}
	} else {
		task.Status = TaskStatusSuccess
		task.Progress = 100
	}

	task.UpdatedAt = time.Now().In(cnLoc)
	tm.saveTaskStatusToRedis(task)
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	task, err := tm.findTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	if task.Status != TaskStatusPending && task.Status != TaskStatusRunning {
		return fmt.Errorf("任务状态为 %s，无法取消", task.Status)
	}

	tm.mu.RLock()
	runningTask, exists := tm.tasks[taskID]
	tm.mu.RUnlock()

	if exists && runningTask.cancel != nil {
		runningTask.cancel()
	}

	if task.Status == TaskStatusPending {
		task.Status = TaskStatusCanceled
		task.UpdatedAt = time.Now().In(cnLoc)
		task.ErrorMessage = "任务被手动取消"
		return tm.saveTaskStatusToRedis(task)
	}

	return nil
}

// InitTaskSystem 初始化任务系统
func InitTaskSystem() error {
	if GlobalTaskManager != nil {
		return nil
	}
	GlobalTaskManager = NewTaskManager(10, 30) // 10个工作线程，30天过期
	middleware.Logger.Info("任务管理系统初始化完成")
	return nil
}
