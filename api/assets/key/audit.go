package key

import (
	"cmdb/models"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"net/http"
	"time"
)

// RecordAudit 记录加解密操作审计
func RecordAudit(r *http.Request, operationType string, project string, batchCount int, status string) {
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)

	// 创建审计日志记录
	auditLog := models.EncryptDecryptAudit{
		Project:       project,
		OperationType: operationType,
		BatchCount:    batchCount,
		Status:        status,
		IPAddress:     utils.GetClientIP(r),
	}

	if ok {
		auditLog.UserID = int(userInfo.UserID)
		auditLog.UserName = userInfo.UserName
	}

	// 异步保存审计日志
	go func(log models.EncryptDecryptAudit) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := database.DB.WithContext(ctx).Create(&log).Error; err != nil {
			// 记录错误，但不影响主流程
			middleware.Logger.Errorf("记录加解密审计日志失败: 项目=%s, 操作=%s, 错误=%v", log.Project, log.OperationType, err)
		}
	}(auditLog)
}
