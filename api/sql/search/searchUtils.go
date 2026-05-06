package search

import (
	"fmt"
	"strings"
)

// ShouldFilterTable 判断表是否应该被过滤（完全匹配，不是包含匹配）
func ShouldFilterTable(tableName string) bool {
	lowerTableName := strings.ToLower(tableName)

	for _, filterTable := range FilteredTables {
		if lowerTableName == strings.ToLower(filterTable) {
			return true
		}
	}
	return false
}

// AddOffsetLimitToQuery 为查询添加OFFSET和LIMIT限制
// 用于分批加载数据
func AddOffsetLimitToQuery(query string, offset int, limit int) string {
	// 移除现有的LIMIT子句(需要确保是SQL关键字,不是表名的一部分)
	query = removeLimitClause(query)

	// 移除末尾可能存在的分号
	query = strings.TrimSuffix(strings.TrimSpace(query), ";")

	return fmt.Sprintf("%s LIMIT %d OFFSET %d;", query, limit, offset)
}

// removeLimitClause 智能移除 LIMIT 子句
// 确保只移除 SQL 关键字 LIMIT,而不是表名中包含的 "limit" 字符串
func removeLimitClause(query string) string {
	lowerQuery := strings.ToLower(query)

	// 查找所有 "limit" 出现的位置
	searchStr := "limit"
	startPos := 0

	for {
		idx := strings.Index(lowerQuery[startPos:], searchStr)
		if idx == -1 {
			break
		}

		actualIdx := startPos + idx

		// 检查 "limit" 前后是否是单词边界
		// 前面应该是空格、换行、制表符或开头
		beforeOK := actualIdx == 0 || isWhitespaceChar(lowerQuery[actualIdx-1])

		// 后面应该是空格、换行、制表符或结尾
		afterIdx := actualIdx + len(searchStr)
		afterOK := afterIdx >= len(lowerQuery) || isWhitespaceChar(lowerQuery[afterIdx])

		// 如果前后都是边界,说明这是一个独立的 LIMIT 关键字
		if beforeOK && afterOK {
			// 截断到 LIMIT 之前
			query = strings.TrimSpace(query[:actualIdx])
			break
		}

		// 继续查找下一个
		startPos = actualIdx + len(searchStr)
	}

	return query
}

// isWhitespaceChar 判断字符是否是空白字符
func isWhitespaceChar(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ';'
}
