package search

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// ContextQuery 上下文查询参数
type ContextQuery struct {
	Project      string        `json:"project"`       // 项目标识符
	IndexPattern string        `json:"index_pattern"` // 索引名称或模式
	TimeField    string        `json:"time_field"`    // 时间字段名称
	DocID        string        `json:"doc_id"`        // 中心文档ID
	Timestamp    int64         `json:"timestamp"`     // 文档时间戳
	LinesBefore  int           `json:"lines_before"`  // 获取中心文档之前的文档数
	LinesAfter   int           `json:"lines_after"`   // 获取中心文档之后的文档数
	Query        string        `json:"query"`         // 查询字符串
	Filters      []interface{} `json:"filters"`       // 过滤条件
	LogType      string        `json:"log_type"`      // 日志类型：elfk或sls
}

// ContextResult 上下文查询结果
type ContextResult struct {
	Before      []SearchHit `json:"before"`       // 中心文档之前的文档
	Center      *SearchHit  `json:"center"`       // 中心文档
	After       []SearchHit `json:"after"`        // 中心文档之后的文档
	Total       int         `json:"total"`        // 文档总数（前+中+后）
	BeforeTotal int         `json:"before_total"` // 前文档总数
	AfterTotal  int         `json:"after_total"`  // 后文档总数
	Took        int         `json:"took"`         // 查询耗时(毫秒)
}

// SearchHit 搜索命中的单个文档
type SearchHit struct {
	Index  string                 `json:"_index"`  // 索引名称
	ID     string                 `json:"_id"`     // 文档ID
	Score  float64                `json:"_score"`  // 相关性评分
	Source map[string]interface{} `json:"_source"` // 文档内容
}

// EsSearchContext 处理上下文查询请求
func EsSearchContext(w http.ResponseWriter, r *http.Request) {
	// 开始时间，用于计算处理时间
	startTime := time.Now()

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败: "+err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// 解析JSON请求
	var contextReq map[string]interface{}
	if err := json.Unmarshal(body, &contextReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}

	// 获取项目参数
	project, ok := contextReq["project"].(string)
	if !ok || project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果不是超级管理员且没有权限访问该项目，返回无权限错误
	if !isSuperAdmin && !utils.Contains(projects, project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 获取log_type参数，默认为elfk
	logType, _ := contextReq["log_type"].(string)
	if logType == "" {
		logType = "elfk" // 默认使用elfk
	}

	// 根据日志类型选择不同的处理方法
	switch logType {
	case "sls":
		// 调用SLS上下文查询处理方法
		SearchSLSContext(w, body, startTime)
	case "elfk":
		// 调用ELFK上下文查询处理方法
		SearchELFKContext(w, body, startTime)
	default:
		middleware.ResponseError(w, http.StatusBadRequest, "不支持的日志类型: "+logType)
	}
}
