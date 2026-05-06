package search

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// SqlHistoryList 查询SQL历史记录
func SqlHistoryList(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 查询按SQL内容和用户名分组的最新100条记录
	var logs []sql.SqlAuditLog
	if err := database.DB.Model(&sql.SqlAuditLog{}).
		Select("query_sql, MAX(created_at) as created_at, project, db_name").
		Where("user_name = ? and query_sql != '' and operation = 'search'", userInfo.UserName).
		Group("query_sql").
		Order("created_at DESC").
		Limit(40).
		Find(&logs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询历史记录失败")
		return
	}
	// 构建响应数据
	list := make([]sql.SqlHistory, 0, len(logs))
	for _, log := range logs {
		list = append(list, sql.SqlHistory{
			CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
			QuerySql:  log.QuerySql,
		})
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", list)
}
