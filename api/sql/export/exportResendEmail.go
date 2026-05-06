package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// ResendEmailRequest 重新发送邮件请求
type ResendEmailRequest struct {
	ID int64 `json:"id"` // 流程ID
}

// SqlExportResendEmail 重新发送邮件接口
func SqlExportResendEmail(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 从上下文中获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "无法获取用户信息")
		return
	}

	// 解析请求体
	var req ResendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 验证流程ID
	if req.ID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID不能为空")
		return
	}

	// 获取流程信息
	var exportInfo sql.SQLExport
	if err := database.DB.Where("id = ?", req.ID).First(&exportInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询流程失败")
		}
		return
	}

	// 权限检查：只有执行人可以重新发送邮件
	if exportInfo.ExecutorID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "只有执行人可以重新发送邮件")
		return
	}

	// 状态检查：只有邮件发送失败或执行失败的流程才能重新发送
	if exportInfo.Status != sql.StatusFailed {
		middleware.ResponseError(w, http.StatusBadRequest, "只有邮件发送失败的流程才能重新发送邮件")
		return
	}

	middleware.Logger.Infof("开始重新发送邮件，流程ID: %d", req.ID)

	// 获取项目代码
	var projectCode string
	err := database.DB.Table("sys_project_dict").
		Select("project").
		Where("id = ?", exportInfo.Project).
		Row().
		Scan(&projectCode)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取项目代码失败")
		return
	}

	// 异步重新发送邮件
	go func(project, dbName, sqlContent string, exportID int64) {
		// 更新状态为执行中
		if err := database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Update("status", sql.StatusExecuting).Error; err != nil {
			middleware.Logger.Errorf("更新状态为执行中失败: %v", err)
			return
		}

		// 更新Redis
		var updatedExport sql.SQLExport
		if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err == nil {
			if redisErr := sql.UpdateExportRedis(&updatedExport); redisErr != nil {
				middleware.Logger.Errorf("更新Redis失败: %v", redisErr)
			}
		}

		// 重新生成Excel并发送邮件
		emailSent, errMsg := regenerateAndSendEmail(project, dbName, sqlContent, nil, nil, exportID)

		// 更新最终状态
		finalStatus := sql.StatusCompleted
		if !emailSent {
			finalStatus = sql.StatusFailed
			middleware.Logger.Errorf("邮件重新发送失败，流程ID: %d, 错误: %s", exportID, errMsg)
			updateStatusToFailed(exportID, errMsg)
		} else {
			middleware.Logger.Infof("邮件重新发送成功，流程ID: %d", exportID)
			if err := database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Updates(map[string]interface{}{
				"status":           finalStatus,
				"execution_result": "邮件重新发送成功",
			}).Error; err != nil {
				middleware.Logger.Errorf("更新最终状态失败: %v", err)
			}
		}

		// 更新Redis
		if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err == nil {
			if redisErr := sql.UpdateExportRedis(&updatedExport); redisErr != nil {
				middleware.Logger.Errorf("更新Redis失败: %v", redisErr)
			}
		}

		// 删除 Redis 下载 key
		key := middleware.GetSqlExportKey(exportInfo.UrlKey)
		if err := database.RDB.Del(context.Background(), key).Err(); err != nil {
			middleware.Logger.Errorf("删除Redis key失败: %v", err)
		}
	}(projectCode, exportInfo.DatabaseName, exportInfo.SQLContent, exportInfo.ID)

	middleware.ResponseSuccess(w, "邮件重新发送任务已启动", nil)
}

// updateStatusToFailed 更新状态为执行失败
func updateStatusToFailed(exportID int64, errMsg string) {
	if err := database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Updates(map[string]interface{}{
		"status":           sql.StatusFailed,
		"execution_result": fmt.Sprintf("邮件发送失败: %s", errMsg),
	}).Error; err != nil {
		middleware.Logger.Errorf("更新状态为邮件发送失败时出错: %v", err)
	}

	// 更新Redis
	var updatedExport sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err == nil {
		if redisErr := sql.UpdateExportRedis(&updatedExport); redisErr != nil {
			middleware.Logger.Errorf("更新Redis失败: %v", redisErr)
		}
	}
}

// regenerateAndSendEmail 重新生成Excel并发送邮件
func regenerateAndSendEmail(project, dbName, sqlContent string, _, _ []string, exportID int64) (bool, string) {
	// 获取导出记录，拿到 submitterID
	var exportRecord sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportRecord).Error; err != nil {
		return false, fmt.Sprintf("获取导出记录失败: %v", err)
	}

	// 重新执行SQL并生成Excel
	result, err := ExecuteSQL(project, dbName, sqlContent, exportRecord.SubmitterID, exportID)
	if err != nil {
		return false, fmt.Sprintf("SQL执行失败: %v", err)
	}

	if result.FailedCount > 0 {
		errMsg := "执行失败"
		if len(result.Errors) > 0 {
			errMsg = strings.Join(result.Errors, "; ")
		}
		return false, errMsg
	}

	// 发送邮件通知
	if err := SendExportNotification(&exportRecord, true, result.ResultCount, "", "", ""); err != nil {
		return false, fmt.Sprintf("邮件发送失败: %v", err)
	}

	return true, ""
}
