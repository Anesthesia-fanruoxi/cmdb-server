package rule

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CheckRuleRequest struct {
	SQL string `json:"sql"` // SQL语句（可能包含多条语句）
}

// AnalysisRequestBody 解析请求体，返回SQL语句列表
func AnalysisRequestBody(r *http.Request) ([]string, error) {
	// 1. 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求数据失败: %v", err)
	}
	defer r.Body.Close()

	// 2. 解析JSON
	var req CheckRuleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("解析请求数据失败: %v", err)
	}

	// 3. 验证SQL不为空
	if req.SQL == "" {
		return nil, fmt.Errorf("SQL语句不能为空")
	}

	// 5. 按行处理SQL
	var sqlStatements []string
	var currentSQL strings.Builder
	lines := strings.Split(req.SQL, "\n")

	for _, line := range lines {
		// 清理行首尾空白
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 跳过以--或//开头的行（纯注释行）
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "//") {
			continue
		}

		// 移除行尾注释（-- 或 //）
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// 移除行尾块注释（/* 开始的注释）
		if idx := strings.Index(line, "/*"); idx >= 0 {
			// 检查是否在同一行结束
			if endIdx := strings.LastIndex(line, "*/"); endIdx > idx+2 {
				// 同一行有开始和结束，需要移除中间部分
				beforeComment := strings.TrimSpace(line[:idx])
				afterComment := strings.TrimSpace(line[endIdx+2:])
				line = beforeComment
				if beforeComment != "" && afterComment != "" {
					line += " " // 保持间隔
				}
				line += afterComment
			} else {
				// 注释未在本行结束，只保留注释前部分
				line = strings.TrimSpace(line[:idx])
			}
		}

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否包含分号（SQL语句结束符）
		if strings.Contains(line, ";") {
			parts := strings.Split(line, ";")
			for i, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// 将部分添加到当前SQL
				if currentSQL.Len() > 0 {
					currentSQL.WriteString(" ")
				}
				currentSQL.WriteString(part)

				// 如果不是最后一部分或行以分号结束，则结束当前SQL
				if i < len(parts)-1 || strings.HasSuffix(line, ";") {
					sql := strings.TrimSpace(currentSQL.String())
					if sql != "" && isValidSQL(sql) {
						sqlStatements = append(sqlStatements, sql)
					}
					currentSQL.Reset()
				}
			}
		} else {
			// 如果当前行没有分号，添加到当前SQL
			if currentSQL.Len() > 0 {
				currentSQL.WriteString(" ")
			}
			currentSQL.WriteString(line)
		}
	}

	// 处理最后一条未结束的SQL（如果有）
	if currentSQL.Len() > 0 {
		sql := strings.TrimSpace(currentSQL.String())
		if sql != "" && isValidSQL(sql) {
			sqlStatements = append(sqlStatements, sql)
		}
	}

	return sqlStatements, nil
}

// isValidSQL 检查是否为有效的SQL语句
func isValidSQL(sql string) bool {
	// 清理字符串并分割为单词
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return false
	}

	words := strings.Fields(sql)
	if len(words) == 0 {
		return false
	}

	// 检查首个关键字
	firstKeyword := strings.ToUpper(words[0])
	validKeywords := map[string]bool{
		"SELECT":   true,
		"INSERT":   true,
		"UPDATE":   true,
		"DELETE":   true,
		"CREATE":   true,
		"ALTER":    true,
		"DROP":     true,
		"TRUNCATE": true,
		"WITH":     true,
		"SET":      true,
	}

	return validKeywords[firstKeyword]
}
