package rule

import (
	models "cmdb/models/sql"
)

// CollectViolationDetails 收集违规明细
// 根据规则类型和解析结果，返回具体的违规项
func CollectViolationDetails(analyzeResult *models.AnalyzeResponse, rule models.SqlRule) []models.ViolationDetail {
	var details []models.ViolationDetail

	switch rule.RuleName {
	case "字段注释必需":
		details = collectFieldCommentViolations(analyzeResult)
	case "ALTER添加字段必须有注释":
		details = collectAlterFieldCommentViolations(analyzeResult)
	case "主键必需":
		details = collectPrimaryKeyViolation(analyzeResult)
	case "表注释必需":
		details = collectTableCommentViolation(analyzeResult)
	case "索引必需":
		details = collectIndexViolation(analyzeResult)
	case "禁止添加字段时指定位置":
		details = collectFieldPositionViolations(analyzeResult)
	case "禁止删除字段":
		details = collectDropFieldViolations(analyzeResult)
	case "禁止自定义存储引擎":
		details = collectEngineViolation(analyzeResult)
	case "禁止自定义字符集":
		details = collectCharsetViolation(analyzeResult)
	case "禁止自定义排序规则":
		details = collectCollationViolation(analyzeResult)
	}

	return details
}

// collectFieldCommentViolations 收集缺少注释的字段（CREATE TABLE）
func collectFieldCommentViolations(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.FieldDetail == nil {
		return details
	}

	for _, col := range ddlDetail.FieldDetail {
		if col.Comment == "" {
			details = append(details, models.ViolationDetail{
				Target:     col.Name,
				TargetType: "column",
				Issue:      "缺少注释",
			})
		}
	}

	return details
}

// collectAlterFieldCommentViolations 收集ALTER添加字段时缺少注释的字段
func collectAlterFieldCommentViolations(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.AlterDetail == nil {
		return details
	}

	for _, act := range ddlDetail.AlterDetail {
		if act.Action == "ADD" && act.Target == "COLUMN" && act.Column != nil {
			if act.Column.Comment == "" {
				details = append(details, models.ViolationDetail{
					Target:     act.Column.Name,
					TargetType: "column",
					Issue:      "添加字段时缺少注释",
				})
			}
		}
	}

	return details
}

// collectPrimaryKeyViolation 收集主键缺失违规
func collectPrimaryKeyViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.PrimaryKey == "" {
		details = append(details, models.ViolationDetail{
			Target:     ddlDetail.TableInfo.TableName,
			TargetType: "table",
			Issue:      "缺少主键",
		})
	}

	return details
}

// collectTableCommentViolation 收集表注释缺失违规
func collectTableCommentViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.Comment == "" {
		details = append(details, models.ViolationDetail{
			Target:     ddlDetail.TableInfo.TableName,
			TargetType: "table",
			Issue:      "缺少表注释",
		})
	}

	return details
}

// collectIndexViolation 收集索引不足违规
func collectIndexViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.IndexCount < 2 {
		details = append(details, models.ViolationDetail{
			Target:        ddlDetail.TableInfo.TableName,
			TargetType:    "table",
			Issue:         "索引数量不足",
			CurrentValue:  string(rune(ddlDetail.TableInfo.IndexCount + '0')),
			ExpectedValue: "≥2（主键+至少一个额外索引）",
		})
	}

	return details
}

// collectFieldPositionViolations 收集指定位置的字段违规
func collectFieldPositionViolations(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.AlterDetail == nil {
		return details
	}

	for _, act := range ddlDetail.AlterDetail {
		if act.Action == "ADD" && act.Target == "COLUMN" && act.Position != "" {
			details = append(details, models.ViolationDetail{
				Target:       act.Column.Name,
				TargetType:   "column",
				Issue:        "添加字段时指定了位置",
				CurrentValue: act.Position,
			})
		}
	}

	return details
}

// collectDropFieldViolations 收集删除字段违规
func collectDropFieldViolations(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.AlterDetail == nil {
		return details
	}

	for _, act := range ddlDetail.AlterDetail {
		if act.Action == "DROP" && act.Target == "COLUMN" {
			details = append(details, models.ViolationDetail{
				Target:     act.OldName,
				TargetType: "column",
				Issue:      "删除字段",
			})
		}
	}

	return details
}

// collectEngineViolation 收集存储引擎违规
func collectEngineViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.Engine != "" {
		details = append(details, models.ViolationDetail{
			Target:       ddlDetail.TableInfo.TableName,
			TargetType:   "table",
			Issue:        "自定义了存储引擎",
			CurrentValue: ddlDetail.TableInfo.Engine,
		})
	}

	return details
}

// collectCharsetViolation 收集字符集违规
func collectCharsetViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.Charset != "" {
		details = append(details, models.ViolationDetail{
			Target:       ddlDetail.TableInfo.TableName,
			TargetType:   "table",
			Issue:        "自定义了字符集",
			CurrentValue: ddlDetail.TableInfo.Charset,
		})
	}

	return details
}

// collectCollationViolation 收集排序规则违规
func collectCollationViolation(analyzeResult *models.AnalyzeResponse) []models.ViolationDetail {
	var details []models.ViolationDetail

	ddlDetail, ok := analyzeResult.Analyze.(*models.DDLDetail)
	if !ok || ddlDetail.TableInfo == nil {
		return details
	}

	if ddlDetail.TableInfo.Collation != "" {
		details = append(details, models.ViolationDetail{
			Target:       ddlDetail.TableInfo.TableName,
			TargetType:   "table",
			Issue:        "自定义了排序规则",
			CurrentValue: ddlDetail.TableInfo.Collation,
		})
	}

	return details
}
