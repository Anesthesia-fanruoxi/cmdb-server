package search

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// AnalysisItem 分析项
type AnalysisItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// getSortedAnalysisItems 获取排序后的分析项（按计数降序）
func getSortedAnalysisItems(values map[string]int) []AnalysisItem {
	items := make([]AnalysisItem, 0, len(values))
	for value, count := range values {
		items = append(items, AnalysisItem{Value: value, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Value < items[j].Value
		}
		return items[i].Count > items[j].Count
	})
	return items
}

// getPreviewResult 获取前N条预览数据
func getPreviewResult(result *AnalysisResult, limit int) *AnalysisResult {
	if len(result.Values) <= limit {
		return result
	}

	// 获取排序后的数据
	items := getSortedAnalysisItems(result.Values)

	// 只取前limit条
	previewValues := make(map[string]int)
	for i := 0; i < limit && i < len(items); i++ {
		previewValues[items[i].Value] = items[i].Count
	}

	return &AnalysisResult{Values: previewValues}
}

// saveAnalysisToExcel 保存分析结果到Excel文件
func saveAnalysisToExcel(userID int64, taskID string, result *AnalysisResult) (string, string, int64, error) {
	// 创建目录: ./export/analysis/{user_id}/{年月日}/
	dateStr := time.Now().Format("20060102")
	dirPath := filepath.Join(".", "export", "analysis", fmt.Sprintf("%d", userID), dateStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", "", 0, fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名
	fileName := fmt.Sprintf("analysis_%s_%s.xlsx", taskID[:8], time.Now().Format("150405"))
	filePath := filepath.Join(dirPath, fileName)

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "分析结果"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	f.SetCellValue(sheetName, "A1", "值")
	f.SetCellValue(sheetName, "B1", "计数")

	// 获取排序后的数据（按计数降序）
	items := getSortedAnalysisItems(result.Values)

	// 写入数据
	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.Value)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Count)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 15)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return "", "", 0, fmt.Errorf("保存Excel文件失败: %v", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return filePath, fileName, 0, nil
	}

	return filePath, fileName, fileInfo.Size(), nil
}
