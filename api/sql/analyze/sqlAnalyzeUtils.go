package analyze

import (
	"regexp"
	"strings"
)

// NormalizeSQL 标准化SQL语句
func NormalizeSQL(sql string) string {
	// 1. 移除多行注释
	sql = removeBlockComments(sql)

	// 2. 移除单行注释
	sql = removeLineComments(sql)

	// 3. 处理换行和制表符
	sql = strings.ReplaceAll(sql, "\r\n", " ")
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")

	// 4. 合并多个空格
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")

	// 5. 去除首尾空格和分号
	sql = strings.TrimSpace(sql)
	sql = strings.TrimSuffix(sql, ";")
	sql = strings.TrimSpace(sql)

	return sql
}

// removeBlockComments 移除多行注释
func removeBlockComments(sql string) string {
	re := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return re.ReplaceAllString(sql, " ")
}

// removeLineComments 移除单行注释
func removeLineComments(sql string) string {
	lines := strings.Split(sql, "\n")
	var result []string

	for _, line := range lines {
		line = removeInlineComment(line, "--")
		line = removeInlineComment(line, "#")
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// removeInlineComment 移除行内注释（考虑字符串）
func removeInlineComment(line, mark string) string {
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(line); i++ {
		c := line[i]

		if c == '\'' || c == '"' || c == '`' {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar && (i == 0 || line[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if !inString && i+len(mark) <= len(line) && line[i:i+len(mark)] == mark {
			return line[:i]
		}
	}

	return line
}

// SplitMultiSQL 拆分多条SQL语句
func SplitMultiSQL(sql string) []string {
	var sqls []string
	var current strings.Builder

	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		c := sql[i]

		if c == '\'' || c == '"' || c == '`' {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar && (i == 0 || sql[i-1] != '\\') {
				inString = false
			}
			current.WriteByte(c)
			continue
		}

		if c == ';' && !inString {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sqls = append(sqls, s)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	s := strings.TrimSpace(current.String())
	if s != "" {
		sqls = append(sqls, s)
	}

	return sqls
}

// GetSQLType 获取SQL类型
func GetSQLType(sql string) string {
	sql = NormalizeSQL(sql)
	upper := strings.ToUpper(sql)

	prefixes := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME",
		"WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN", "USE",
	}

	for _, p := range prefixes {
		if strings.HasPrefix(upper, p+" ") || strings.HasPrefix(upper, p+"(") || upper == p {
			if p == "DESC" {
				return "DESCRIBE"
			}
			return p
		}
	}

	return "UNKNOWN"
}

// GetSQLCategory 获取SQL分类
func GetSQLCategory(sqlType string) string {
	switch sqlType {
	case "SELECT", "WITH", "SHOW", "DESCRIBE", "EXPLAIN":
		return "DQL"
	case "INSERT", "UPDATE", "DELETE":
		return "DML"
	case "CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME":
		return "DDL"
	case "USE":
		return "OTHER"
	default:
		return "UNKNOWN"
	}
}

// HasClause 检查SQL是否包含指定子句
func HasClause(sql, clause string) bool {
	upper := strings.ToUpper(sql)
	clause = strings.ToUpper(clause)
	return strings.Contains(upper, " "+clause+" ") ||
		strings.Contains(upper, " "+clause+"(") ||
		strings.HasSuffix(upper, " "+clause)
}

// IsKeyword 判断是否为SQL关键字
func IsKeyword(word string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
		"JOIN": true, "LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true,
		"ON": true, "SET": true, "VALUES": true, "INTO": true, "AS": true,
		"ORDER": true, "BY": true, "GROUP": true, "HAVING": true, "LIMIT": true,
		"OFFSET": true, "UNION": true, "ALL": true, "DISTINCT": true,
		"IF": true, "EXISTS": true, "NOT": true, "NULL": true,
	}
	return keywords[strings.ToUpper(word)]
}
