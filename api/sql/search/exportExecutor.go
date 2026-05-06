package search

import (
	"cmdb/api/task"
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SQLExportExecutor SQL查询导出执行器
type SQLExportExecutor struct{}

// SQLExportRequest SQL导出请求参数
type SQLExportRequest struct {
	QueryID string `json:"query_id"`
	Agent   string `json:"agent"`
	DBName  string `json:"db_name"`
}

// RegisterSQLExportExecutor 注册SQL导出执行器
func RegisterSQLExportExecutor() {
	if task.GlobalTaskManager == nil {
		middleware.Logger.Error("TaskManager未初始化，无法注册SQL导出执行器")
		return
	}

	task.GlobalTaskManager.RegisterExecutor("sql_export", &SQLExportExecutor{})
	middleware.Logger.Info("SQL导出执行器注册成功，类型: sql_export")
}

// Execute 实现 TaskExecutor 接口
func (e *SQLExportExecutor) Execute(ctx context.Context, t *task.Task) error {
	middleware.Logger.Infof("开始执行SQL导出任务: %s", t.ID)

	// 步骤1：初始化
	t.Setup.Setup1.Status = "running"
	t.Setup.Setup1.Description = "初始化导出任务"
	t.Setup.Setup1.StartTime = time.Now()
	t.Status = task.TaskStatusRunning
	t.Progress = 0
	t.ProcessedCount = 0
	t.TotalCount = 0
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 解析参数
	var req SQLExportRequest
	if err := json.Unmarshal(t.Params, &req); err != nil {
		return e.failStep(t, 1, fmt.Sprintf("解析参数失败: %v", err))
	}

	if req.QueryID == "" {
		return e.failStep(t, 1, "查询ID不能为空")
	}

	if database.RDB == nil {
		return e.failStep(t, 1, "Redis客户端未初始化")
	}

	e.completeStep(t, 1, "初始化完成")
	t.Progress = 25
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤2：获取查询数据
	t.Setup.Setup2.Status = "running"
	t.Setup.Setup2.Description = "获取查询数据"
	t.Setup.Setup2.StartTime = time.Now()
	task.GlobalTaskManager.SaveTaskStatus(t)

	var cachedResult *sql.QueryResult
	var dbName string

	// 首先尝试执行原始查询获取完整数据
	completeResult, err := ExecuteOriginalQueryForExport(ctx, req.QueryID)
	if err == nil {
		cachedResult = completeResult
		dbName = cachedResult.DBName
		middleware.Logger.Infof("成功通过原始SQL查询获取完整数据，记录数: %d", len(cachedResult.Rows))
	} else {
		// 尝试从缓存获取
		middleware.Logger.Warnf("执行原始查询失败: %v，尝试从缓存获取", err)

		_, cacheDbName, agentName, queryInfoErr := GetQueryInfo(ctx, database.RDB, req.QueryID)
		if queryInfoErr == nil {
			dbName = cacheDbName
			if req.Agent == "" {
				req.Agent = agentName
			}
		}

		var found bool
		cachedResult, found = GetQueryResultFromCacheByID(ctx, database.RDB, req.QueryID)
		if !found {
			return e.failStep(t, 2, "无法获取查询数据，请重新执行查询")
		}
		middleware.Logger.Infof("成功从缓存获取查询结果，记录数: %d", len(cachedResult.Rows))
	}

	if req.DBName == "" {
		req.DBName = dbName
	}

	totalCount := len(cachedResult.Rows)
	t.ProcessedCount = totalCount
	t.TotalCount = totalCount

	e.completeStep(t, 2, fmt.Sprintf("获取数据完成，共%d条", totalCount))
	t.Progress = 50
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤3：生成Excel文件
	t.Setup.Setup3.Status = "running"
	t.Setup.Setup3.Description = "生成Excel文件"
	t.Setup.Setup3.StartTime = time.Now()
	task.GlobalTaskManager.SaveTaskStatus(t)

	searchResult := &sql.SearchResult{
		DBName:  cachedResult.DBName,
		Results: []*sql.QueryResult{cachedResult},
		Took:    cachedResult.Took,
		Total:   cachedResult.Total,
	}

	// 生成Excel（带进度更新）
	fileContent, err := generateExcelWithProgress(searchResult, func(processed int) {
		t.ProcessedCount = processed
		t.Setup.Setup3.Description = fmt.Sprintf("生成Excel文件 (%d/%d)", processed, totalCount)
		task.GlobalTaskManager.SaveTaskStatus(t)
	})
	if err != nil {
		return e.failStep(t, 3, fmt.Sprintf("生成Excel失败: %v", err))
	}

	e.completeStep(t, 3, "Excel生成完成")
	t.Progress = 75
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤4：保存文件和预览数据
	t.Setup.Setup4.Status = "running"
	t.Setup.Setup4.Description = "保存导出文件"
	t.Setup.Setup4.StartTime = time.Now()

	filePath, fileName, fileSize, err := saveSQLExportFile(t.UserID, t.ID, req.DBName, fileContent)
	if err != nil {
		return e.failStep(t, 4, fmt.Sprintf("保存文件失败: %v", err))
	}

	// 保存前100条预览数据到Redis
	previewData := getPreviewRows(cachedResult, 100)
	if previewData != nil {
		previewJSON, err := json.Marshal(map[string]interface{}{
			"columns": cachedResult.Columns,
			"rows":    previewData,
		})
		if err != nil {
			middleware.Logger.Warnf("序列化预览数据失败: %v", err)
		} else {
			if err := task.GlobalTaskManager.UpdateTaskResult(t.ID, string(previewJSON)); err != nil {
				middleware.Logger.Warnf("保存预览数据到Redis失败: %v", err)
			}
		}
	}

	// 更新MySQL
	if err := task.UpdateTaskSuccess(t.ID, filePath, fileName, fileSize, totalCount); err != nil {
		middleware.Logger.Warnf("更新MySQL失败: %v", err)
	}

	e.completeStep(t, 4, "保存完成")
	t.Status = task.TaskStatusSuccess
	t.Progress = 100

	middleware.Logger.Infof("SQL导出任务完成: %s, 文件: %s, 总数: %d", t.ID, filePath, totalCount)
	return nil
}

// failStep 标记步骤失败
func (e *SQLExportExecutor) failStep(t *task.Task, step int, msg string) error {
	now := time.Now()
	switch step {
	case 1:
		t.Setup.Setup1.Status = "failed"
		t.Setup.Setup1.EndTime = now
		t.Setup.Setup1.Duration = now.Sub(t.Setup.Setup1.StartTime).Milliseconds()
		t.Setup.Setup1.Description = msg
	case 2:
		t.Setup.Setup2.Status = "failed"
		t.Setup.Setup2.EndTime = now
		t.Setup.Setup2.Duration = now.Sub(t.Setup.Setup2.StartTime).Milliseconds()
		t.Setup.Setup2.Description = msg
	case 3:
		t.Setup.Setup3.Status = "failed"
		t.Setup.Setup3.EndTime = now
		t.Setup.Setup3.Duration = now.Sub(t.Setup.Setup3.StartTime).Milliseconds()
		t.Setup.Setup3.Description = msg
	case 4:
		t.Setup.Setup4.Status = "failed"
		t.Setup.Setup4.EndTime = now
		t.Setup.Setup4.Duration = now.Sub(t.Setup.Setup4.StartTime).Milliseconds()
		t.Setup.Setup4.Description = msg
	}
	t.Status = task.TaskStatusFailed
	t.ErrorMessage = msg
	return fmt.Errorf(msg)
}

// completeStep 标记步骤完成
func (e *SQLExportExecutor) completeStep(t *task.Task, step int, msg string) {
	now := time.Now()
	switch step {
	case 1:
		t.Setup.Setup1.Status = "success"
		t.Setup.Setup1.EndTime = now
		t.Setup.Setup1.Duration = now.Sub(t.Setup.Setup1.StartTime).Milliseconds()
		t.Setup.Setup1.Description = msg
	case 2:
		t.Setup.Setup2.Status = "success"
		t.Setup.Setup2.EndTime = now
		t.Setup.Setup2.Duration = now.Sub(t.Setup.Setup2.StartTime).Milliseconds()
		t.Setup.Setup2.Description = msg
	case 3:
		t.Setup.Setup3.Status = "success"
		t.Setup.Setup3.EndTime = now
		t.Setup.Setup3.Duration = now.Sub(t.Setup.Setup3.StartTime).Milliseconds()
		t.Setup.Setup3.Description = msg
	case 4:
		t.Setup.Setup4.Status = "success"
		t.Setup.Setup4.EndTime = now
		t.Setup.Setup4.Duration = now.Sub(t.Setup.Setup4.StartTime).Milliseconds()
		t.Setup.Setup4.Description = msg
	}
}

// saveSQLExportFile 保存SQL导出文件
func saveSQLExportFile(userID int64, taskID, dbName string, data []byte) (string, string, int64, error) {
	dateStr := time.Now().Format("20060102")
	dirPath := filepath.Join(".", "export", "sql-apply", fmt.Sprintf("%d", userID), dateStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", "", 0, fmt.Errorf("创建目录失败: %v", err)
	}

	fileName := fmt.Sprintf("sql_%s_%s_%s.xlsx", dbName, taskID[:8], time.Now().Format("150405"))
	fileName = sanitizeFilename(fileName)
	filePath := filepath.Join(dirPath, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", "", 0, fmt.Errorf("写入文件失败: %v", err)
	}

	return filePath, fileName, int64(len(data)), nil
}

// getPreviewRows 获取前N行数据作为预览
func getPreviewRows(result *sql.QueryResult, limit int) [][]interface{} {
	if result == nil || len(result.Rows) == 0 {
		return nil
	}

	if len(result.Rows) <= limit {
		return result.Rows
	}

	return result.Rows[:limit]
}
