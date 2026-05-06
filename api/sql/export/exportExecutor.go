package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportProgress 导出进度
type ExportProgress struct {
	ExportID    int64  `json:"export_id"`
	Status      string `json:"status"`       // querying/exporting/completed/failed
	TotalRows   int64  `json:"total_rows"`   // 总行数
	CurrentRows int64  `json:"current_rows"` // 当前已写入行数
	Message     string `json:"message"`      // 状态消息
}

// UpdateExportProgress 更新导出进度到Redis
func UpdateExportProgress(exportID int64, progress *ExportProgress) {
	ctx := context.Background()
	key := middleware.GetSqlExportProgressKey(exportID)
	data, _ := json.Marshal(progress)
	database.RDB.Set(ctx, key, string(data), 10*time.Minute)
}

// GetExportProgress 获取导出进度
func GetExportProgress(exportID int64) *ExportProgress {
	ctx := context.Background()
	key := middleware.GetSqlExportProgressKey(exportID)
	data, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	var progress ExportProgress
	json.Unmarshal([]byte(data), &progress)
	return &progress
}

// ExecuteSQL 执行SQL语句并生成Excel，保存到服务器
func ExecuteSQL(project, dbName, sqlContent string, submitterID int64, exportID int64) (*sql.SQLExecuteResult, error) {
	startTime := time.Now()
	middleware.Logger.Infof("开始执行SQL导出，项目: %s, 数据库: %s, 流程ID: %d", project, dbName, exportID)

	result := &sql.SQLExecuteResult{
		SuccessCount: 0,
		FailedCount:  0,
		ResultCount:  0,
		Errors:       []string{},
	}

	// 初始化进度
	UpdateExportProgress(exportID, &ExportProgress{
		ExportID: exportID,
		Status:   "querying",
		Message:  "正在查询数据...",
	})

	// 更新状态为导出中
	database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Update("status", sql.StatusExporting)

	// 确保导出目录存在
	exportDir, err := middleware.EnsureExportDir(middleware.ExportTypeSQLApproval, submitterID)
	if err != nil {
		updateProgressFailed(exportID, fmt.Sprintf("创建目录失败: %v", err))
		updateExportFailed(exportID, fmt.Sprintf("创建导出目录失败: %v", err))
		result.FailedCount++
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%d_%s_%s.xlsx", exportID, project, timestamp)
	filePath := filepath.Join(exportDir, fileName)

	// 执行SQL并导出到Excel（带进度回调）
	rowCount, err := streamExportToExcelWithProgress(sqlContent, dbName, project, filePath, exportID)
	if err != nil {
		updateProgressFailed(exportID, fmt.Sprintf("导出失败: %v", err))
		updateExportFailed(exportID, fmt.Sprintf("导出失败: %v", err))
		result.FailedCount++
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	var fileSize int64
	if fileInfo, err := os.Stat(filePath); err == nil {
		fileSize = fileInfo.Size()
	}

	result.ResultCount = rowCount
	result.SuccessCount = 1
	result.ExcelFilePath = filePath
	result.ExecutionTime = time.Since(startTime).Milliseconds()

	// 更新完成进度
	UpdateExportProgress(exportID, &ExportProgress{
		ExportID:    exportID,
		Status:      "completed",
		TotalRows:   rowCount,
		CurrentRows: rowCount,
		Message:     fmt.Sprintf("导出完成，共%d行", rowCount),
	})

	// 更新数据库记录
	database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Updates(map[string]interface{}{
		"status":           sql.StatusCompleted,
		"file_path":        filePath,
		"file_name":        fileName,
		"file_size":        fileSize,
		"download_limit":   3,
		"execution_result": fmt.Sprintf("导出成功，共%d行，%.2fMB", rowCount, float64(fileSize)/(1024*1024)),
	})

	middleware.Logger.Infof("SQL导出完成，流程ID: %d, 行数: %d, 耗时: %dms", exportID, rowCount, result.ExecutionTime)

	// 获取完整的导出记录用于发送邮件
	var exportRecord sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportRecord).Error; err != nil {
		middleware.Logger.Errorf("获取导出记录失败，无法发送邮件通知: %v", err)
	} else {
		// 自动生成下载链接
		downloadURL, _, expireAt, err := GenerateDownloadLink(exportID)
		if err != nil {
			middleware.Logger.Errorf("自动生成下载链接失败: %v", err)
			downloadURL = "" // 链接生成失败，邮件中不包含链接
		}

		expireAtStr := ""
		if !expireAt.IsZero() {
			expireAtStr = expireAt.Format("2006-01-02 15:04:05")
		}

		// 异步发送成功通知邮件
		go func() {
			if err := SendExportNotification(&exportRecord, true, rowCount, "", downloadURL, expireAtStr); err != nil {
				middleware.Logger.Errorf("发送导出成功通知邮件失败: %v", err)
			}
		}()
	}

	return result, nil
}

func updateProgressFailed(exportID int64, msg string) {
	UpdateExportProgress(exportID, &ExportProgress{
		ExportID: exportID,
		Status:   "failed",
		Message:  msg,
	})
}

func updateExportFailed(exportID int64, errMsg string) {
	database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Updates(map[string]interface{}{
		"status":           sql.StatusFailed,
		"execution_result": errMsg,
	})

	// 获取完整的导出记录用于发送邮件
	var exportRecord sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportRecord).Error; err != nil {
		middleware.Logger.Errorf("获取导出记录失败，无法发送邮件通知: %v", err)
	} else {
		// 异步发送失败通知邮件
		go func() {
			if err := SendExportNotification(&exportRecord, false, 0, errMsg, "", ""); err != nil {
				middleware.Logger.Errorf("发送导出失败通知邮件失败: %v", err)
			}
		}()
	}
}

// streamExportToExcelWithProgress 带进度的导出
func streamExportToExcelWithProgress(sqlQuery, dbName, project, filePath string, exportID int64) (int64, error) {
	// 构建请求体
	requestData := map[string]interface{}{
		"query":  sqlQuery,
		"dbName": dbName,
	}

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/sql-plugs/api/sql/export"
	decryptedData, err := middleware.CallAgentWithBody(project, pluginPath, requestData)
	if err != nil {
		return 0, fmt.Errorf("调用Agent失败: %w", err)
	}

	// 解析响应数据
	var apiResponse struct {
		Code int `json:"code"`
		Data struct {
			Results []struct {
				Columns []string        `json:"columns"`
				Rows    [][]interface{} `json:"rows"`
			} `json:"results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(decryptedData, &apiResponse); err != nil {
		return 0, fmt.Errorf("解析响应数据失败: %w", err)
	}

	if apiResponse.Code != 200 || len(apiResponse.Data.Results) == 0 {
		return 0, nil
	}

	columns := apiResponse.Data.Results[0].Columns
	rows := apiResponse.Data.Results[0].Rows
	totalRows := int64(len(rows))

	// 更新进度：开始写入Excel
	UpdateExportProgress(exportID, &ExportProgress{
		ExportID:    exportID,
		Status:      "exporting",
		TotalRows:   totalRows,
		CurrentRows: 0,
		Message:     fmt.Sprintf("开始写入Excel，共%d行", totalRows),
	})

	return writeExcelWithProgress(columns, rows, filePath, exportID)
}

// writeExcelWithProgress 带进度的Excel写入
func writeExcelWithProgress(columns []string, rows [][]interface{}, filePath string, exportID int64) (int64, error) {
	excelFile := excelize.NewFile()
	sheetName := "查询结果"
	excelFile.NewSheet(sheetName)
	excelFile.DeleteSheet("Sheet1")

	streamWriter, err := excelFile.NewStreamWriter(sheetName)
	if err != nil {
		return 0, fmt.Errorf("创建写入器失败: %w", err)
	}

	// 写入表头
	headerRow := make([]interface{}, len(columns))
	for i, col := range columns {
		headerRow[i] = col
	}
	cellName, _ := excelize.CoordinatesToCellName(1, 1)
	streamWriter.SetRow(cellName, headerRow)

	totalRows := int64(len(rows))
	// 写入数据行，每100行更新一次进度
	for rowIndex, row := range rows {
		rowData := make([]interface{}, len(row))
		for colIndex, val := range row {
			if val == nil {
				rowData[colIndex] = "NULL"
			} else {
				rowData[colIndex] = val
			}
		}
		cellName, _ := excelize.CoordinatesToCellName(1, rowIndex+2)
		streamWriter.SetRow(cellName, rowData)

		// 每100行更新进度
		if (rowIndex+1)%100 == 0 || rowIndex == len(rows)-1 {
			UpdateExportProgress(exportID, &ExportProgress{
				ExportID:    exportID,
				Status:      "exporting",
				TotalRows:   totalRows,
				CurrentRows: int64(rowIndex + 1),
				Message:     fmt.Sprintf("正在写入 %d/%d 行", rowIndex+1, totalRows),
			})
		}
	}

	streamWriter.Flush()
	excelFile.SaveAs(filePath)
	return totalRows, nil
}
