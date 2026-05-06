package search

import (
	"cmdb/api/task"
	"cmdb/models/elfk"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ESExportExecutor ES日志导出执行器
type ESExportExecutor struct{}

// ESExportRequest ES导出请求参数
type ESExportRequest struct {
	Project       string   `json:"project"`
	ViewName      string   `json:"view_name"`
	IndexPattern  string   `json:"index_pattern"`
	TimeField     string   `json:"time_field"`
	StartTime     string   `json:"start_time"`
	EndTime       string   `json:"end_time"`
	Keyword       string   `json:"keyword"`
	Size          int      `json:"size"`
	IncludeFields []string `json:"include_fields"`
}

// RegisterESExportExecutor 注册ES导出执行器
func RegisterESExportExecutor() {
	if task.GlobalTaskManager == nil {
		middleware.Logger.Error("TaskManager未初始化，无法注册ES导出执行器")
		return
	}

	task.GlobalTaskManager.RegisterExecutor("es_export", &ESExportExecutor{})
	middleware.Logger.Info("ES导出执行器注册成功，类型: es_export")
}

// Execute 实现 TaskExecutor 接口
func (e *ESExportExecutor) Execute(_ context.Context, t *task.Task) error {
	middleware.Logger.Infof("开始执行ES导出任务: %s", t.ID)

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
	var req ESExportRequest
	if err := json.Unmarshal(t.Params, &req); err != nil {
		return e.failStep(t, 1, fmt.Sprintf("解析参数失败: %v", err))
	}

	if req.Project == "" || req.IndexPattern == "" {
		return e.failStep(t, 1, "项目或索引不能为空")
	}

	e.completeStep(t, 1, "初始化完成")
	t.Progress = 25
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤2：准备查询
	t.Setup.Setup2.Status = "running"
	t.Setup.Setup2.Description = "准备导出数据"
	t.Setup.Setup2.StartTime = time.Now()
	task.GlobalTaskManager.SaveTaskStatus(t)

	scrollReq := &elfk.ScrollRequest{
		Project:        req.Project,
		ViewName:       req.ViewName,
		IndexPattern:   req.IndexPattern,
		TimeField:      req.TimeField,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Keyword:        req.Keyword,
		Size:           req.Size,
		IncludeFields:  req.IncludeFields,
		UseFieldFilter: len(req.IncludeFields) > 0,
	}

	if scrollReq.Size <= 0 {
		scrollReq.Size = 1000
	}

	e.completeStep(t, 2, "准备完成")
	t.Progress = 50
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤3：执行导出
	t.Setup.Setup3.Status = "running"
	t.Setup.Setup3.Description = "执行数据导出"
	t.Setup.Setup3.StartTime = time.Now()
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 进度回调函数
	progressCallback := func(processed, total int) {
		t.ProcessedCount = processed
		t.TotalCount = total
		task.GlobalTaskManager.SaveTaskStatus(t)
	}

	excelData, totalCount, previewData, err := ExportToExcelWithProgress(scrollReq, progressCallback)
	if err != nil {
		return e.failStep(t, 3, fmt.Sprintf("导出失败: %v", err))
	}

	e.completeStep(t, 3, fmt.Sprintf("导出完成，共%d条", totalCount))
	t.Progress = 75
	t.ProcessedCount = totalCount
	t.TotalCount = totalCount
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤4：保存文件和预览数据
	t.Setup.Setup4.Status = "running"
	t.Setup.Setup4.Description = "保存导出文件"
	t.Setup.Setup4.StartTime = time.Now()

	filePath, fileName, fileSize, err := saveESExportFile(t.UserID, t.ID, req.Project, excelData)
	if err != nil {
		return e.failStep(t, 4, fmt.Sprintf("保存文件失败: %v", err))
	}

	// 保存前100条预览数据到Redis（转换为统一的表格格式）
	if previewData != nil && len(previewData) > 0 {
		// 提取所有列名
		columnsMap := make(map[string]bool)
		for _, row := range previewData {
			for col := range row {
				columnsMap[col] = true
			}
		}

		// 转换为列名数组
		columns := make([]string, 0, len(columnsMap))
		for col := range columnsMap {
			columns = append(columns, col)
		}

		// 转换为行数组
		rows := make([][]interface{}, 0, len(previewData))
		for _, rowMap := range previewData {
			row := make([]interface{}, len(columns))
			for i, col := range columns {
				if val, exists := rowMap[col]; exists {
					row[i] = val
				} else {
					row[i] = nil
				}
			}
			rows = append(rows, row)
		}

		// 统一格式：{columns: [...], rows: [[...], ...]}
		previewResult := map[string]interface{}{
			"columns": columns,
			"rows":    rows,
		}

		previewJSON, err := json.Marshal(previewResult)
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

	middleware.Logger.Infof("ES导出任务完成: %s, 文件: %s, 总数: %d", t.ID, filePath, totalCount)
	return nil
}

// failStep 标记步骤失败
func (e *ESExportExecutor) failStep(t *task.Task, step int, msg string) error {
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
func (e *ESExportExecutor) completeStep(t *task.Task, step int, msg string) {
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

// saveESExportFile 保存ES导出文件
func saveESExportFile(userID int64, taskID, project string, data []byte) (string, string, int64, error) {
	dateStr := time.Now().Format("20060102")
	dirPath := filepath.Join(".", "export", "es_export", fmt.Sprintf("%d", userID), dateStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", "", 0, fmt.Errorf("创建目录失败: %v", err)
	}

	fileName := fmt.Sprintf("es_%s_%s_%s.xlsx", project, taskID[:8], time.Now().Format("150405"))
	filePath := filepath.Join(dirPath, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", "", 0, fmt.Errorf("写入文件失败: %v", err)
	}

	return filePath, fileName, int64(len(data)), nil
}
