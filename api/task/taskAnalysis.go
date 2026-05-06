package task

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

// AnalysisResult 存储在Redis中的分析结果结构
type AnalysisResult struct {
	Values map[string]int `json:"values"` // 值和计数的映射
}

// TopValue 表示一个值及其计数
type TopValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// 分析数据预览分页响应
type AnalysisPreviewResponse struct {
	Total      int         `json:"total"`       // 真实总数量（从MySQL获取）
	CacheTotal int         `json:"cache_total"` // Redis缓存数量
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Items      interface{} `json:"items"`
}

// getSortedValues 获取按值排序的结果
func getSortedValues(values map[string]int) []TopValue {
	sortedValues := make([]TopValue, 0, len(values))
	for value, count := range values {
		sortedValues = append(sortedValues, TopValue{Value: value, Count: count})
	}
	// 按值字典序排序
	sort.Slice(sortedValues, func(i, j int) bool {
		return sortedValues[i].Value < sortedValues[j].Value
	})
	return sortedValues
}

// getSortedCounts 获取按计数降序排序的结果
func getSortedCounts(values map[string]int) []TopValue {
	sortedCounts := make([]TopValue, 0, len(values))
	for value, count := range values {
		sortedCounts = append(sortedCounts, TopValue{Value: value, Count: count})
	}
	// 按计数降序排序，计数相同则按值排序
	sort.Slice(sortedCounts, func(i, j int) bool {
		if sortedCounts[i].Count == sortedCounts[j].Count {
			return sortedCounts[i].Value < sortedCounts[j].Value
		}
		return sortedCounts[i].Count > sortedCounts[j].Count
	})
	return sortedCounts
}

// AnalysisPreview 处理分析结果数据预览
// 提供分页功能，仅处理类型为"analysis"的任务
func AnalysisPreview(w http.ResponseWriter, r *http.Request) {
	// 获取用户信息 - 不再重复检查，这部分在TaskPreview中已经处理过了
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 从URL参数获取任务ID
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 分页参数
	page := 1
	pageSize := 20
	sortBy := "count" // 默认按计数排序

	// 解析分页参数
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 排序方式
	if sortByStr := r.URL.Query().Get("sort_by"); sortByStr != "" {
		if sortByStr == "value" || sortByStr == "count" {
			sortBy = sortByStr
		}
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	var task *Task
	var err error

	// 管理员需要先从数据库获取任务的真实用户ID
	if isAdmin {
		dbTask, dbErr := GetTaskByTaskID(taskID)
		if dbErr != nil {
			middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
			return
		}
		// 使用任务的真实用户ID从Redis获取详情
		task, err = GlobalTaskManager.GetTask(dbTask.UserID, taskID)
	} else {
		// 普通用户直接用自己的ID查询
		task, err = GlobalTaskManager.GetTask(userInfo.UserID, taskID)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在: "+err.Error())
		return
	}

	// 普通用户只能查看自己的任务
	if !isAdmin && task.UserID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "无权限查看该任务")
		return
	}

	// 验证任务类型
	if task.Type != "analysis" {
		middleware.ResponseError(w, http.StatusBadRequest, "该任务不是分析任务")
		return
	}

	// 验证任务状态
	if task.Status != TaskStatusSuccess {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("任务状态为 %s，尚未完成", task.Status))
		return
	}

	// 检查是否有结果键
	if task.ResultKey == "" {
		middleware.ResponseError(w, http.StatusNotFound, "任务没有结果数据")
		return
	}

	// 从Redis获取任务结果
	resultData, err := database.RDB.Get(r.Context(), task.ResultKey).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务结果失败: "+err.Error())
		return
	}

	// 解析JSON结果
	var analysisResult AnalysisResult
	if err := json.Unmarshal([]byte(resultData), &analysisResult); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析任务结果失败: "+err.Error())
		return
	}

	// 根据排序方式选择数据
	var items []TopValue
	var total int

	// 动态计算不同的排序结果
	if sortBy == "value" {
		items = getSortedValues(analysisResult.Values)
	} else {
		items = getSortedCounts(analysisResult.Values)
	}

	total = len(items)
	cacheTotal := total

	// 从MySQL获取实际总数
	if record, err := GetTaskByTaskID(taskID); err == nil && record.TotalCount > 0 {
		total = record.TotalCount
	}

	// 分页处理
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= cacheTotal {
		// 页码超出范围，返回空数组
		items = []TopValue{}
	} else if endIdx > cacheTotal {
		// 最后一页可能不足pageSize个元素
		items = items[startIdx:cacheTotal]
	} else {
		items = items[startIdx:endIdx]
	}

	// 构造分页响应
	response := AnalysisPreviewResponse{
		Total:      total,
		CacheTotal: cacheTotal,
		Page:       page,
		PageSize:   pageSize,
		Items:      items,
	}

	middleware.ResponseSuccess(w, "获取分析结果成功", response)
}

// ExportAnalysis 导出分析结果到Excel
// 如果MySQL有文件记录则直接下载，否则从Redis生成
func ExportAnalysis(w http.ResponseWriter, r *http.Request) {
	// 获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 从Context中获取导出参数
	params, ok := r.Context().Value("export_params").(struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	})
	if !ok {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取导出参数失败")
		return
	}

	// 先尝试从MySQL获取任务记录（有文件路径则直接下载）
	record, err := GetTaskByTaskIDAndUser(params.ID, userInfo.UserID)
	if err == nil && record.FilePath != "" {
		// 检查文件是否存在
		fileContent, err := os.ReadFile(record.FilePath)
		if err == nil {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.FileName))
			w.Header().Set("Content-Length", strconv.FormatInt(int64(len(fileContent)), 10))
			w.Write(fileContent)
			middleware.Logger.Infof("从文件下载分析结果: taskID=%s, file=%s", params.ID, record.FileName)
			return
		}
		middleware.Logger.Warnf("文件不存在，尝试从Redis生成: %v", err)
	}

	// 文件不存在，从Redis获取数据重新生成
	task, err := GlobalTaskManager.GetTask(userInfo.UserID, params.ID)
	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在: "+err.Error())
		return
	}

	// 验证用户权限
	if task.UserID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "无权限查看该任务")
		return
	}

	// 验证任务状态
	if task.Status != TaskStatusSuccess {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("任务状态为 %s，尚未完成", task.Status))
		return
	}

	// 检查是否有结果键
	if task.ResultKey == "" {
		middleware.ResponseError(w, http.StatusNotFound, "任务没有结果数据")
		return
	}

	// 从Redis获取任务结果
	resultData, err := database.RDB.Get(r.Context(), task.ResultKey).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务结果失败: "+err.Error())
		return
	}

	// 解析JSON结果
	var analysisResult AnalysisResult
	if err := json.Unmarshal([]byte(resultData), &analysisResult); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析任务结果失败: "+err.Error())
		return
	}

	// 创建新的Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 创建工作表
	sheetName := "分析结果"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"值", "计数"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// 获取排序后的数据
	items := getSortedCounts(analysisResult.Values)

	// 写入数据
	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.Value)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Count)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 30)
	f.SetColWidth(sheetName, "B", "B", 15)

	// 设置响应头
	fileName := fmt.Sprintf("analysis_result_%s_%s.xlsx", params.ID, time.Now().Format("20060102150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	// 写入响应
	if err := f.Write(w); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成Excel文件失败: "+err.Error())
		return
	}
}
