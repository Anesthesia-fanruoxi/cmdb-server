package search

import (
	"bytes"
	"cmdb/api/task"
	"cmdb/models/sql"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xuri/excelize/v2"
)

// SqlExportData 创建SQL异步导出任务（统一走任务中心）
func SqlExportData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	var exportReq struct {
		QueryID string `json:"query_id"`
		Agent   string `json:"agent"`
		DBName  string `json:"db_name"`
	}

	if err := json.Unmarshal(body, &exportReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	if exportReq.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query_id字段不能为空")
		return
	}

	// 验证项目权限
	if exportReq.Agent != "" {
		projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !isSuperAdmin && len(projects) == 0 {
			middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
			return
		}
		if !isSuperAdmin && !utils.Contains(projects, exportReq.Agent) {
			middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
			return
		}
	}

	// 构建任务参数
	taskParams, err := json.Marshal(map[string]interface{}{
		"query_id": exportReq.QueryID,
		"agent":    exportReq.Agent,
		"db_name":  exportReq.DBName,
	})
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化参数失败")
		return
	}

	// 创建异步导出任务
	exportTask, err := task.GlobalTaskManager.CreateTask(
		"sql_export",
		json.RawMessage(taskParams),
		userInfo.UserName,
		userInfo.UserID,
	)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("创建导出任务失败: %v", err))
		return
	}

	// 记录审计日志
	RecordSqlAudit(r, exportReq.Agent, exportReq.DBName,
		fmt.Sprintf("创建导出任务: %s", exportReq.QueryID), exportReq.QueryID,
		0, "success", "", OperationExport, 1, "")

	middleware.ResponseSuccess(w, "导出任务已创建", map[string]interface{}{
		"task_id": exportTask.ID,
		"message": "导出任务已创建，请在任务中心查看进度",
	})
}

// sanitizeFilename 清理文件名中的特殊字符
func sanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", ";", "_", " ", "_",
	)
	return replacer.Replace(filename)
}

// generateExcelWithProgress 生成Excel格式的导出文件（带进度回调）
func generateExcelWithProgress(result *sql.SearchResult, progressCallback func(processed int)) ([]byte, error) {
	if result == nil || len(result.Results) == 0 {
		return nil, fmt.Errorf("查询结果为空")
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	columns := result.Results[0].Columns

	// 写入表头
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col)
	}

	// 写入数据行（带进度回调）
	batchSize := 1000 // 每1000行回调一次
	for rowIndex, row := range result.Results[0].Rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			f.SetCellValue(sheetName, cell, value)
		}

		// 每处理 batchSize 行或最后一行时回调
		if (rowIndex+1)%batchSize == 0 || rowIndex == len(result.Results[0].Rows)-1 {
			if progressCallback != nil {
				progressCallback(rowIndex + 1)
			}
		}
	}

	// 设置列宽
	for i := range columns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0EBF5"}, Pattern: 1},
	})
	colName, _ := excelize.ColumnNumberToName(len(columns))
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", colName), headerStyle)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return buf.Bytes(), nil
}
