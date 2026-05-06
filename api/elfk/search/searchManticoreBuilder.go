package search

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ManticoreQueryBuilder Manticore查询构建器
type ManticoreQueryBuilder struct {
	Index      string // 索引名称（即Manticore中的表名）
	StartTime  string // 开始时间
	EndTime    string // 结束时间
	TimeField  string // 时间字段
	TimeFormat string // 时间格式，如 "iso8601"、"epoch_millis"
	Size       int    // 返回结果数量
	Query      string // 查询条件（SQL WHERE子句）
	OrderBy    string // 排序字段和方向
}

// ParseKeyword 解析关键词并构建查询条件
func (qb *ManticoreQueryBuilder) ParseKeyword(keyword string) error {
	if keyword == "" || keyword == "*" {
		// 空关键词或通配符，不添加额外过滤条件，只使用时间范围
		qb.Query = qb.buildTimeRangeQuery()
		return nil
	}

	// 处理关键词查询
	parsedQuery, err := qb.parseQuery(keyword)
	if err != nil {
		return err
	}

	// 添加时间范围条件
	timeRangeQuery := qb.buildTimeRangeQuery()
	if timeRangeQuery != "" {
		if parsedQuery != "" {
			// 组合查询条件
			qb.Query = fmt.Sprintf("(%s) AND %s", parsedQuery, timeRangeQuery)
		} else {
			qb.Query = timeRangeQuery
		}
	} else if parsedQuery != "" {
		qb.Query = parsedQuery
	} else {
		// 确保即使没有有效的查询条件，也返回基本的条件
		qb.Query = "1=1" // Manticore中的永真条件
	}

	return nil
}

// parseQuery 解析查询表达式，返回WHERE子句
func (qb *ManticoreQueryBuilder) parseQuery(input string) (string, error) {
	tokens := tokenizeManticore(input)
	return buildManticoreQuery(tokens)
}

// buildTimeRangeQuery 构建时间范围查询
func (qb *ManticoreQueryBuilder) buildTimeRangeQuery() string {
	startTime, endTime := qb.formatTimeValues()
	if startTime != "" && endTime != "" {
		return fmt.Sprintf("%s BETWEEN %s AND %s", qb.TimeField, startTime, endTime)
	} else if startTime != "" {
		return fmt.Sprintf("%s >= %s", qb.TimeField, startTime)
	} else if endTime != "" {
		return fmt.Sprintf("%s <= %s", qb.TimeField, endTime)
	}
	return ""
}

// BuildTimeRangeQuery 构建时间范围查询（公共方法）
func (qb *ManticoreQueryBuilder) BuildTimeRangeQuery() string {
	return qb.buildTimeRangeQuery()
}

// formatTimeValues 格式化时间值为Manticore可接受的格式
func (qb *ManticoreQueryBuilder) formatTimeValues() (string, string) {
	var startTimeStr, endTimeStr string

	// 解析开始时间
	if qb.StartTime != "" {
		startTime, err := parseTime(qb.StartTime, qb.TimeFormat)
		if err == nil {
			// 根据时间格式不同，选择不同的时间表示方式
			if qb.TimeFormat == "epoch_millis" || qb.TimeFormat == "epoch_second" {
				startTimeStr = fmt.Sprintf("%d", startTime.Unix())
			} else {
				// 对于ISO格式，使用引号包裹时间字符串
				startTimeStr = fmt.Sprintf("'%s'", startTime.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// 解析结束时间
	if qb.EndTime != "" {
		endTime, err := parseTime(qb.EndTime, qb.TimeFormat)
		if err == nil {
			if qb.TimeFormat == "epoch_millis" || qb.TimeFormat == "epoch_second" {
				endTimeStr = fmt.Sprintf("%d", endTime.Unix())
			} else {
				endTimeStr = fmt.Sprintf("'%s'", endTime.Format("2006-01-02 15:04:05"))
			}
		}
	}

	return startTimeStr, endTimeStr
}

// parseTime 根据指定格式解析时间字符串
func parseTime(timeStr, format string) (time.Time, error) {
	switch format {
	case "epoch_millis":
		// 毫秒级时间戳
		var millis int64
		_, err := fmt.Sscanf(timeStr, "%d", &millis)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(millis/1000, (millis%1000)*1000000), nil
	case "epoch_second":
		// 秒级时间戳
		var seconds int64
		_, err := fmt.Sscanf(timeStr, "%d", &seconds)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(seconds, 0), nil
	case "iso8601":
		fallthrough
	default:
		// 尝试解析ISO8601格式
		return time.Parse(time.RFC3339, timeStr)
	}
}

// BuildSQL 构建完整的Manticore SQL查询
func (qb *ManticoreQueryBuilder) BuildSQL() string {
	var queryStr strings.Builder

	// 构建基本SELECT语句
	queryStr.WriteString(fmt.Sprintf("SELECT * FROM %s", qb.Index))

	// 添加WHERE子句
	if qb.Query != "" {
		queryStr.WriteString(fmt.Sprintf(" WHERE %s", qb.Query))
	}

	// 添加ORDER BY子句
	if qb.OrderBy != "" {
		queryStr.WriteString(fmt.Sprintf(" ORDER BY %s", qb.OrderBy))
	}

	// 添加LIMIT子句
	queryStr.WriteString(fmt.Sprintf(" LIMIT %d", qb.Size))

	return queryStr.String()
}

// ManticoreToken 表示Manticore查询语句中的一个标记
type ManticoreToken struct {
	Type  string // field, operator, value, logic
	Value string
}

// tokenizeManticore 将查询字符串分割成标记
func tokenizeManticore(input string) []ManticoreToken {
	var tokens []ManticoreToken

	// 处理特殊情况：星号表示匹配所有
	if input == "*" {
		return []ManticoreToken{}
	}

	// 预处理：处理引号包裹的短语
	inQuote := false
	var currentPhrase strings.Builder
	var quoteChar rune
	var processedParts []string

	for _, char := range input {
		if (char == '"' || char == '\'') && (currentPhrase.Len() == 0 || !inQuote) {
			inQuote = !inQuote
			quoteChar = char
			if currentPhrase.Len() > 0 {
				currentPhrase.WriteRune(char)
			}
		} else if inQuote && char == quoteChar {
			inQuote = false
			currentPhrase.WriteRune(char)
			processedParts = append(processedParts, currentPhrase.String())
			currentPhrase.Reset()
		} else if inQuote {
			currentPhrase.WriteRune(char)
		} else if char == ' ' || char == '\t' {
			if currentPhrase.Len() > 0 {
				processedParts = append(processedParts, currentPhrase.String())
				currentPhrase.Reset()
			}
		} else {
			currentPhrase.WriteRune(char)
		}
	}

	if currentPhrase.Len() > 0 {
		processedParts = append(processedParts, currentPhrase.String())
	}

	// 处理分割后的部分
	for i := 0; i < len(processedParts); i++ {
		part := processedParts[i]

		// 处理逻辑运算符
		if strings.EqualFold(part, "and") {
			tokens = append(tokens, ManticoreToken{Type: "logic", Value: "AND"})
			continue
		}
		if strings.EqualFold(part, "or") {
			tokens = append(tokens, ManticoreToken{Type: "logic", Value: "OR"})
			continue
		}
		if strings.EqualFold(part, "not") {
			tokens = append(tokens, ManticoreToken{Type: "logic", Value: "NOT"})
			continue
		}

		// 先检查是否为字段:值格式 (支持多种分隔符: :, =, !=)
		fieldValuePattern := regexp.MustCompile(`^([a-zA-Z0-9_@]+)(:|=|!=)(.*)$`)
		if matches := fieldValuePattern.FindStringSubmatch(part); len(matches) == 4 {
			field := matches[1]
			operator := matches[2]
			value := matches[3]

			tokens = append(tokens, ManticoreToken{Type: "field", Value: field})
			tokens = append(tokens, ManticoreToken{Type: "operator", Value: operator})
			tokens = append(tokens, ManticoreToken{Type: "value", Value: value})
			continue
		}

		// 普通值
		tokens = append(tokens, ManticoreToken{Type: "value", Value: part})
	}

	return tokens
}

// buildManticoreQuery 根据标记构建Manticore查询语句
func buildManticoreQuery(tokens []ManticoreToken) (string, error) {
	if len(tokens) == 0 {
		return "", nil
	}

	// 处理单个值
	if len(tokens) == 1 && tokens[0].Type == "value" {
		value := tokens[0].Value
		// 移除引号（如果存在）
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		// 为模糊匹配在关键字前后添加*
		value = fmt.Sprintf("*%s*", value)
		// 构建全文搜索查询
		return fmt.Sprintf("MATCH('%s')", escapeManticoreValue(value)), nil
	}

	// 处理字段查询 (字段:值 模式)
	if len(tokens) == 3 && tokens[0].Type == "field" && tokens[1].Type == "operator" && tokens[2].Type == "value" {
		field := tokens[0].Value
		// 处理@开头的字段名
		field = strings.TrimPrefix(field, "@")

		operator := tokens[1].Value
		value := tokens[2].Value

		// 移除引号（如果存在）
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		switch operator {
		case ":", "=":
			// 特殊处理：字段为*表示任意字段匹配
			if field == "*" {
				// 为模糊匹配在关键字前后添加*
				value = fmt.Sprintf("*%s*", value)
				return fmt.Sprintf("MATCH('%s')", escapeManticoreValue(value)), nil
			}
			// 检查值是否包含通配符
			if value == "*" {
				return fmt.Sprintf("%s IS NOT NULL", field), nil
			}
			// 对于普通字段匹配，使用两种方式:
			// 1. 如果字段是message或content等可能包含全文的字段，使用MATCH在该字段内搜索
			if field == "message" || field == "content" || field == "log" {
				// 为模糊匹配在关键字前后添加*
				value = fmt.Sprintf("*%s*", value)
				return fmt.Sprintf("MATCH('@%s %s')", field, escapeManticoreValue(value)), nil
			}
			// 2. 否则使用精确匹配
			return fmt.Sprintf("%s = '%s'", field, escapeManticoreValue(value)), nil
		case "!=":
			if value == "*" {
				return fmt.Sprintf("%s IS NULL", field), nil
			}
			return fmt.Sprintf("%s != '%s'", field, escapeManticoreValue(value)), nil
		}
	}

	// 处理复杂查询 - 需要专门处理逻辑关系
	var fieldConditions []string       // 字段条件（field:value形式）
	var matchTerms []string            // 全文匹配词（AND关系）
	var orTerms []string               // OR关系的词
	var notTerms []string              // NOT关系的词
	var notFieldConditions []string    // 否定的字段条件
	var currentOperator string = "AND" // 默认为AND

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case "logic":
			currentOperator = token.Value
		case "value":
			value := token.Value
			// 移除引号（如果存在）
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}

			// 为模糊匹配在关键字前后添加*
			value = fmt.Sprintf("*%s*", value)

			// 根据当前操作符决定如何处理这个值
			if currentOperator == "NOT" {
				notTerms = append(notTerms, value)
				// 重置为默认操作符
				currentOperator = "AND"
			} else if currentOperator == "OR" {
				orTerms = append(orTerms, value)
			} else { // AND 或默认
				matchTerms = append(matchTerms, value)
			}
		case "field":
			// 确保有足够的标记
			if i+2 < len(tokens) && tokens[i+1].Type == "operator" && tokens[i+2].Type == "value" {
				field := strings.TrimPrefix(token.Value, "@") // 处理@开头的字段名
				operator := tokens[i+1].Value
				value := tokens[i+2].Value

				// 移除引号（如果存在）
				if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
					(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
					value = value[1 : len(value)-1]
				}

				// 处理字段查询
				var condition string
				switch operator {
				case ":", "=":
					if value == "*" {
						condition = fmt.Sprintf("%s IS NOT NULL", field)
					} else if field == "message" || field == "content" || field == "log" {
						// 对于全文字段，使用MATCH
						value = fmt.Sprintf("*%s*", value) // 添加模糊匹配
						condition = fmt.Sprintf("MATCH('@%s %s')", field, escapeManticoreValue(value))
					} else {
						// 对于普通字段，使用精确匹配
						condition = fmt.Sprintf("%s = '%s'", field, escapeManticoreValue(value))
					}
				case "!=":
					if value == "*" {
						condition = fmt.Sprintf("%s IS NULL", field)
					} else {
						condition = fmt.Sprintf("%s != '%s'", field, escapeManticoreValue(value))
					}
				}

				// 根据当前操作符决定怎么处理这个条件
				if currentOperator == "NOT" {
					notFieldConditions = append(notFieldConditions, condition)
					currentOperator = "AND" // 重置操作符
				} else {
					fieldConditions = append(fieldConditions, condition)
				}

				i += 2 // 跳过已处理的操作符和值
			}
		case "operator":
			if token.Value == "NOT" {
				currentOperator = "NOT"
			}
		}
	}

	// 构建最终的查询
	var queryParts []string

	// 处理全文搜索部分
	if len(matchTerms) > 0 || len(orTerms) > 0 || len(notTerms) > 0 {
		var matchQueryParts []string

		// 处理AND关系的词
		if len(matchTerms) > 0 {
			matchQueryParts = append(matchQueryParts, strings.Join(matchTerms, " "))
		}

		// 处理OR关系的词
		if len(orTerms) > 0 {
			orExpr := strings.Join(orTerms, "|")
			if len(matchQueryParts) > 0 {
				matchQueryParts[0] = fmt.Sprintf("%s %s", matchQueryParts[0], orExpr)
			} else {
				matchQueryParts = append(matchQueryParts, orExpr)
			}
		}

		// 构建基本match表达式
		matchExpr := strings.Join(matchQueryParts, " ")

		// 处理NOT关系的词
		for _, term := range notTerms {
			if matchExpr != "" {
				matchExpr = fmt.Sprintf("%s -%s", matchExpr, term)
			} else {
				matchExpr = fmt.Sprintf("-%s", term)
			}
		}

		if matchExpr != "" {
			queryParts = append(queryParts, fmt.Sprintf("MATCH('%s')", escapeManticoreValue(matchExpr)))
		}
	}

	// 添加字段条件
	for _, condition := range fieldConditions {
		queryParts = append(queryParts, condition)
	}

	// 处理否定的字段条件
	for _, condition := range notFieldConditions {
		queryParts = append(queryParts, fmt.Sprintf("NOT(%s)", condition))
	}

	// 组合所有条件
	if len(queryParts) > 0 {
		return strings.Join(queryParts, " AND "), nil
	}

	return "", nil
}

// escapeManticoreValue 转义Manticore值中的特殊字符
func escapeManticoreValue(value string) string {
	// 转义单引号（在SQL中单引号需要使用两个单引号表示）
	return strings.ReplaceAll(value, "'", "''")
}
