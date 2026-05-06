package analyze

import (
	"fmt"
	"regexp"
	"strings"

	models "cmdb/models/sql"
)

// AnalyzeDDL 分析DDL语句
func AnalyzeDDL(sql, sqlType string) *models.AnalyzeResponse {
	normalized := NormalizeSQL(sql)

	result := &models.AnalyzeResponse{
		BasicInfo: &models.ResponseBasicInfo{
			SQLType:     sqlType,
			SQLCategory: "DDL",
		},
	}

	detail := &models.DDLDetail{}

	// 根据操作类型分析
	switch sqlType {
	case "CREATE":
		parseCreate(normalized, detail)
	case "ALTER":
		parseAlter(normalized, detail)
	case "DROP":
		parseDrop(normalized, detail)
	case "TRUNCATE":
		parseTruncate(normalized, detail)
	case "RENAME":
		parseRename(normalized, detail)
	}

	result.Analyze = detail
	evaluateDDLRisk(result, detail, sqlType)

	return result
}

// parseCreate 解析CREATE语句
func parseCreate(sql string, detail *models.DDLDetail) {
	upper := strings.ToUpper(sql)

	if strings.Contains(upper, " TABLE ") {
		parseCreateTable(sql, detail)
	}
	// 其他类型暂不处理详细信息
}

// parseCreateTable 解析CREATE TABLE
func parseCreateTable(sql string, detail *models.DDLDetail) {
	tableName := extractName(sql, "TABLE")

	start := strings.Index(sql, "(")
	end := strings.LastIndex(sql, ")")
	if start < 0 || end <= start {
		return
	}

	defPart := sql[start+1 : end]
	optPart := ""
	if end+1 < len(sql) {
		optPart = sql[end+1:]
	}

	// 解析字段和索引
	var columns []models.ColumnDef
	var indexes []models.IndexDef
	parseTableDefinitions(defPart, &columns, &indexes)

	// 解析表选项
	var opts *models.TableOptions
	if optPart != "" {
		opts = parseTableOptions(optPart)
	}

	// 构建DDL详情
	buildDDLDetail(detail, tableName, columns, indexes, opts)
}

// buildDDLDetail 构建DDL详情结构
func buildDDLDetail(detail *models.DDLDetail, tableName string, columns []models.ColumnDef, indexes []models.IndexDef, opts *models.TableOptions) {
	// 1. 表信息
	detail.TableInfo = &models.DDLTableInfo{
		TableName:  tableName,
		FieldCount: len(columns),
		IndexCount: len(indexes),
		PrimaryKey: "", // 默认为空
	}
	if opts != nil {
		detail.TableInfo.Engine = opts.Engine
		detail.TableInfo.Charset = opts.Charset
		detail.TableInfo.Collation = opts.Collate
		detail.TableInfo.Comment = opts.Comment
	}

	// 提取主键字段
	for _, idx := range indexes {
		if idx.Type == "PRIMARY" {
			detail.TableInfo.PrimaryKey = strings.Join(idx.Columns, ",")
			break
		}
	}

	// 2. 字段信息汇总（格式化为字符串）
	totalCount := len(columns)
	commentCount := 0
	defaultCount := 0
	nullableCount := 0

	for _, col := range columns {
		if col.Comment != "" {
			commentCount++
		}
		if col.Default != "" || col.AutoIncr || col.PrimaryKey {
			defaultCount++
		}
		if col.Nullable {
			nullableCount++
		}
	}

	// 统计自定义排序规则的字段
	collateCount := 0
	for _, col := range columns {
		if col.Collate != "" {
			collateCount++
		}
	}

	commentPercent := 0
	if totalCount > 0 {
		commentPercent = (commentCount * 100) / totalCount
	}

	detail.FieldInfo = &models.DDLFieldInfo{
		CommentSummary:  fmt.Sprintf("%d/%d (%d%%)", commentCount, totalCount, commentPercent),
		DefaultSummary:  fmt.Sprintf("%d/%d", defaultCount, totalCount),
		NullableSummary: fmt.Sprintf("%d/%d", nullableCount, totalCount),
		CollateSummary:  fmt.Sprintf("%d/%d", collateCount, totalCount),
	}

	// 3. 索引详情（表格展示）
	detail.IndexDetail = indexes

	// 4. 字段详情（表格展示）
	detail.FieldDetail = columns
}

// parseAlter 解析ALTER语句
func parseAlter(sql string, detail *models.DDLDetail) {
	tableName := extractName(sql, "TABLE")

	// 提取ALTER TABLE xxx 后面的操作部分
	// 支持 `database`.`table` 格式
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+` + "`?[\\w]+`?\\.?`?[\\w]+`?" + `\s+`)
	actionPart := re.ReplaceAllString(sql, "")

	var actions []models.AlterAction
	actionItems := splitDefinitions(actionPart)

	for _, act := range actionItems {
		act = strings.TrimSpace(act)
		if act == "" {
			continue
		}
		action := parseAlterAction(act)
		if action.Action != "" {
			actions = append(actions, action)
		}
	}

	// 统计各类操作数量
	totalCount := len(actions)
	addFieldCount := 0
	modifyFieldCount := 0 // 包含 MODIFY 和 ALTER COLUMN
	changeFieldCount := 0
	dropFieldCount := 0
	addIndexCount := 0
	dropIndexCount := 0
	constraintCount := 0
	tableAttrCount := 0
	addFieldWithPositionCount := 0   // 添加字段时指定位置的数量
	addFieldWithoutCommentCount := 0 // 添加字段时缺少注释的数量
	fieldWithCollateCount := 0       // 字段自定义排序规则的数量

	for _, act := range actions {
		switch {
		case act.Action == "ADD" && act.Target == "COLUMN":
			addFieldCount++
			// 检查是否指定了位置（FIRST/AFTER）
			if act.Column != nil && act.Column.Position != "" {
				addFieldWithPositionCount++
			}
			// 检查是否缺少注释
			if act.Column != nil && act.Column.Comment == "" {
				addFieldWithoutCommentCount++
			}
			// 检查是否自定义排序规则
			if act.Column != nil && act.Column.Collate != "" {
				fieldWithCollateCount++
			}
		case act.Action == "MODIFY" && act.Target == "COLUMN":
			modifyFieldCount++
			// MODIFY也检查排序规则
			if act.Column != nil && act.Column.Collate != "" {
				fieldWithCollateCount++
			}
		case act.Action == "ALTER" && act.Target == "COLUMN":
			modifyFieldCount++ // ALTER COLUMN 算作修改字段
		case act.Action == "CHANGE" && act.Target == "COLUMN":
			changeFieldCount++
			// CHANGE也检查排序规则
			if act.Column != nil && act.Column.Collate != "" {
				fieldWithCollateCount++
			}
		case act.Action == "DROP" && act.Target == "COLUMN":
			dropFieldCount++
		case act.Action == "ADD" && act.Target == "INDEX":
			addIndexCount++
		case act.Action == "DROP" && act.Target == "INDEX":
			dropIndexCount++
		case act.Target == "CONSTRAINT":
			constraintCount++
		case act.Target == "TABLE" || act.Target == "TABLE_ATTR":
			tableAttrCount++
		}
	}

	detail.AlterInfo = &models.DDLAlterInfo{
		TableName:                   tableName,
		TotalActions:                totalCount,
		AddFieldCount:               fmt.Sprintf("%d/%d", addFieldCount, totalCount),
		ModifyFieldCount:            fmt.Sprintf("%d/%d", modifyFieldCount, totalCount),
		ChangeFieldCount:            fmt.Sprintf("%d/%d", changeFieldCount, totalCount),
		DropFieldCount:              fmt.Sprintf("%d/%d", dropFieldCount, totalCount),
		AddIndexCount:               fmt.Sprintf("%d/%d", addIndexCount, totalCount),
		DropIndexCount:              fmt.Sprintf("%d/%d", dropIndexCount, totalCount),
		ConstraintActions:           fmt.Sprintf("%d/%d", constraintCount, totalCount),
		TableAttrActions:            fmt.Sprintf("%d/%d", tableAttrCount, totalCount),
		AddFieldWithPositionCount:   fmt.Sprintf("%d/%d", addFieldWithPositionCount, totalCount),
		AddFieldWithoutCommentCount: fmt.Sprintf("%d/%d", addFieldWithoutCommentCount, totalCount),
		FieldWithCollateCount:       fmt.Sprintf("%d/%d", fieldWithCollateCount, totalCount),
	}

	detail.AlterDetail = actions
}

// parseDrop 解析DROP语句
func parseDrop(sql string, detail *models.DDLDetail) {
	upper := strings.ToUpper(sql)

	info := &models.DDLSimpleInfo{
		Operation: "DROP",
		IfExists:  strings.Contains(upper, "IF EXISTS"),
	}

	switch {
	case strings.Contains(upper, " TABLE "):
		info.ObjectType = "TABLE"
		info.ObjectName = extractName(sql, "TABLE")
	case strings.Contains(upper, " INDEX "):
		info.ObjectType = "INDEX"
		info.ObjectName = extractName(sql, "INDEX")
	case strings.Contains(upper, " VIEW "):
		info.ObjectType = "VIEW"
		info.ObjectName = extractName(sql, "VIEW")
	case strings.Contains(upper, " DATABASE "):
		info.ObjectType = "DATABASE"
		info.ObjectName = extractName(sql, "DATABASE")
	}

	detail.SimpleInfo = info
}

// parseTruncate 解析TRUNCATE语句
func parseTruncate(sql string, detail *models.DDLDetail) {
	info := &models.DDLSimpleInfo{
		Operation:  "TRUNCATE",
		ObjectType: "TABLE",
		ObjectName: extractName(sql, "TABLE"),
	}

	if info.ObjectName == "" {
		re := regexp.MustCompile(`(?i)TRUNCATE\s+` + "`?" + `(\w+)` + "`?")
		if m := re.FindStringSubmatch(sql); len(m) > 1 {
			info.ObjectName = m[1]
		}
	}

	detail.SimpleInfo = info
}

// parseRename 解析RENAME语句
func parseRename(sql string, detail *models.DDLDetail) {
	info := &models.DDLSimpleInfo{
		Operation:  "RENAME",
		ObjectType: "TABLE",
	}

	re := regexp.MustCompile(`(?i)RENAME\s+TABLE\s+` + "`?" + `(\w+)` + "`?" + `\s+TO\s+` + "`?" + `(\w+)` + "`?")
	if m := re.FindStringSubmatch(sql); len(m) > 2 {
		info.ObjectName = m[1]
		info.NewName = m[2]
	}

	detail.SimpleInfo = info
}

// evaluateDDLRisk 评估DDL风险
func evaluateDDLRisk(result *models.AnalyzeResponse, detail *models.DDLDetail, sqlType string) {
	switch sqlType {
	case "CREATE":
		result.BasicInfo.RiskLevel = "safe"
		result.BasicInfo.RiskReason = "CREATE操作，创建新对象"

		// CREATE TABLE 额外检查
		if detail.TableInfo != nil {
			var warnings []string

			if detail.TableInfo.PrimaryKey == "" {
				warnings = append(warnings, "缺少主键")
			}
			if detail.FieldInfo != nil && !strings.HasPrefix(detail.FieldInfo.CommentSummary, fmt.Sprintf("%d/%d", detail.TableInfo.FieldCount, detail.TableInfo.FieldCount)) {
				warnings = append(warnings, "部分字段缺少注释")
			}
			if detail.TableInfo.Comment == "" {
				warnings = append(warnings, "缺少表注释")
			}
			if detail.TableInfo.Charset == "" {
				warnings = append(warnings, "未指定字符集")
			}

			if len(warnings) > 0 {
				result.BasicInfo.RiskLevel = "warning"
				result.BasicInfo.RiskReason = "CREATE TABLE: " + strings.Join(warnings, "、")
			}
		}

	case "ALTER":
		result.BasicInfo.RiskLevel = "warning"
		result.BasicInfo.RiskReason = "ALTER操作，修改结构"
		if detail.AlterDetail != nil {
			for _, act := range detail.AlterDetail {
				if act.Action == "DROP" && act.Target == "COLUMN" {
					result.BasicInfo.RiskLevel = "danger"
					result.BasicInfo.RiskReason = "ALTER TABLE DROP COLUMN，删除字段将丢失数据！"
					return
				}
			}
		}

	case "DROP":
		result.BasicInfo.RiskLevel = "danger"
		result.BasicInfo.RiskReason = "DROP操作，删除对象"
		if detail.SimpleInfo != nil && detail.SimpleInfo.ObjectType == "INDEX" {
			result.BasicInfo.RiskLevel = "warning"
			result.BasicInfo.RiskReason = "DROP INDEX，删除索引可能影响性能"
		}

	case "TRUNCATE":
		result.BasicInfo.RiskLevel = "danger"
		result.BasicInfo.RiskReason = "TRUNCATE TABLE，清空所有数据！"

	case "RENAME":
		result.BasicInfo.RiskLevel = "warning"
		result.BasicInfo.RiskReason = "RENAME TABLE，重命名表"
	}
}
