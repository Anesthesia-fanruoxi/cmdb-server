package rule

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// SqlRulesList 获取规则列表
func SqlRulesList(w http.ResponseWriter, _ *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	var rules []sql.SqlRule

	// 直接查询所有规则
	if err := database.DB.Table(middleware.TableSqlRule).Find(&rules).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询规则列表失败："+err.Error())
		return
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", rules)
}
