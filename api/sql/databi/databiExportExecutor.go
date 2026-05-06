package databi

import (
	"cmdb/api/task"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

// DatabiExportExecutor DataBI导出执行器
type DatabiExportExecutor struct{}

// DatabiExportRequest DataBI导出请求参数
type DatabiExportRequest struct {
	Project string `json:"project"`
	Context string `json:"context"`
	Type    string `json:"type"`
}

// RegisterDatabiExportExecutor 注册DataBI导出执行器
func RegisterDatabiExportExecutor() {
	if task.GlobalTaskManager == nil {
		middleware.Logger.Error("TaskManager未初始化，无法注册DataBI导出执行器")
		return
	}

	task.GlobalTaskManager.RegisterExecutor("databi_export", &DatabiExportExecutor{})
	middleware.Logger.Info("DataBI导出执行器注册成功，类型: databi_export")
}

// Execute 实现 TaskExecutor 接口
func (e *DatabiExportExecutor) Execute(_ context.Context, t *task.Task) error {
	middleware.Logger.Infof("开始执行DataBI导出任务: %s", t.ID)

	// 步骤1：初始化
	t.Setup.Setup1.Status = "running"
	t.Setup.Setup1.Description = "初始化导出任务"
	t.Setup.Setup1.StartTime = time.Now()
	t.Status = task.TaskStatusRunning
	t.Progress = 0
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 解析参数
	var req DatabiExportRequest
	if err := json.Unmarshal(t.Params, &req); err != nil {
		return e.failStep(t, 1, fmt.Sprintf("解析参数失败: %v", err))
	}

	if req.Project == "" {
		return e.failStep(t, 1, "项目名称不能为空")
	}

	if req.Context == "" {
		return e.failStep(t, 1, "SQL语句不能为空")
	}

	e.completeStep(t, 1, "初始化完成")
	t.Progress = 20
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 步骤2：执行查询获取数据
	t.Setup.Setup2.Status = "running"
	t.Setup.Setup2.Description = "执行查询获取数据"
	t.Setup.Setup2.StartTime = time.Now()
	task.GlobalTaskManager.SaveTaskStatus(t)

	// 调用Agent执行查询
	pluginPath := "/proxy/databi-plugs/spark/common/query"
	agentRequest := map[string]string{
		"context": req.Context,
		"type":    req.Type,
	}

	decryptedData, err := middleware.CallAgentWithBody(req.Project, pluginPath, agentRequest)
	if err != nil {
		return e.failStep(t, 2, fmt.Sprintf("调用Agent失败: %v", err))
	}

	// 解析响应
	var agentResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Head  []string                 `json:"head"`
			Table []map[string]interface{} `json:"table"`
			File  interface{}              `json:"file"`
		} `json:"data"`
	}

	if err := json.Unmarshal(decryptedData, &agentResp); err != nil {
		return e.failStep(t, 2, fmt.Sprintf("解析Agent响应失败: %v", err))
	}

	if agentResp.Code != 200 {
		return e.failStep(t, 2, fmt.Sprintf("Agent返回错误: %s", agentResp.Msg))
	}

	totalCount := len(agentResp.Data.Table)
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

	// 生成Excel
	fileContent, err := e.generateExcel(agentResp.Data.Head, agentResp.Data.Table, func(processed int) {
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

	// 步骤4：保存文件
	t.Setup.Setup4.Status = "running"
	t.Setup.Setup4.Description = "保存导出文件"
	t.Setup.Setup4.StartTime = time.Now()

	filePath, fileName, fileSize, err := e.saveDatabiExportFile(t.UserID, t.ID, req.Project, fileContent)
	if err != nil {
		return e.failStep(t, 4, fmt.Sprintf("保存文件失败: %v", err))
	}

	// 保存前100条预览数据到Redis
	previewData := e.getPreviewRows(agentResp.Data.Table, 100)
	if previewData != nil {
		previewJSON, err := json.Marshal(map[string]interface{}{
			"columns": agentResp.Data.Head,
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

	middleware.Logger.Infof("DataBI导出任务完成: %s, 文件: %s, 总数: %d", t.ID, filePath, totalCount)
	return nil
}

// generateExcel 生成Excel文件
func (e *DatabiExportExecutor) generateExcel(head []string, table []map[string]interface{}, progressCallback func(processed int)) ([]byte, error) {
	if len(table) == 0 {
		return nil, fmt.Errorf("查询结果为空")
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// 写入表头
	for i, col := range head {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col)
	}

	// 写入数据行
	for rowIndex, row := range table {
		for colIndex, colName := range head {
			value, ok := row[colName]
			if !ok {
				value = ""
			}
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			f.SetCellValue(sheetName, cell, value)
		}

		// 每处理1000行或最后一行时回调
		if (rowIndex+1)%1000 == 0 || rowIndex == len(table)-1 {
			if progressCallback != nil {
				progressCallback(rowIndex + 1)
			}
		}
	}

	// 设置列宽
	for i := range head {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0EBF5"}, Pattern: 1},
	})
	colName, _ := excelize.ColumnNumberToName(len(head))
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", colName), headerStyle)

	var buf []byte
	var err error
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("保存Excel文件失败: %v", err)
	}
	buf = buffer.Bytes()

	return buf, nil
}

// saveDatabiExportFile 保存DataBI导出文件
func (e *DatabiExportExecutor) saveDatabiExportFile(userID int64, taskID, projectName string, data []byte) (string, string, int64, error) {
	dateStr := time.Now().Format("20060102")
	dirPath := filepath.Join(".", "export", "databi", fmt.Sprintf("%d", userID), dateStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", "", 0, fmt.Errorf("创建目录失败: %v", err)
	}

	fileName := fmt.Sprintf("databi_%s_%s_%s.xlsx", projectName, taskID[:8], time.Now().Format("150405"))
	filePath := filepath.Join(dirPath, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", "", 0, fmt.Errorf("写入文件失败: %v", err)
	}

	return filePath, fileName, int64(len(data)), nil
}

// getPreviewRows 获取前N行数据作为预览
func (e *DatabiExportExecutor) getPreviewRows(table []map[string]interface{}, limit int) []map[string]interface{} {
	if len(table) == 0 {
		return nil
	}

	if len(table) <= limit {
		return table
	}

	return table[:limit]
}

// failStep 标记步骤失败
func (e *DatabiExportExecutor) failStep(t *task.Task, step int, msg string) error {
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
func (e *DatabiExportExecutor) completeStep(t *task.Task, step int, msg string) {
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
