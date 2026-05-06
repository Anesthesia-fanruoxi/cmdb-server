package rule

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// RuleCheckRequest SQL规则检查请求
type RuleCheckRequest struct {
	SQLContent string `json:"sql"`                // SQL内容
	Project    string `json:"project,omitempty"`  // 项目名称（用于估算影响行数）
	Database   string `json:"database,omitempty"` // 数据库名称（用于估算影响行数）
}

// RuleCheckAPI 对外的SQL规则检查接口
func RuleCheckAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 解析请求
	var req RuleCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 验证SQL内容
	if req.SQLContent == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "SQL内容不能为空")
		return
	}

	// 执行规则检查（传入 project 和 database 用于估算影响行数）
	result, err := CheckSQLRulesWithEstimate(req.SQLContent, req.Project, req.Database)
	if err != nil {
		middleware.Logger.Errorf("SQL规则检查失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "SQL规则检查失败: "+err.Error())
		return
	}

	middleware.Logger.Infof("SQL规则检查完成，有违规: %v, 有阻断: %v",
		result.HasViolation, result.HasBlocker)

	// 返回检查结果
	middleware.ResponseSuccess(w, "SQL规则检查完成", result)
}
