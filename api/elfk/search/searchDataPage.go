package search

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// ESSearchPageRequest ES搜索分页请求结构体
type ESSearchPageRequest struct {
	QueryID string `json:"query_id"` // 查询ID，必须提供，用于标识查询
	Page    int    `json:"page"`     // 页码，从1开始
	LogType string `json:"log_type"` // 日志类型: elfk 或 sls
}

// ESSearchDataPage 处理搜索分页请求的统一入口
func ESSearchDataPage(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 记录开始时间，用于计算执行时间
	startTime := time.Now()

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// 解析JSON
	var pageReq ESSearchPageRequest
	if err := json.Unmarshal(body, &pageReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 验证必填字段
	if pageReq.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query_id字段不能为空")
		return
	}

	// 设置默认分页参数
	if pageReq.Page <= 0 {
		pageReq.Page = 1
	}

	// 检查Redis是否已初始化
	if database.RDB == nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "Redis客户端未初始化")
		return
	}

	// 获取关联查询的完整信息
	project, viewName, indexPattern, timeField, _, queryStartTime, queryEndTime, keyword, infoErr := GetQueryInfoByID(r.Context(), pageReq.QueryID)
	if infoErr != nil {
		middleware.Logger.Warnf("获取查询信息失败: %v", infoErr)
		// 继续执行，不中断分析过程
	}

	// 如果类型未指定，默认为elfk
	if pageReq.LogType == "" {
		pageReq.LogType = "elfk"
	}

	// 根据LogType路由到相应的处理函数
	var pagedResult *ESSearchResult
	ctx := r.Context()

	switch pageReq.LogType {
	case "sls":
		// 调用SLS分页处理函数
		middleware.Logger.Infof("使用SLS翻页处理: QID=%s, Page=%d", pageReq.QueryID, pageReq.Page)
		pagedResult, err = HandleSLSPagedSearch(pageReq, project, indexPattern, queryStartTime, queryEndTime, keyword)
	default:
		// 默认使用ELFK缓存分页处理
		middleware.Logger.Infof("使用ELFK缓存翻页处理: QID=%s, Page=%d", pageReq.QueryID, pageReq.Page)
		pagedResult, err = HandleESPagedSearch(ctx, pageReq)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 计算执行时间
	executionTime := time.Since(startTime).Milliseconds()

	// 记录审计日志（翻页操作）
	RecordESSearchAudit(
		r,
		pageReq.QueryID, // 使用请求中的查询ID
		project,         // 使用查询到的项目
		viewName,        // 使用查询到的视图名
		indexPattern,    // 使用查询到的索引模式
		timeField,       // 使用查询到的时间字段
		queryStartTime,  // 使用查询到的开始时间
		queryEndTime,    // 使用查询到的结束时间
		keyword,         // 使用查询到的关键词
		executionTime,
		len(pagedResult.Hits),
		OperationPageTurn,
		pageReq.Page,
		http.StatusOK,
		nil,
	)

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", pagedResult)
}
