package search

import (
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// 定义操作类型常量
const (
	OperationQuery    = "search"
	OperationPageTurn = "page"
	OperationExport   = "export"
	OperationAnalysis = "analysis"
)

// RecordESSearchAudit 记录ES搜索审计日志
func RecordESSearchAudit(
	r *http.Request, // 请求对象
	queryID string, // 查询ID
	project string, // 项目标识
	viewName string, // 视图名称
	indexPattern string, // 索引模式
	timeField string, // 时间字段
	startTime string, // 开始时间
	endTime string, // 结束时间
	keyword string, // 搜索关键词
	queryTimeMs int64, // 查询耗时
	docCount int, // 文档数量
	operation string, // 操作类型
	page int, // 页码
	responseCode int, // 响应状态码
	err error, // 错误信息
) {
	// 获取用户信息
	userInfo, userOk := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)

	// 准备审计日志记录
	esLog := elfk.SysEsLog{
		QueryID:       queryID,
		Project:       project,
		ViewName:      viewName,
		QIndexPattern: indexPattern,
		QTimeField:    timeField,
		QStartTime:    startTime,
		QEndTime:      endTime,
		QKeyword:      keyword,
		QueryTimeMs:   queryTimeMs,
		DocCount:      docCount,
		Operation:     operation,
		Page:          page,
		ResponseCode:  responseCode,
	}

	// 设置用户相关信息
	if userOk {
		esLog.UserID = userInfo.UserID
		esLog.UserName = userInfo.UserName
		esLog.ClientIP = utils.GetClientIP(r)

		// 获取详细的地理位置信息
		country, region, city, district := utils.GetIPLocationDetails(esLog.ClientIP)
		esLog.Country = country
		esLog.Region = region
		esLog.City = city
		esLog.District = district
	}

	// 记录错误信息
	if err != nil {
		esLog.ErrorMessage = err.Error()
		middleware.Logger.Error("[审计] 错误: %v", err)
	}

	// 写入数据库
	if err := database.DB.Create(&esLog).Error; err != nil {
		middleware.Logger.Error("[审计] 记录ES查询日志失败: %v", err)
	} else {
		middleware.Logger.Info("[审计] 记录ES查询日志成功")
	}
}
