package search

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"net/http"
	"time"
)

// RecordSqlAudit 记录SQL审计日志
func RecordSqlAudit(r *http.Request, project, dbName, querySql, queryID string, executionTime int64, status string, errorMsg string, operation string, page int, platform string, affectedRows ...int) {
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)

	country, region, city, district := utils.GetIPLocationDetails(utils.GetClientIP(r))

	// 创建审计日志记录
	auditLog := sql.SqlAuditLog{
		Project:       project,
		DBName:        dbName,
		QuerySql:      querySql,
		QueryID:       queryID,
		ExecutionTime: int(executionTime),
		Status:        status,
		ErrorMessage:  errorMsg,
		ClientIP:      utils.GetClientIP(r),
		Country:       country,
		Region:        region,
		City:          city,
		District:      district,
		Operation:     operation, // 操作类型
		Page:          page,      // 页码
		Platform:      platform,  // 平台标识
	}

	// 如果提供了影响行数，直接设置
	if len(affectedRows) > 0 && affectedRows[0] > 0 {
		auditLog.AffectedRows = affectedRows[0]
	}

	if ok {
		auditLog.UserID = userInfo.UserID
		auditLog.UserName = userInfo.UserName
	}

	// 异步保存审计日志
	go func(log sql.SqlAuditLog) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := database.DB.WithContext(ctx).Create(&log).Error; err != nil {
			// 记录错误，但不影响主流程
			middleware.Logger.Errorf("记录SQL审计日志失败: queryID=%s, error=%v", log.QueryID, err)
		}
	}(auditLog)
}
