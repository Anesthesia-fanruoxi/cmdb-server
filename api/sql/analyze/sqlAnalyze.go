package analyze

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	models "cmdb/models/sql"
	"cmdb/pkg/middleware"
)

// SQLAnalyzeHandler 处理SQL分析请求
func SQLAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许POST请求")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求失败")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭请求体失败: %v", closeErr)
		}
	}()

	var req models.AnalyzeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败")
		return
	}

	if strings.TrimSpace(req.SQL) == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "SQL不能为空")
		return
	}

	// 分析SQL，统一返回数组格式
	results := AnalyzeMultiSQL(req.SQL)
	middleware.ResponseSuccess(w, "分析成功", results)
}

// AnalyzeMultiSQL 分析多条SQL语句
func AnalyzeMultiSQL(sql string) []*models.AnalyzeResponse {
	var results []*models.AnalyzeResponse

	// 拆分多条SQL
	sqls := SplitMultiSQL(sql)

	for _, s := range sqls {
		result := AnalyzeSQL(s)
		results = append(results, result)
	}

	return results
}

// AnalyzeSQL 分析单条SQL语句
func AnalyzeSQL(sql string) *models.AnalyzeResponse {
	// 获取SQL类型和分类
	sqlType := GetSQLType(sql)
	category := GetSQLCategory(sqlType)

	// 根据分类调用对应的分析器
	switch category {
	case "DQL":
		return AnalyzeDQL(sql, sqlType)
	case "DML":
		return AnalyzeDML(sql, sqlType)
	case "DDL":
		return AnalyzeDDL(sql, sqlType)
	default:
		return analyzeOther(sql, sqlType)
	}
}

// analyzeOther 分析其他类型SQL
func analyzeOther(_ string, sqlType string) *models.AnalyzeResponse {
	return &models.AnalyzeResponse{
		BasicInfo: &models.ResponseBasicInfo{
			SQLType:     sqlType,
			SQLCategory: "OTHER",
			RiskLevel:   "safe",
			RiskReason:  "辅助操作",
		},
		Analyze: nil,
	}
}
