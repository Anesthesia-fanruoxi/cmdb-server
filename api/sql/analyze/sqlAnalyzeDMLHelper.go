package analyze

import (
	"regexp"
	"strings"
)

// extractInsertValues 提取INSERT的VALUES值
func extractInsertValues(sql string) [][]string {
	var result [][]string

	upper := strings.ToUpper(sql)
	idx := strings.Index(upper, " VALUES")
	if idx < 0 {
		idx = strings.Index(upper, " VALUE")
	}
	if idx < 0 {
		return result
	}

	valuesPart := sql[idx:]
	start := strings.Index(valuesPart, "(")
	if start < 0 {
		return result
	}
	valuesPart = valuesPart[start:]

	rows := splitValueRows(valuesPart)
	for _, row := range rows {
		values := parseValueRow(row)
		if len(values) > 0 {
			result = append(result, values)
		}
	}

	return result
}

// splitValueRows 拆分VALUES中的多行
func splitValueRows(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if (c == '\'' || c == '"') && (i == 0 || s[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
			current.WriteByte(c)
			continue
		}

		if inString {
			current.WriteByte(c)
			continue
		}

		if c == '(' {
			if depth == 0 {
				current.Reset()
			}
			depth++
			if depth > 1 {
				current.WriteByte(c)
			}
		} else if c == ')' {
			depth--
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else if depth > 0 {
				current.WriteByte(c)
			}
		} else if depth > 0 {
			current.WriteByte(c)
		}
	}

	return result
}

// parseValueRow 解析单行VALUES
func parseValueRow(row string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(row); i++ {
		c := row[i]

		if (c == '\'' || c == '"') && (i == 0 || row[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
			current.WriteByte(c)
			continue
		}

		if inString {
			current.WriteByte(c)
			continue
		}

		if c == '(' {
			depth++
			current.WriteByte(c)
		} else if c == ')' {
			depth--
			current.WriteByte(c)
		} else if c == ',' && depth == 0 {
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// splitSetClauses 拆分SET子句（考虑函数调用中的逗号）
func splitSetClauses(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if (c == '\'' || c == '"') && (i == 0 || s[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
			current.WriteByte(c)
			continue
		}

		if inString {
			current.WriteByte(c)
			continue
		}

		if c == '(' {
			depth++
			current.WriteByte(c)
		} else if c == ')' {
			depth--
			current.WriteByte(c)
		} else if c == ',' && depth == 0 {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// isExpression 判断值是否是表达式
func isExpression(value string) bool {
	upper := strings.ToUpper(value)

	// 简单值：NULL
	if value == "NULL" || value == "null" {
		return false
	}
	// 简单值：字符串
	if strings.HasPrefix(value, "'") || strings.HasPrefix(value, "\"") {
		return false
	}
	// 简单值：数字
	numRe := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	if numRe.MatchString(value) {
		return false
	}

	// 包含函数调用、运算符等
	if strings.Contains(value, "(") ||
		strings.Contains(value, "+") ||
		strings.Contains(value, "-") ||
		strings.Contains(value, "*") ||
		strings.Contains(value, "/") ||
		strings.Contains(upper, "CONCAT") ||
		strings.Contains(upper, "NOW()") ||
		strings.Contains(upper, "CURRENT_") {
		return true
	}

	return false
}
