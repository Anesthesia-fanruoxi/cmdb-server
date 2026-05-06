package rule

import (
	"cmdb/api/sql/analyze"
	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"strings"
)

// CheckSQLRulesWithEstimate 带影响行数估算的SQL规则检查
// project 和 database 用于连接数据库估算 UPDATE/DELETE 的影响行数
// 执行顺序：1.调用Agent统一check接口（EXPLAIN+影响行数） → 2.规则检查
func CheckSQLRulesWithEstimate(sqlContent, project, database string) (*models.BatchRuleCheckResult, error) {
	middleware.Logger.Infof("开始检查SQL规则，SQL长度: %d", len(sqlContent))

	// 1. 拆分SQL语句
	sqls := analyze.SplitMultiSQL(sqlContent)
	if len(sqls) == 0 {
		return nil, fmt.Errorf("SQL分析失败，未找到有效的SQL语句")
	}

	middleware.Logger.Infof("SQL拆分完成，共 %d 条SQL", len(sqls))

	// 2. 调用Agent统一check接口（EXPLAIN验证 + 影响行数估算）
	var checkResults []CheckSQLResult
	if project != "" && database != "" {
		var err error
		checkResults, err = CheckSQLWithAgent(sqls, project, database)
		if err != nil {
			middleware.Logger.Errorf("调用Agent check接口失败: %v", err)
			// 失败时继续执行，但不进行语法验证和影响行数估算
		}
	}

	// 3. 分析每条SQL并进行规则检查
	sqlResults := make([]*models.SQLRuleCheckResult, 0)
	hasViolation := false
	hasBlocker := false

	for i, singleSQL := range sqls {
		middleware.Logger.Infof("检查第 %d 条SQL", i+1)

		// 分析单条SQL
		analyzeResult := analyze.AnalyzeSQL(singleSQL)
		middleware.Logger.Infof("第 %d 条SQL类型: %s", i+1, analyzeResult.BasicInfo.SQLType)

		// 提取表名
		tableNames := extractTableNames(analyzeResult)

		// 规则检查结果
		ruleResults := make([]*models.RuleCheckResult, 0)
		violationCount := 0

		// ========== 处理Agent check结果 ==========
		if checkResults != nil && i < len(checkResults) {
			checkResult := checkResults[i]

			// 如果语法验证失败，添加错误并跳过后续检查
			if !checkResult.Valid {
				// 构建详细的错误信息：优先使用Error字段，如果没有则使用Message
				errorMsg := checkResult.Message
				if checkResult.Error != "" {
					errorMsg = checkResult.Error
				}

				syntaxError := &models.RuleCheckResult{
					RuleID:       0,
					RuleName:     "SQL语法验证",
					RuleType:     "syntax_check",
					RuleCategory: "syntax",
					Passed:       false,
					ErrorMessage: errorMsg,
				}

				ruleResults = append(ruleResults, syntaxError)
				violationCount++
				hasViolation = true
				hasBlocker = true
				middleware.Logger.Warnf("SQL语法验证失败: %s", errorMsg)

				// 构建结果，不包含解析结果
				sqlResult := &models.SQLRuleCheckResult{
					SQL:          singleSQL,
					SQLType:      analyzeResult.BasicInfo.SQLType,
					SQLCategory:  analyzeResult.BasicInfo.SQLCategory,
					TableNames:   tableNames,
					RuleResults:  ruleResults,
					HasViolation: true,
					HasBlocker:   true,
				}
				sqlResults = append(sqlResults, sqlResult)
				continue // 跳过当前SQL的后续检查
			}

			// 设置影响行数（仅 UPDATE/DELETE）
			if dmlDetail, ok := analyzeResult.Analyze.(*models.DMLDetail); ok {
				if dmlDetail.Operation == "UPDATE" || dmlDetail.Operation == "DELETE" {
					dmlDetail.EstimatedRows = &checkResult.AffectedCount
					middleware.Logger.Infof("第 %d 条SQL设置影响行数: %d", i+1, checkResult.AffectedCount)
				}
			}
		}

		// ========== 规则检查 ==========
		rules, err := LoadRulesByType(analyzeResult.BasicInfo.SQLType)
		if err != nil {
			middleware.Logger.Errorf("加载规则失败: %v", err)
			continue
		}

		middleware.Logger.Infof("第 %d 条SQL加载了 %d 条规则", i+1, len(rules))

		passedCount := 0
		for _, rule := range rules {
			checkResult := CheckRuleEngine(analyzeResult, rule)
			ruleResults = append(ruleResults, checkResult)

			if checkResult.Passed {
				passedCount++
			} else {
				violationCount++
				hasViolation = true
				hasBlocker = true
				middleware.Logger.Warnf("规则检查失败: [%s] %s", checkResult.RuleName, checkResult.ErrorMessage)
			}
		}

		middleware.Logger.Infof("第 %d 条SQL规则检查完成，通过: %d, 违规: %d", i+1, passedCount, violationCount)

		// 构建结果
		sqlResult := &models.SQLRuleCheckResult{
			SQL:          singleSQL,
			SQLType:      analyzeResult.BasicInfo.SQLType,
			SQLCategory:  analyzeResult.BasicInfo.SQLCategory,
			TableNames:   tableNames,
			RuleResults:  ruleResults,
			HasViolation: violationCount > 0,
			HasBlocker:   violationCount > 0,
			ParseResult:  analyzeResult,
		}

		sqlResults = append(sqlResults, sqlResult)
	}

	// 3. 构建汇总信息
	summary := buildSummary(sqlResults)

	// 4. 构建响应
	result := &models.BatchRuleCheckResult{
		SQLResults:   sqlResults,
		HasViolation: hasViolation,
		HasBlocker:   hasBlocker,
		Summary:      summary,
	}

	middleware.Logger.Infof("SQL规则检查完成，总计: %d条SQL, 违规: %d条, 阻断: %d条",
		summary.TotalSQL, summary.ViolationSQL, summary.BlockerSQL)

	return result, nil
}

// extractTableNames 提取表名
func extractTableNames(analyzeResult *models.AnalyzeResponse) []string {
	tableNames := make([]string, 0)

	if analyzeResult.Analyze == nil {
		return tableNames
	}

	switch detail := analyzeResult.Analyze.(type) {
	case *models.DDLDetail:
		if detail.TableInfo != nil {
			tableNames = append(tableNames, detail.TableInfo.TableName)
		} else if detail.SimpleInfo != nil {
			tableNames = append(tableNames, detail.SimpleInfo.ObjectName)
		} else if detail.AlterInfo != nil {
			tableNames = append(tableNames, detail.AlterInfo.TableName)
		}
	case *models.DMLDetail:
		tableNames = append(tableNames, detail.Table)
	case *models.DQLDetail:
		for _, table := range detail.Tables {
			tableNames = append(tableNames, table.Name)
		}
	}

	return tableNames
}

// buildSummary 构建汇总信息
func buildSummary(sqlResults []*models.SQLRuleCheckResult) *models.CheckSummary {
	summary := &models.CheckSummary{
		TotalSQL:     len(sqlResults),
		PassedSQL:    0,
		ViolationSQL: 0,
		BlockerSQL:   0,
		TotalRules:   0,
	}

	// 统计规则数量
	ruleMap := make(map[int64]bool)

	for _, sqlResult := range sqlResults {
		// 统计SQL
		if !sqlResult.HasViolation {
			summary.PassedSQL++
		} else {
			summary.ViolationSQL++
			if sqlResult.HasBlocker {
				summary.BlockerSQL++
			}
		}

		// 统计规则
		for _, ruleResult := range sqlResult.RuleResults {
			ruleMap[ruleResult.RuleID] = true
			if !ruleResult.Passed {
				summary.ViolatedRules++
			}
		}
	}

	summary.TotalRules = len(ruleMap)

	return summary
}

// LoadRulesByType 根据SQL类型加载适用的规则
func LoadRulesByType(sqlType string) ([]models.SqlRule, error) {
	var rules []models.SqlRule

	// 查询启用的规则，且sql_tag匹配
	err := database.DB.Table(middleware.TableSqlRule).
		Where("is_enabled = ? AND sql_tag = ?", 1, sqlType).
		Find(&rules).Error

	if err != nil {
		middleware.Logger.Errorf("加载规则失败: %v", err)
		return nil, err
	}

	middleware.Logger.Infof("为SQL类型 %s 加载了 %d 条规则", sqlType, len(rules))
	return rules, nil
}

// CheckSQLResult 单条SQL检查结果
type CheckSQLResult struct {
	Success       bool   `json:"success"`
	SQLType       string `json:"sql_type"`
	Valid         bool   `json:"valid"`
	AffectedCount int64  `json:"affected_count"`
	Message       string `json:"message"`
	Error         string `json:"error,omitempty"` // 详细错误信息
	Took          int    `json:"took"`
}

// CheckSQLResponse Agent返回的检查响应
type CheckSQLResponse struct {
	Results []CheckSQLResult `json:"results"`
	Total   int              `json:"total"`
	Took    int              `json:"took"`
	DBName  string           `json:"db_name"`
}

// CheckSQLWithAgent 调用Agent的统一check接口验证SQL
// 返回每条SQL的验证结果
func CheckSQLWithAgent(sqls []string, project, database string) ([]CheckSQLResult, error) {
	// 如果没有提供项目和数据库，返回空结果
	if project == "" || database == "" {
		middleware.Logger.Infof("未提供项目或数据库信息，跳过SQL检查")
		return nil, nil
	}

	// 将多条SQL拼接成一个字符串（用换行分隔）
	sqlContent := ""
	for i, sql := range sqls {
		if i > 0 {
			sqlContent += "\n"
		}

		// 清理SQL：统一换行符
		cleanedSQL := strings.ReplaceAll(sql, "\r\n", "\n")
		cleanedSQL = strings.ReplaceAll(cleanedSQL, "\r", "\n")
		cleanedSQL = strings.TrimSpace(cleanedSQL)

		// 确保SQL末尾有分号
		if !strings.HasSuffix(cleanedSQL, ";") {
			cleanedSQL += ";"
		}

		sqlContent += cleanedSQL
	}

	middleware.Logger.Infof("拼接后的SQL内容（长度: %d）", len(sqlContent))

	// 构建请求
	requestData := map[string]interface{}{
		"sql":    sqlContent,
		"dbName": database,
	}

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/sql-plugs/api/sql/check"
	decompressedData, err := middleware.CallAgentWithBody(project, pluginPath, requestData)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %w", err)
	}

	// 解析响应
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(decompressedData, &rawResponse); err != nil {
		return nil, fmt.Errorf("解析原始响应失败: %v", err)
	}

	// 检查code
	code, ok := rawResponse["code"].(float64)
	if !ok {
		return nil, fmt.Errorf("响应中缺少code字段")
	}

	if int(code) != 200 {
		message, _ := rawResponse["message"].(string)
		return nil, fmt.Errorf("Agent返回错误: %s", message)
	}

	// 提取data字段
	dataField, ok := rawResponse["data"]
	if !ok {
		return nil, fmt.Errorf("响应中缺少data字段")
	}

	// 将data字段转换为JSON再解析
	dataJSON, err := json.Marshal(dataField)
	if err != nil {
		return nil, fmt.Errorf("序列化data字段失败: %v", err)
	}

	var checkResponse CheckSQLResponse
	if err := json.Unmarshal(dataJSON, &checkResponse); err != nil {
		return nil, fmt.Errorf("解析CheckSQLResponse失败: %v, data: %s", err, string(dataJSON))
	}

	middleware.Logger.Infof("SQL检查完成: 总计%d条SQL, 耗时%dms", checkResponse.Total, checkResponse.Took)

	return checkResponse.Results, nil
}
