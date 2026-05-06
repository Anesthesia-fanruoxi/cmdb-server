package apply

import (
	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// 调度器轮询间隔
const schedulerInterval = 3 * time.Second

var (
	schedulerOnce    sync.Once
	schedulerRunning bool
)

// StartApplyScheduler 启动SQL审批定时执行调度器
func StartApplyScheduler() {
	schedulerOnce.Do(func() {
		schedulerRunning = true
		go runScheduler()
		middleware.Logger.Info("SQL审批定时执行调度器已启动")
	})
}

// runScheduler 运行调度器主循环
func runScheduler() {
	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	// 启动时立即检查一次
	checkAndExecutePendingApplies()

	for schedulerRunning {
		<-ticker.C
		checkAndExecutePendingApplies()
	}
}

// checkAndExecutePendingApplies 检查并执行到期的审批任务
func checkAndExecutePendingApplies() {
	now := time.Now()

	// 查询执行中且已到期的定时任务
	// 条件：status=2(执行中) AND execution_time <= now AND execution_time IS NOT NULL
	var applies []models.SQLApply
	err := database.DB.Where(
		"status = ? AND execution_time IS NOT NULL AND execution_time <= ?",
		models.ApplyStatusExecuting, // 执行中状态
		now,
	).Find(&applies).Error

	if err != nil {
		middleware.Logger.Errorf("查询定时执行任务失败: %v", err)
		return
	}

	if len(applies) == 0 {
		return
	}

	middleware.Logger.Infof("查询到 %d 个到期的定时执行任务", len(applies))

	// 逐个执行
	for i := range applies {
		go executeApply(&applies[i])
	}
}

// executeApply 执行单个审批任务
func executeApply(apply *models.SQLApply) {
	middleware.Logger.Infof("开始执行SQL审批任务 [ID=%d, Project=%s]", apply.ID, apply.Project)

	// 更新状态为执行中
	if err := updateApplyStatus(apply.ID, models.ApplyStatusExecuting); err != nil {
		middleware.Logger.Errorf("更新任务状态失败 [ID=%d]: %v", apply.ID, err)
		return
	}

	// 执行SQL
	result, err := executeSQLViaAgent(apply.Project, apply.DatabaseName, apply.SQLContent)

	// 记录执行时间
	executionTime := time.Now()

	if err != nil {
		// 执行失败
		middleware.Logger.Errorf("SQL执行失败 [ID=%d]: %v", apply.ID, err)
		updateApplyResult(apply.ID, models.ApplyStatusFailed, err.Error(), &executionTime)
		return
	}

	// 执行成功
	middleware.Logger.Infof("SQL执行成功 [ID=%d], 结果: %s", apply.ID, result)
	updateApplyResult(apply.ID, models.ApplyStatusCompleted, result, &executionTime)
}

// executeSQLViaAgent 通过Agent执行SQL
func executeSQLViaAgent(project, dbName, sqlContent string) (string, error) {
	middleware.Logger.Infof("执行SQL请求: project=%s, dbName=%s, sql=%s", project, dbName, sqlContent)

	// 构建请求数据
	requestData := map[string]string{
		"query":  sqlContent,
		"dbName": dbName,
	}

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/sql-plugs/api/sql/execute"
	decryptedData, err := middleware.CallAgentWithBody(project, pluginPath, requestData)
	if err != nil {
		return "", fmt.Errorf("调用Agent失败: %w", err)
	}

	middleware.Logger.Infof("Agent解密后响应: %s", string(decryptedData))

	// 解析解密后的响应
	var agentResponse struct {
		Code    int         `json:"code"`
		Msg     string      `json:"msg"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(decryptedData, &agentResponse); err != nil {
		return "", fmt.Errorf("解析解密数据失败: %v", err)
	}

	// 检查响应状态
	if agentResponse.Code != 200 {
		errorMsg := agentResponse.Msg
		if errorMsg == "" {
			errorMsg = agentResponse.Message
		}
		return "", fmt.Errorf("Agent返回错误: code=%d, msg=%s", agentResponse.Code, errorMsg)
	}

	// 检查message字段是否为error
	if agentResponse.Message == "error" {
		if dataStr, ok := agentResponse.Data.(string); ok {
			return "", fmt.Errorf("SQL执行错误: %s", dataStr)
		}
		return "", fmt.Errorf("SQL执行错误")
	}

	// 返回执行结果
	resultJSON, _ := json.Marshal(agentResponse.Data)
	return string(resultJSON), nil
}

// updateApplyStatus 更新审批状态
func updateApplyStatus(applyID int64, status int) error {
	return database.DB.Model(&models.SQLApply{}).
		Where("id = ?", applyID).
		Update("status", status).Error
}

// updateApplyResult 更新审批执行结果
func updateApplyResult(applyID int64, status int, result string, executionTime *time.Time) error {
	updates := map[string]interface{}{
		"status":                status,
		"execution_result":      result,
		"actual_execution_time": executionTime,
		"executor_status":       models.ApplyRoleCannotProcess, // 执行完成后执行人不能再处理
	}
	return database.DB.Model(&models.SQLApply{}).
		Where("id = ?", applyID).
		Updates(updates).Error
}
