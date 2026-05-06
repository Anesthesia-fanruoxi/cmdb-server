package analyze

import (
	"regexp"
	"strings"

	models "cmdb/models/sql"
)

// parseColumnDef 解析字段定义
func parseColumnDef(def string) models.ColumnDef {
	col := models.ColumnDef{Nullable: true}

	def = regexp.MustCompile(`\s+`).ReplaceAllString(def, " ")
	def = strings.TrimSpace(def)

	parts := strings.SplitN(def, " ", 2)
	if len(parts) < 1 {
		return col
	}
	col.Name = strings.Trim(parts[0], "`")

	if len(parts) < 2 {
		return col
	}

	rest := parts[1]
	upper := strings.ToUpper(rest)

	// 提取类型
	if re := regexp.MustCompile(`^(\w+(?:\([^)]+\))?)`); true {
		if m := re.FindStringSubmatch(rest); len(m) > 1 {
			col.Type = strings.ToUpper(m[1])
		}
	}

	// 检查约束
	col.Nullable = !strings.Contains(upper, "NOT NULL")
	col.PrimaryKey = strings.Contains(upper, "PRIMARY KEY")
	col.AutoIncr = strings.Contains(upper, "AUTO_INCREMENT")
	col.Unique = strings.Contains(upper, " UNIQUE")

	// 提取默认值
	if re := regexp.MustCompile(`(?i)DEFAULT\s+('([^']*)'|"([^"]*)"|(\S+))`); true {
		if m := re.FindStringSubmatch(rest); len(m) > 1 {
			if m[2] != "" {
				col.Default = m[2]
			} else if m[3] != "" {
				col.Default = m[3]
			} else {
				col.Default = m[4]
			}
		}
	}

	// 提取注释
	if re := regexp.MustCompile(`(?i)COMMENT\s+'([^']*)'`); true {
		if m := re.FindStringSubmatch(rest); len(m) > 1 {
			col.Comment = m[1]
		}
	}

	// 提取位置信息 FIRST / AFTER xxx
	if strings.Contains(upper, " FIRST") {
		col.Position = "FIRST"
	} else if re := regexp.MustCompile(`(?i)AFTER\s+` + "`?" + `(\w+)` + "`?"); true {
		if m := re.FindStringSubmatch(rest); len(m) > 1 {
			col.Position = "AFTER " + m[1]
		}
	}

	// 提取排序规则 COLLATE xxx
	if re := regexp.MustCompile(`(?i)COLLATE\s+(\w+)`); true {
		if m := re.FindStringSubmatch(rest); len(m) > 1 {
			col.Collate = m[1]
		}
	}

	return col
}

// parseIndexDef 解析索引定义
func parseIndexDef(def string, idxType string) models.IndexDef {
	idx := models.IndexDef{Type: idxType}

	// 提取索引名
	re := regexp.MustCompile(`(?i)(?:KEY|INDEX)\s+` + "`?" + `(\w+)` + "`?" + `\s*\(`)
	if m := re.FindStringSubmatch(def); len(m) > 1 {
		idx.Name = m[1]
	}

	// 提取字段列表
	if start := strings.Index(def, "("); start >= 0 {
		if end := strings.LastIndex(def, ")"); end > start {
			colPart := def[start+1 : end]
			cols := strings.Split(colPart, ",")
			for _, c := range cols {
				c = strings.TrimSpace(c)
				c = strings.Trim(c, "`")
				c = regexp.MustCompile(`\s*\(\d+\)`).ReplaceAllString(c, "")
				c = regexp.MustCompile(`(?i)\s+(ASC|DESC)`).ReplaceAllString(c, "")
				c = strings.TrimSpace(c)
				if c != "" {
					idx.Columns = append(idx.Columns, c)
				}
			}
		}
	}

	return idx
}

// parseTableOptions 解析表选项
func parseTableOptions(optPart string) *models.TableOptions {
	opts := &models.TableOptions{}

	if re := regexp.MustCompile(`(?i)ENGINE\s*=\s*(\w+)`); true {
		if m := re.FindStringSubmatch(optPart); len(m) > 1 {
			opts.Engine = m[1]
		}
	}

	if re := regexp.MustCompile(`(?i)(?:DEFAULT\s+)?(?:CHARACTER\s+SET|CHARSET)\s*=?\s*(\w+)`); true {
		if m := re.FindStringSubmatch(optPart); len(m) > 1 {
			opts.Charset = m[1]
		}
	}

	if re := regexp.MustCompile(`(?i)COLLATE\s*=?\s*(\w+)`); true {
		if m := re.FindStringSubmatch(optPart); len(m) > 1 {
			opts.Collate = m[1]
		}
	}

	if re := regexp.MustCompile(`(?i)COMMENT\s*=?\s*'([^']*)'`); true {
		if m := re.FindStringSubmatch(optPart); len(m) > 1 {
			opts.Comment = m[1]
		}
	}

	return opts
}

// parseAlterAction 解析单个ALTER操作
func parseAlterAction(act string) models.AlterAction {
	action := models.AlterAction{}
	upper := strings.ToUpper(act)

	switch {
	// ========== 字段操作 ==========
	case strings.HasPrefix(upper, "ADD COLUMN") || (regexp.MustCompile(`(?i)^ADD\s+`+"`?"+`\w+`).MatchString(act) &&
		!strings.HasPrefix(upper, "ADD INDEX") && !strings.HasPrefix(upper, "ADD KEY") &&
		!strings.HasPrefix(upper, "ADD UNIQUE") && !strings.HasPrefix(upper, "ADD PRIMARY") &&
		!strings.HasPrefix(upper, "ADD FULLTEXT") && !strings.HasPrefix(upper, "ADD SPATIAL") &&
		!strings.HasPrefix(upper, "ADD CONSTRAINT")):
		action.Action = "ADD"
		action.Target = "COLUMN"
		colDef := regexp.MustCompile(`(?i)^ADD\s+(?:COLUMN\s+)?`).ReplaceAllString(act, "")
		col := parseColumnDef(colDef)
		action.Column = &col
		action.Position = col.Position // 复制位置信息到操作详情
		action.Description = buildAddColumnDesc(&col)
		action.ChangeSummary = "新增字段"

	case strings.HasPrefix(upper, "DROP COLUMN"):
		action.Action = "DROP"
		action.Target = "COLUMN"
		re := regexp.MustCompile(`(?i)DROP\s+COLUMN\s+` + "`?" + `(\w+)` + "`?")
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.OldName = m[1]
			action.Description = "删除字段 " + m[1]
			action.ChangeSummary = "删除字段"
		}

	case strings.HasPrefix(upper, "MODIFY"):
		action.Action = "MODIFY"
		action.Target = "COLUMN"
		colDef := regexp.MustCompile(`(?i)^MODIFY\s+(?:COLUMN\s+)?`).ReplaceAllString(act, "")
		col := parseColumnDef(colDef)
		action.Column = &col
		action.Position = col.Position // 复制位置信息到操作详情
		action.Description = buildModifyColumnDesc(&col)
		action.ChangeSummary = "类型 → " + col.Type

	case strings.HasPrefix(upper, "CHANGE"):
		action.Action = "CHANGE"
		action.Target = "COLUMN"
		re := regexp.MustCompile(`(?i)^CHANGE\s+(?:COLUMN\s+)?` + "`?" + `(\w+)` + "`?" + `\s+`)
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.OldName = m[1]
			rest := re.ReplaceAllString(act, "")
			col := parseColumnDef(rest)
			action.Column = &col
			action.Position = col.Position // 复制位置信息到操作详情
			action.NewName = col.Name
			action.Description = buildChangeColumnDesc(m[1], &col)
			// 如果改名了显示改名，否则显示类型变化
			if m[1] != col.Name {
				action.ChangeSummary = m[1] + " → " + col.Name
			} else {
				action.ChangeSummary = "类型 → " + col.Type
			}
		}

	case strings.HasPrefix(upper, "ALTER COLUMN"):
		action.Action = "ALTER"
		action.Target = "COLUMN"
		if strings.Contains(upper, "SET DEFAULT") {
			action.Description = "设置字段默认值"
			action.ChangeSummary = "设置默认值"
		} else if strings.Contains(upper, "DROP DEFAULT") {
			action.Description = "删除字段默认值"
			action.ChangeSummary = "删除默认值"
		}

	// ========== 索引操作 ==========
	case strings.HasPrefix(upper, "ADD SPATIAL"):
		action.Action = "ADD"
		action.Target = "INDEX"
		idx := parseAlterIndexDef(act, "SPATIAL")
		action.Index = &idx
		action.Description = buildAddIndexDesc(&idx)
		action.ChangeSummary = "新增空间索引"

	case strings.HasPrefix(upper, "ADD UNIQUE"):
		action.Action = "ADD"
		action.Target = "INDEX"
		idx := parseAlterIndexDef(act, "UNIQUE")
		action.Index = &idx
		action.Description = buildAddIndexDesc(&idx)
		action.ChangeSummary = "新增唯一索引"

	case strings.HasPrefix(upper, "ADD PRIMARY"):
		action.Action = "ADD"
		action.Target = "INDEX"
		idx := parseAlterIndexDef(act, "PRIMARY")
		action.Index = &idx
		action.Description = buildAddIndexDesc(&idx)
		action.ChangeSummary = "新增主键"

	case strings.HasPrefix(upper, "ADD FULLTEXT"):
		action.Action = "ADD"
		action.Target = "INDEX"
		idx := parseAlterIndexDef(act, "FULLTEXT")
		action.Index = &idx
		action.Description = buildAddIndexDesc(&idx)
		action.ChangeSummary = "新增全文索引"

	case strings.HasPrefix(upper, "ADD INDEX") || strings.HasPrefix(upper, "ADD KEY"):
		action.Action = "ADD"
		action.Target = "INDEX"
		idx := parseAlterIndexDef(act, "INDEX")
		action.Index = &idx
		action.Description = buildAddIndexDesc(&idx)
		action.ChangeSummary = "新增索引"

	case strings.HasPrefix(upper, "DROP PRIMARY"):
		action.Action = "DROP"
		action.Target = "INDEX"
		action.Index = &models.IndexDef{Name: "PRIMARY", Type: "PRIMARY"}
		action.Description = "删除主键"
		action.ChangeSummary = "删除主键"

	case strings.HasPrefix(upper, "DROP INDEX") || strings.HasPrefix(upper, "DROP KEY"):
		action.Action = "DROP"
		action.Target = "INDEX"
		re := regexp.MustCompile(`(?i)DROP\s+(?:INDEX|KEY)\s+` + "`?" + `(\w+)` + "`?")
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.Index = &models.IndexDef{Name: m[1]}
			action.Description = "删除索引 " + m[1]
			action.ChangeSummary = "删除索引"
		}

	// ========== 约束操作 ==========
	case strings.HasPrefix(upper, "ADD CONSTRAINT"):
		action.Action = "ADD"
		action.Target = "CONSTRAINT"
		if strings.Contains(upper, "FOREIGN KEY") {
			action.Description = "添加外键约束"
			action.ChangeSummary = "新增外键"
		} else {
			action.Description = "添加约束"
			action.ChangeSummary = "新增约束"
		}

	case strings.HasPrefix(upper, "DROP FOREIGN"):
		action.Action = "DROP"
		action.Target = "CONSTRAINT"
		re := regexp.MustCompile(`(?i)DROP\s+FOREIGN\s+KEY\s+` + "`?" + `(\w+)` + "`?")
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.OldName = m[1]
			action.Description = "删除外键 " + m[1]
			action.ChangeSummary = "删除外键"
		}

	// ========== 表属性操作 ==========
	case strings.HasPrefix(upper, "RENAME TO") || strings.HasPrefix(upper, "RENAME AS"):
		action.Action = "RENAME"
		action.Target = "TABLE"
		re := regexp.MustCompile(`(?i)RENAME\s+(?:TO|AS)\s+` + "`?" + `(\w+)` + "`?")
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.NewName = m[1]
			action.Description = "重命名表为 " + m[1]
			action.ChangeSummary = "重命名表"
		}

	case strings.HasPrefix(upper, "ENGINE"):
		action.Action = "MODIFY"
		action.Target = "TABLE_ATTR"
		re := regexp.MustCompile(`(?i)ENGINE\s*=\s*(\w+)`)
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.Description = "修改存储引擎为 " + m[1]
			action.ChangeSummary = "ENGINE → " + m[1]
		}

	case strings.HasPrefix(upper, "CHARACTER SET") || strings.HasPrefix(upper, "CHARSET"):
		action.Action = "MODIFY"
		action.Target = "TABLE_ATTR"
		action.Description = "修改字符集"
		action.ChangeSummary = "修改字符集"

	case strings.HasPrefix(upper, "CONVERT TO"):
		action.Action = "CONVERT"
		action.Target = "TABLE_ATTR"
		action.Description = "转换字符集（会转换数据）"
		action.ChangeSummary = "转换字符集"

	case strings.HasPrefix(upper, "COMMENT"):
		action.Action = "MODIFY"
		action.Target = "TABLE_ATTR"
		re := regexp.MustCompile(`(?i)COMMENT\s*=\s*'([^']*)'`)
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.Description = "修改表注释为: " + m[1]
			action.ChangeSummary = "修改表注释"
		}

	case strings.HasPrefix(upper, "AUTO_INCREMENT"):
		action.Action = "MODIFY"
		action.Target = "TABLE_ATTR"
		re := regexp.MustCompile(`(?i)AUTO_INCREMENT\s*=\s*(\d+)`)
		if m := re.FindStringSubmatch(act); len(m) > 1 {
			action.Description = "设置自增起始值为 " + m[1]
			action.ChangeSummary = "AUTO_INCREMENT → " + m[1]
		}
	}

	return action
}

// parseAlterIndexDef 解析ALTER中的索引定义
func parseAlterIndexDef(act string, idxType string) models.IndexDef {
	idx := models.IndexDef{Type: idxType}

	// 提取索引名
	var nameRe *regexp.Regexp
	switch idxType {
	case "PRIMARY":
		idx.Name = "PRIMARY"
	case "UNIQUE":
		nameRe = regexp.MustCompile(`(?i)UNIQUE\s+(?:INDEX|KEY)?\s*` + "`?" + `(\w+)` + "`?" + `\s*\(`)
	case "FULLTEXT":
		nameRe = regexp.MustCompile(`(?i)FULLTEXT\s+(?:INDEX|KEY)?\s*` + "`?" + `(\w+)` + "`?" + `\s*\(`)
	case "SPATIAL":
		nameRe = regexp.MustCompile(`(?i)SPATIAL\s+(?:INDEX|KEY)?\s*` + "`?" + `(\w+)` + "`?" + `\s*\(`)
	default:
		nameRe = regexp.MustCompile(`(?i)(?:INDEX|KEY)\s+` + "`?" + `(\w+)` + "`?" + `\s*\(`)
	}

	if nameRe != nil {
		if m := nameRe.FindStringSubmatch(act); len(m) > 1 {
			idx.Name = m[1]
		}
	}

	// 提取字段列表
	if start := strings.Index(act, "("); start >= 0 {
		if end := strings.LastIndex(act, ")"); end > start {
			colPart := act[start+1 : end]
			cols := strings.Split(colPart, ",")
			for _, c := range cols {
				c = strings.TrimSpace(c)
				c = strings.Trim(c, "`")
				c = regexp.MustCompile(`\s*\(\d+\)`).ReplaceAllString(c, "")
				c = regexp.MustCompile(`(?i)\s+(ASC|DESC)`).ReplaceAllString(c, "")
				c = strings.TrimSpace(c)
				if c != "" {
					idx.Columns = append(idx.Columns, c)
				}
			}
		}
	}

	return idx
}

// extractName 提取对象名称（支持 database.table 格式）
func extractName(sql, objType string) string {
	// 支持 database.table 格式：`db`.`table` 或 db.table
	pattern := `(?i)` + objType + `\s+(?:IF\s+(?:NOT\s+)?EXISTS\s+)?` + "`?" + `([\w]+)` + "`?" + `(?:\.` + "`?" + `([\w]+)` + "`?" + `)?`
	re := regexp.MustCompile(pattern)
	if m := re.FindStringSubmatch(sql); len(m) > 1 {
		if m[2] != "" {
			// database.table 格式，返回完整名称
			return m[1] + "." + m[2]
		}
		return m[1]
	}
	return ""
}

// splitDefinitions 拆分定义（考虑括号嵌套和反引号）
func splitDefinitions(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inString := false
	inBacktick := false
	stringChar := rune(0)

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		// 处理反引号
		if c == '`' {
			inBacktick = !inBacktick
			current.WriteRune(c)
			continue
		}

		// 在反引号内，直接添加字符
		if inBacktick {
			current.WriteRune(c)
			continue
		}

		// 处理字符串引号
		if (c == '\'' || c == '"') && (i == 0 || runes[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
			}
			current.WriteRune(c)
			continue
		}

		// 在字符串内，直接添加字符
		if inString {
			current.WriteRune(c)
			continue
		}

		// 处理括号
		if c == '(' {
			depth++
			current.WriteRune(c)
		} else if c == ')' {
			depth--
			current.WriteRune(c)
		} else if c == ',' && depth == 0 {
			// 只有在括号外的逗号才分割
			trimmed := strings.TrimSpace(current.String())
			if trimmed != "" {
				result = append(result, trimmed)
			}
			current.Reset()
		} else {
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		trimmed := strings.TrimSpace(current.String())
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseTableDefinitions 解析表定义
func parseTableDefinitions(defPart string, columns *[]models.ColumnDef, indexes *[]models.IndexDef) {
	items := splitDefinitions(defPart)

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		upper := strings.ToUpper(item)

		switch {
		case strings.HasPrefix(upper, "PRIMARY KEY"):
			*indexes = append(*indexes, parseIndexDef(item, "PRIMARY"))
		case strings.HasPrefix(upper, "UNIQUE"):
			*indexes = append(*indexes, parseIndexDef(item, "UNIQUE"))
		case strings.HasPrefix(upper, "KEY ") || strings.HasPrefix(upper, "INDEX "):
			*indexes = append(*indexes, parseIndexDef(item, "INDEX"))
		case strings.HasPrefix(upper, "FULLTEXT"):
			*indexes = append(*indexes, parseIndexDef(item, "FULLTEXT"))
		case strings.HasPrefix(upper, "FOREIGN KEY") || strings.HasPrefix(upper, "CONSTRAINT"):
			continue
		default:
			col := parseColumnDef(item)
			if col.Name != "" {
				*columns = append(*columns, col)
			}
		}
	}
}
