package analyze

import (
	"regexp"
	"strings"

	models "cmdb/models/sql"
)

// AnalyzeDQL 分析DQL语句
func AnalyzeDQL(sql, sqlType string) *models.AnalyzeResponse {
	normalized := NormalizeSQL(sql)

	result := &models.AnalyzeResponse{
		BasicInfo: &models.ResponseBasicInfo{
			SQLType:     sqlType,
			SQLCategory: "DQL",
		},
	}

	detail := &models.DQLDetail{
		SelectType:  "SIMPLE",
		HasWhere:    HasClause(normalized, "WHERE"),
		HasJoin:     hasJoin(normalized),
		HasGroup:    HasClause(normalized, "GROUP BY"),
		HasOrder:    HasClause(normalized, "ORDER BY"),
		HasLimit:    HasClause(normalized, "LIMIT"),
		HasSubquery: hasSubquery(normalized),
		HasUnion:    HasClause(normalized, "UNION"),
		HasCTE:      strings.HasPrefix(strings.ToUpper(normalized), "WITH "),
	}

	// 判断查询类型
	if detail.HasCTE {
		detail.SelectType = "CTE"
	} else if detail.HasUnion {
		detail.SelectType = "UNION"
	} else if detail.HasSubquery {
		detail.SelectType = "SUBQUERY"
	}

	// 提取表引用
	detail.Tables = extractTableRefs(normalized)

	// 提取查询字段
	detail.Columns = extractSelectColumns(normalized)

	result.Analyze = detail
	evaluateDQLRisk(result, detail, normalized)

	return result
}

// hasJoin 检查是否有JOIN
func hasJoin(sql string) bool {
	upper := strings.ToUpper(sql)
	return strings.Contains(upper, " JOIN ") ||
		strings.Contains(upper, " LEFT ") ||
		strings.Contains(upper, " RIGHT ") ||
		strings.Contains(upper, " INNER ") ||
		strings.Contains(upper, " CROSS ")
}

// hasSubquery 检查是否有子查询
func hasSubquery(sql string) bool {
	upper := strings.ToUpper(sql)
	// 统计SELECT数量，大于1说明有子查询
	return strings.Count(upper, "SELECT") > 1
}

// extractTableRefs 提取表引用
func extractTableRefs(sql string) []models.TableRef {
	var tables []models.TableRef
	seen := make(map[string]bool)

	// 匹配 FROM table [AS] alias
	fromRe := regexp.MustCompile(`(?i)FROM\s+` + "`?" + `(\w+)` + "`?" + `(?:\s+(?:AS\s+)?` + "`?" + `(\w+)` + "`?" + `)?`)
	if m := fromRe.FindStringSubmatch(sql); len(m) > 1 {
		name := m[1]
		alias := ""
		if len(m) > 2 && m[2] != "" && !IsKeyword(m[2]) {
			alias = m[2]
		}
		key := name + ":" + alias
		if !seen[key] && !IsKeyword(name) {
			seen[key] = true
			tables = append(tables, models.TableRef{Name: name, Alias: alias, Type: "TABLE"})
		}
	}

	// 匹配 JOIN table [AS] alias
	joinRe := regexp.MustCompile(`(?i)JOIN\s+` + "`?" + `(\w+)` + "`?" + `(?:\s+(?:AS\s+)?` + "`?" + `(\w+)` + "`?" + `)?`)
	matches := joinRe.FindAllStringSubmatch(sql, -1)
	for _, m := range matches {
		if len(m) > 1 {
			name := m[1]
			alias := ""
			if len(m) > 2 && m[2] != "" && !IsKeyword(m[2]) {
				alias = m[2]
			}
			key := name + ":" + alias
			if !seen[key] && !IsKeyword(name) {
				seen[key] = true
				tables = append(tables, models.TableRef{Name: name, Alias: alias, Type: "TABLE"})
			}
		}
	}

	return tables
}

// extractSelectColumns 提取SELECT字段
func extractSelectColumns(sql string) []string {
	var columns []string
	upper := strings.ToUpper(sql)

	// 找到SELECT位置
	selectIdx := strings.Index(upper, "SELECT")
	if selectIdx < 0 {
		return columns
	}

	// 找到FROM位置（主查询的FROM）
	fromIdx := findMainFromIndex(sql)
	if fromIdx < 0 {
		fromIdx = len(sql)
	}

	start := selectIdx + 6
	if start >= fromIdx {
		return columns
	}

	fieldPart := strings.TrimSpace(sql[start:fromIdx])

	// 移除DISTINCT
	if strings.HasPrefix(strings.ToUpper(fieldPart), "DISTINCT ") {
		fieldPart = strings.TrimSpace(fieldPart[9:])
	}

	if fieldPart == "*" {
		return []string{"*"}
	}

	// 拆分字段
	fields := splitDefinitions(fieldPart)
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			col := extractColumnAlias(f)
			if col != "" {
				columns = append(columns, col)
			}
		}
	}

	return columns
}

// findMainFromIndex 找到主查询的FROM位置
func findMainFromIndex(sql string) int {
	upper := strings.ToUpper(sql)
	depth := 0

	for i := 0; i < len(sql)-4; i++ {
		if sql[i] == '(' {
			depth++
		} else if sql[i] == ')' {
			depth--
		} else if depth == 0 && upper[i:i+4] == "FROM" {
			if (i == 0 || !isAlphaNum(sql[i-1])) && (i+4 >= len(sql) || !isAlphaNum(sql[i+4])) {
				return i
			}
		}
	}

	return -1
}

// extractColumnAlias 提取字段名或别名
func extractColumnAlias(field string) string {
	field = strings.TrimSpace(field)
	upper := strings.ToUpper(field)

	// 检查 AS 别名
	if idx := strings.LastIndex(upper, " AS "); idx > 0 {
		return strings.Trim(strings.TrimSpace(field[idx+4:]), "`'\"")
	}

	// 检查空格别名
	if !strings.Contains(field, "(") {
		parts := strings.Fields(field)
		if len(parts) > 1 {
			last := parts[len(parts)-1]
			if !IsKeyword(last) {
				return strings.Trim(last, "`")
			}
		}
	}

	// table.column
	if strings.Contains(field, ".") && !strings.Contains(field, "(") {
		parts := strings.Split(field, ".")
		return strings.Trim(parts[len(parts)-1], "`")
	}

	// 函数
	if strings.Contains(field, "(") {
		if idx := strings.Index(field, "("); idx > 0 {
			return strings.ToUpper(strings.TrimSpace(field[:idx]))
		}
	}

	return strings.Trim(field, "`")
}

// isAlphaNum 判断是否为字母数字下划线
func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// evaluateDQLRisk 评估DQL风险
func evaluateDQLRisk(result *models.AnalyzeResponse, detail *models.DQLDetail, normalized string) {
	result.BasicInfo.RiskLevel = "safe"
	result.BasicInfo.RiskReason = "查询操作，不修改数据"

	// 检查是否有聚合函数
	hasAggFunc := hasAggregateFunction(detail.Columns)

	// 无WHERE条件
	if !detail.HasWhere {
		// 如果没有聚合函数，无WHERE是高风险
		if !hasAggFunc {
			result.BasicInfo.RiskLevel = "danger"
			result.BasicInfo.RiskReason = "无WHERE条件的全表查询，可能返回大量数据！"
			return
		}
		// 有聚合函数，无WHERE是中风险
		result.BasicInfo.RiskLevel = "warning"
		result.BasicInfo.RiskReason = "聚合查询无WHERE条件，统计全表数据"
		return
	}

	// 有WHERE条件，检查条件类型
	whereClause := extractWhereClause(normalized)
	if whereClause != "" {
		upper := strings.ToUpper(whereClause)
		// 检查是否包含范围查询符号或BETWEEN
		if strings.Contains(whereClause, ">") || strings.Contains(whereClause, "<") || strings.Contains(upper, "BETWEEN") {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "WHERE条件包含范围查询（>、< 或 BETWEEN），可能返回较多数据"
			return
		}
		// 精确查询（=）
		if strings.Contains(whereClause, "=") {
			result.BasicInfo.RiskLevel = "safe"
			result.BasicInfo.RiskReason = "WHERE条件使用精确匹配（=），查询效率较高"
			return
		}
	}

	// 其他情况默认safe
	result.BasicInfo.RiskLevel = "safe"
	result.BasicInfo.RiskReason = "查询操作，有WHERE条件"
}

// hasAggregateFunction 检查是否包含聚合函数
func hasAggregateFunction(columns []string) bool {
	aggFuncs := []string{"COUNT", "AVG", "SUM", "MAX", "MIN"}
	for _, col := range columns {
		upper := strings.ToUpper(col)
		for _, fn := range aggFuncs {
			if strings.HasPrefix(upper, fn+"(") || upper == fn {
				return true
			}
		}
	}
	return false
}
