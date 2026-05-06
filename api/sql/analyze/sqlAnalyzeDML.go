package analyze

import (
	"regexp"
	"strings"

	models "cmdb/models/sql"
)

// AnalyzeDML 分析DML语句
func AnalyzeDML(sql, sqlType string) *models.AnalyzeResponse {
	normalized := NormalizeSQL(sql)

	result := &models.AnalyzeResponse{
		BasicInfo: &models.ResponseBasicInfo{
			SQLType:     sqlType,
			SQLCategory: "DML",
		},
	}

	detail := &models.DMLDetail{
		Operation: sqlType,
		HasWhere:  HasClause(normalized, "WHERE"),
		HasLimit:  HasClause(normalized, "LIMIT"),
	}

	switch sqlType {
	case "INSERT":
		parseInsert(normalized, detail)
	case "UPDATE":
		parseUpdate(normalized, detail)
	case "DELETE":
		parseDelete(normalized, detail)
	}

	result.Analyze = detail
	evaluateDMLRisk(result, detail)

	return result
}

// parseInsert 解析INSERT语句
func parseInsert(sql string, detail *models.DMLDetail) {
	// 提取表名（支持 database.table 格式）
	re := regexp.MustCompile(`(?i)INSERT\s+(?:INTO\s+)?` + "`?" + `([\w]+)` + "`?" + `(?:\.` + "`?" + `([\w]+)` + "`?" + `)?`)
	if m := re.FindStringSubmatch(sql); len(m) > 1 {
		if m[2] != "" {
			detail.Table = m[1] + "." + m[2] // database.table
		} else {
			detail.Table = m[1]
		}
	}

	// 提取字段
	colRe := regexp.MustCompile(`(?i)INSERT\s+(?:INTO\s+)?` + "`?" + `\w+` + "`?" + `\s*\(([^)]+)\)`)
	if m := colRe.FindStringSubmatch(sql); len(m) > 1 {
		cols := strings.Split(m[1], ",")
		for _, c := range cols {
			c = strings.Trim(strings.TrimSpace(c), "`")
			if c != "" {
				detail.Columns = append(detail.Columns, c)
			}
		}
	}

	// 判断数据来源
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, " SELECT ") {
		detail.DataSource = "SELECT"
	} else if strings.Contains(upper, " VALUES") || strings.Contains(upper, " VALUE") {
		detail.DataSource = "VALUES"
		detail.RowCount = countInsertRows(sql)
		detail.Values = extractInsertValues(sql)
	}
}

// parseUpdate 解析UPDATE语句
func parseUpdate(sql string, detail *models.DMLDetail) {
	// 提取表名（支持 database.table 格式）
	re := regexp.MustCompile(`(?i)UPDATE\s+` + "`?" + `([\w]+)` + "`?" + `(?:\.` + "`?" + `([\w]+)` + "`?" + `)?`)
	if m := re.FindStringSubmatch(sql); len(m) > 1 {
		if m[2] != "" {
			detail.Table = m[1] + "." + m[2] // database.table
		} else {
			detail.Table = m[1]
		}
	}

	// 提取SET字段和值
	setRe := regexp.MustCompile(`(?i)SET\s+(.+?)(?:\s+WHERE|\s+ORDER|\s+LIMIT|$)`)
	if m := setRe.FindStringSubmatch(sql); len(m) > 1 {
		parts := splitSetClauses(m[1])
		for _, p := range parts {
			if idx := strings.Index(p, "="); idx > 0 {
				col := strings.TrimSpace(p[:idx])
				col = strings.Trim(col, "`")
				if dotIdx := strings.LastIndex(col, "."); dotIdx >= 0 {
					col = col[dotIdx+1:]
				}
				if col != "" {
					detail.Columns = append(detail.Columns, col)
					value := strings.TrimSpace(p[idx+1:])
					setClause := models.SetClause{
						Column: col,
						Value:  value,
						IsExpr: isExpression(value),
					}
					detail.SetClauses = append(detail.SetClauses, setClause)
				}
			}
		}
	}

	if detail.HasWhere {
		detail.WhereClause = extractWhereClause(sql)
	}
}

// parseDelete 解析DELETE语句
func parseDelete(sql string, detail *models.DMLDetail) {
	// 提取表名（支持 database.table 格式）
	re := regexp.MustCompile(`(?i)DELETE\s+FROM\s+` + "`?" + `([\w]+)` + "`?" + `(?:\.` + "`?" + `([\w]+)` + "`?" + `)?`)
	if m := re.FindStringSubmatch(sql); len(m) > 1 {
		if m[2] != "" {
			detail.Table = m[1] + "." + m[2] // database.table
		} else {
			detail.Table = m[1]
		}
	}

	if detail.HasWhere {
		detail.WhereClause = extractWhereClause(sql)
	}
}

// extractWhereClause 提取WHERE条件
func extractWhereClause(sql string) string {
	re := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:\s+ORDER|\s+LIMIT|\s+GROUP|$)`)
	if m := re.FindStringSubmatch(sql); len(m) > 1 {
		where := strings.TrimSpace(m[1])
		if len(where) > 100 {
			where = where[:100] + "..."
		}
		return where
	}
	return ""
}

// countInsertRows 统计INSERT行数
func countInsertRows(sql string) int {
	upper := strings.ToUpper(sql)
	idx := strings.Index(upper, " VALUES")
	if idx < 0 {
		idx = strings.Index(upper, " VALUE")
	}
	if idx < 0 {
		return 0
	}

	valuesPart := sql[idx:]
	count := 0
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(valuesPart); i++ {
		c := valuesPart[i]
		if c == '\'' || c == '"' {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar && (i == 0 || valuesPart[i-1] != '\\') {
				inString = false
			}
			continue
		}
		if inString {
			continue
		}
		if c == '(' {
			if depth == 0 {
				count++
			}
			depth++
		} else if c == ')' {
			depth--
		}
	}
	return count
}

// evaluateDMLRisk 评估DML风险
func evaluateDMLRisk(result *models.AnalyzeResponse, detail *models.DMLDetail) {
	switch detail.Operation {
	case "INSERT":
		result.BasicInfo.RiskLevel = "safe"
		result.BasicInfo.RiskReason = "INSERT操作，新增数据"
		if detail.DataSource == "SELECT" {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "INSERT...SELECT，数据量取决于SELECT结果"
		} else if detail.RowCount > 100 {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "批量INSERT，插入大量数据"
		}
	case "UPDATE":
		if !detail.HasWhere {
			result.BasicInfo.RiskLevel = "danger"
			result.BasicInfo.RiskReason = "UPDATE无WHERE条件，将更新全表数据！"
		} else {
			// 根据WHERE条件类型判断风险
			riskLevel, riskReason := evaluateWhereRisk(detail.WhereClause, "UPDATE")
			result.BasicInfo.RiskLevel = riskLevel
			result.BasicInfo.RiskReason = riskReason
		}
		if detail.HasLimit && result.BasicInfo.RiskLevel == "danger" {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "UPDATE有LIMIT限制"
		}
	case "DELETE":
		if !detail.HasWhere {
			result.BasicInfo.RiskLevel = "danger"
			result.BasicInfo.RiskReason = "DELETE无WHERE条件，将删除全表数据！"
		} else {
			// 根据WHERE条件类型判断风险
			riskLevel, riskReason := evaluateWhereRisk(detail.WhereClause, "DELETE")
			result.BasicInfo.RiskLevel = riskLevel
			result.BasicInfo.RiskReason = riskReason
		}
		if detail.HasLimit && result.BasicInfo.RiskLevel == "danger" {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "DELETE有LIMIT限制"
		}
	}
}

// evaluateWhereRisk 根据WHERE条件类型评估风险
// 返回风险等级和风险原因
func evaluateWhereRisk(whereClause string, operation string) (string, string) {
	if whereClause == "" {
		return "danger", operation + "无WHERE条件"
	}

	upper := strings.ToUpper(whereClause)

	// 检查是否包含范围条件（中风险）
	rangeOperators := []string{
		" > ", " < ", " >= ", " <= ", " != ", " <> ",
		" BETWEEN ", " LIKE ", " IN ", " NOT IN ",
		" IS NULL", " IS NOT NULL",
	}

	for _, op := range rangeOperators {
		if strings.Contains(upper, op) {
			return "warning", operation + "使用范围条件，可能影响多行数据"
		}
	}

	// 只有精确匹配（=）则为低风险，不返回reason
	if strings.Contains(whereClause, "=") {
		return "safe", ""
	}

	// 默认中风险
	return "warning", operation + "操作，请确认WHERE条件"
}
