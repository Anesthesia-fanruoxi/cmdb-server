package ipLocation

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"time"
)

// AuditParams 审计参数
type AuditParams struct {
	RequestIP string
	QueryIP   string
	Status    int
}

// recordIPLocationAudit 记录IP位置查询审计（MySQL入库 + Redis统计）
func recordIPLocationAudit(params AuditParams) {
	// 异步记录，不阻塞主流程
	go func() {
		defer func() {
			if err := recover(); err != nil {
				middleware.Logger.Errorf("记录IP位置查询审计失败: %v", err)
			}
		}()

		// 获取当前日期
		today := time.Now().Format("20060102")

		// MySQL入库 - 简洁审计记录
		audit := system.IPLocationAudit{
			RequestIP: params.RequestIP,
			QueryIP:   params.QueryIP,
			Status:    params.Status,
			CreatedAt: time.Now(),
			Date:      today,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := database.DB.WithContext(ctx).Create(&audit).Error; err != nil {
			middleware.Logger.Errorf("保存IP位置审计记录失败: %v", err)
			return
		}

		// Redis实时统计（用于快速统计）
		redisCtx := context.Background()
		statKey := middleware.GetIPLocationStatKey(today, params.RequestIP)
		database.RDB.Incr(redisCtx, statKey)
		database.RDB.Expire(redisCtx, statKey, 7*24*time.Hour)

		middleware.Logger.Debugf("IP位置查询审计记录成功: ID=%d, 请求IP=%s, 查询IP=%s, 状态=%d",
			audit.ID, params.RequestIP, params.QueryIP, params.Status)
	}()
}
