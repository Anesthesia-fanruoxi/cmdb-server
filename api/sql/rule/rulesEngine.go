package rule

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	models "cmdb/models/sql"
	"cmdb/pkg/middleware"
)

// RuleConfig 规则配置
type RuleConfig struct {
	CheckField    string      `json:"check_field"`              // 要检查的字段路径
	CheckType     string      `json:"check_type"`               // 检查类型
	CheckValue    interface{} `json:"check_value,omitempty"`    // 检查值
	ErrorTemplate string      `json:"error_template,omitempty"` // 错误信息模板
}

// CheckRuleEngine 通用规则检查引擎
func CheckRuleEngine(analyzeResult *models.AnalyzeResponse, rule models.SqlRule) *models.RuleCheckResult {
	result := &models.RuleCheckResult{
		RuleID:          rule.ID,
		RuleName:        rule.RuleName,
		RuleType:        rule.RuleType,
		RuleCategory:    rule.RuleCategory,
		RuleDescription: rule.RuleDescription,
		Passed:          true,
	}

	// 解析规则配置
	var config RuleConfig
	if rule.RuleContent != "" {
		if err := json.Unmarshal([]byte(rule.RuleContent), &config); err != nil {
			middleware.Logger.Warnf("规则配置解析失败: %s, 错误: %v", rule.RuleName, err)
			result.Passed = false
			result.ErrorMessage = "规则配置解析失败"
			return result
		}
	}

	// 获取字段值
	fieldValue, err := getFieldValue(analyzeResult, config.CheckField)
	if err != nil {
		middleware.Logger.Warnf("获取字段值失败: %s, 字段: %s, 错误: %v", rule.RuleName, config.CheckField, err)
		result.Passed = true // 字段不存在时默认通过
		return result
	}

	// 执行检查
	passed, errorMsg, details := executeCheck(fieldValue, config)
	result.Passed = passed
	if !passed {
		if config.ErrorTemplate != "" {
			result.ErrorMessage = config.ErrorTemplate
		} else if errorMsg != "" {
			result.ErrorMessage = errorMsg
		} else {
			result.ErrorMessage = rule.ErrorMessage
		}
		result.Details = details

		// 收集具体的违规明细
		result.ViolationDetails = CollectViolationDetails(analyzeResult, rule)
	}

	return result
}

// getFieldValue 获取字段值（支持嵌套路径）
func getFieldValue(analyzeResult *models.AnalyzeResponse, fieldPath string) (interface{}, error) {
	if fieldPath == "" {
		return nil, fmt.Errorf("字段路径为空")
	}

	parts := strings.Split(fieldPath, ".")

	// 第一层：从 Analyze 中获取具体类型
	var current interface{}

	switch parts[0] {
	case "table_info", "field_info", "index_detail", "field_detail", "alter_info", "alter_detail", "simple_info":
		// DDL 类型
		ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
		if !ok {
			return nil, fmt.Errorf("不是DDL类型")
		}
		current = ddlDetail

	case "operation", "table", "columns", "has_where", "where_clause", "has_limit", "row_count", "data_source", "values", "set_clauses":
		// DML 类型
		dmlDetail, ok := analyzeResult.Analyze.(*models.DMLDetail)
		if !ok {
			return nil, fmt.Errorf("不是DML类型")
		}
		current = dmlDetail

	case "select_type", "tables", "has_join", "has_group", "has_order", "has_subquery", "has_union", "has_cte":
		// DQL 类型
		dqlDetail, ok := analyzeResult.Analyze.(*models.DQLDetail)
		if !ok {
			return nil, fmt.Errorf("不是DQL类型")
		}
		current = dqlDetail

	default:
		return nil, fmt.Errorf("未知的字段路径: %s", parts[0])
	}

	// 遍历路径获取值
	for _, part := range parts {
		val := reflect.ValueOf(current)

		// 处理指针
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil, fmt.Errorf("字段为nil: %s", part)
			}
			val = val.Elem()
		}

		// 处理结构体
		if val.Kind() == reflect.Struct {
			// 转换字段名：snake_case -> PascalCase
			fieldName := snakeToPascal(part)
			field := val.FieldByName(fieldName)
			if !field.IsValid() {
				return nil, fmt.Errorf("字段不存在: %s", part)
			}
			current = field.Interface()
		} else {
			// 已经是最终值
			return current, nil
		}
	}

	return current, nil
}

// snakeToPascal 将 snake_case 转换为 PascalCase
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// executeCheck 执行检查
func executeCheck(fieldValue interface{}, config RuleConfig) (bool, string, []string) {
	switch config.CheckType {
	case "empty":
		return checkEmpty(fieldValue)
	case "not_empty":
		return checkNotEmpty(fieldValue)
	case "equals":
		return checkEquals(fieldValue, config.CheckValue)
	case "not_equals":
		return checkNotEquals(fieldValue, config.CheckValue)
	case "contains":
		return checkContains(fieldValue, config.CheckValue)
	case "not_contains":
		return checkNotContains(fieldValue, config.CheckValue)
	case "in":
		return checkIn(fieldValue, config.CheckValue)
	case "not_in":
		return checkNotIn(fieldValue, config.CheckValue)
	case "gt":
		return checkGt(fieldValue, config.CheckValue)
	case "lt":
		return checkLt(fieldValue, config.CheckValue)
	case "gte":
		return checkGte(fieldValue, config.CheckValue)
	case "lte":
		return checkLte(fieldValue, config.CheckValue)
	case "percent_lt":
		return checkPercentLt(fieldValue, config.CheckValue)
	case "percent_gte":
		return checkPercentGte(fieldValue, config.CheckValue)
	case "ratio_gt":
		return checkRatioGt(fieldValue, config.CheckValue)
	case "array_len_gt":
		return checkArrayLenGt(fieldValue, config.CheckValue)
	default:
		return false, fmt.Sprintf("不支持的检查类型: %s", config.CheckType), nil
	}
}

// checkEmpty 检查是否为空
func checkEmpty(value interface{}) (bool, string, []string) {
	if value == nil {
		return false, "字段值为空", nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return false, "字段值为空", nil
		}
	case int, int64:
		if v == 0 {
			return false, "字段值为0", nil
		}
	}

	return true, "", nil
}

// checkNotEmpty 检查是否不为空
func checkNotEmpty(value interface{}) (bool, string, []string) {
	passed, _, details := checkEmpty(value)
	if passed {
		return false, "", nil // 返回空字符串，使用数据库配置的error_message
	}
	return true, "", details
}

// checkEquals 检查是否相等
func checkEquals(value interface{}, checkValue interface{}) (bool, string, []string) {
	if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", checkValue) {
		return false, fmt.Sprintf("字段值不应等于 %v", checkValue), nil
	}
	return true, "", nil
}

// checkNotEquals 检查是否不相等
func checkNotEquals(value interface{}, checkValue interface{}) (bool, string, []string) {
	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", checkValue) {
		return false, fmt.Sprintf("字段值应等于 %v", checkValue), nil
	}
	return true, "", nil
}

// checkIn 检查是否在列表中
func checkIn(value interface{}, checkValue interface{}) (bool, string, []string) {
	list, ok := checkValue.([]interface{})
	if !ok {
		return false, "检查值格式错误", nil
	}

	valueStr := fmt.Sprintf("%v", value)
	for _, item := range list {
		if fmt.Sprintf("%v", item) == valueStr {
			return true, "", nil
		}
	}

	return false, fmt.Sprintf("字段值 '%v' 不在允许列表中", value), nil
}

// checkNotIn 检查是否不在列表中
func checkNotIn(value interface{}, checkValue interface{}) (bool, string, []string) {
	list, ok := checkValue.([]interface{})
	if !ok {
		return false, "检查值格式错误", nil
	}

	valueStr := fmt.Sprintf("%v", value)
	for _, item := range list {
		if fmt.Sprintf("%v", item) == valueStr {
			return false, fmt.Sprintf("字段值 '%v' 不允许使用", value), nil
		}
	}

	return true, "", nil
}

// checkContains 检查是否包含
func checkContains(value interface{}, checkValue interface{}) (bool, string, []string) {
	valueStr := fmt.Sprintf("%v", value)
	checkStr := fmt.Sprintf("%v", checkValue)

	if strings.Contains(valueStr, checkStr) {
		return false, fmt.Sprintf("字段值不应包含 '%s'", checkStr), nil
	}
	return true, "", nil
}

// checkNotContains 检查是否不包含
func checkNotContains(value interface{}, checkValue interface{}) (bool, string, []string) {
	valueStr := fmt.Sprintf("%v", value)
	checkStr := fmt.Sprintf("%v", checkValue)

	if !strings.Contains(valueStr, checkStr) {
		return false, fmt.Sprintf("字段值应包含 '%s'", checkStr), nil
	}
	return true, "", nil
}

// checkGt 检查是否大于
func checkGt(value interface{}, checkValue interface{}) (bool, string, []string) {
	v1, err1 := toFloat64(value)
	v2, err2 := toFloat64(checkValue)

	if err1 != nil || err2 != nil {
		return false, "数值比较失败", nil
	}

	if v1 > v2 {
		return false, fmt.Sprintf("字段值 %v 超过限制 %v", value, checkValue), nil
	}
	return true, "", nil
}

// checkLt 检查是否小于
func checkLt(value interface{}, checkValue interface{}) (bool, string, []string) {
	v1, err1 := toFloat64(value)
	v2, err2 := toFloat64(checkValue)

	if err1 != nil || err2 != nil {
		return false, "数值比较失败", nil
	}

	if v1 < v2 {
		return false, fmt.Sprintf("字段值 %v 低于要求 %v", value, checkValue), nil
	}
	return true, "", nil
}

// checkGte 检查是否大于等于
func checkGte(value interface{}, checkValue interface{}) (bool, string, []string) {
	v1, err1 := toFloat64(value)
	v2, err2 := toFloat64(checkValue)

	if err1 != nil || err2 != nil {
		return false, "数值比较失败", nil
	}

	if v1 >= v2 {
		return true, "", nil
	}
	return false, fmt.Sprintf("字段值 %v 低于要求 %v", value, checkValue), nil
}

// checkLte 检查是否小于等于
func checkLte(value interface{}, checkValue interface{}) (bool, string, []string) {
	v1, err1 := toFloat64(value)
	v2, err2 := toFloat64(checkValue)

	if err1 != nil || err2 != nil {
		return false, "数值比较失败", nil
	}

	if v1 <= v2 {
		return true, "", nil
	}
	return false, fmt.Sprintf("字段值 %v 超过限制 %v", value, checkValue), nil
}

// checkPercentLt 检查百分比是否小于（从 "x/y (z%)" 格式中提取）
func checkPercentLt(value interface{}, checkValue interface{}) (bool, string, []string) {
	valueStr := fmt.Sprintf("%v", value)

	// 解析 "x/y (z%)" 格式
	parts := strings.Split(valueStr, " ")
	if len(parts) < 2 {
		return false, "百分比格式错误", nil
	}

	percentStr := strings.Trim(parts[1], "()")
	percentStr = strings.TrimSuffix(percentStr, "%")
	percent, err := strconv.Atoi(percentStr)
	if err != nil {
		return false, "百分比解析失败", nil
	}

	threshold, err := toFloat64(checkValue)
	if err != nil {
		return false, "阈值解析失败", nil
	}

	if float64(percent) < threshold {
		return false, fmt.Sprintf("覆盖率为 %d%%，要求至少 %.0f%%", percent, threshold), nil
	}

	return true, "", nil
}

// checkPercentGte 检查百分比是否大于等于
func checkPercentGte(value interface{}, checkValue interface{}) (bool, string, []string) {
	passed, _, details := checkPercentLt(value, checkValue)
	return !passed, "", details
}

// checkRatioGt 检查比率是否大于（从 "x/y" 格式中提取）
func checkRatioGt(value interface{}, checkValue interface{}) (bool, string, []string) {
	valueStr := fmt.Sprintf("%v", value)

	// 解析 "x/y" 格式
	parts := strings.Split(valueStr, "/")
	if len(parts) != 2 {
		return false, "比率格式错误", nil
	}

	numerator, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, "比率解析失败", nil
	}

	threshold, err := toFloat64(checkValue)
	if err != nil {
		return false, "阈值解析失败", nil
	}

	if float64(numerator) > threshold {
		return false, fmt.Sprintf("数量为 %d，超过限制 %.0f", numerator, threshold), nil
	}

	return true, "", nil
}

// checkArrayLenGt 检查数组长度是否大于
func checkArrayLenGt(value interface{}, checkValue interface{}) (bool, string, []string) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false, "字段不是数组类型", nil
	}

	length := val.Len()
	threshold, err := toFloat64(checkValue)
	if err != nil {
		return false, "阈值解析失败", nil
	}

	if float64(length) > threshold {
		return false, fmt.Sprintf("数组长度为 %d，超过限制 %.0f", length, threshold), nil
	}

	return true, "", nil
}

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为数值: %v", value)
	}
}
