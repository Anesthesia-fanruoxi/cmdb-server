package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// 导出文件存储根目录（当前目录下的export文件夹）
const ExportBasePath = "./export"

// 业务类型常量
const (
	ExportTypeSQLApproval = "sql-apply" // SQL审批导出
)

// GetExportDir 获取导出目录路径（不含文件名）
func GetExportDir(exportType string, userID int64) string {
	dateStr := time.Now().Format("20060102")
	return filepath.Join(ExportBasePath, exportType, fmt.Sprintf("%d", userID), dateStr)
}

// EnsureExportDir 确保导出目录存在，不存在则创建
func EnsureExportDir(exportType string, userID int64) (string, error) {
	dir := GetExportDir(exportType, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建导出目录失败: %v", err)
	}
	return dir, nil
}
