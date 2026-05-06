package analyze

import (
	"strings"

	models "cmdb/models/sql"
)

// buildAddColumnDesc 构建添加字段描述
func buildAddColumnDesc(col *models.ColumnDef) string {
	var parts []string
	parts = append(parts, "添加字段 "+col.Name+" ("+col.Type+")")

	if col.Default != "" {
		parts = append(parts, "默认值: "+col.Default)
	} else {
		parts = append(parts, "无默认值")
	}

	if col.Nullable {
		parts = append(parts, "允许NULL")
	} else {
		parts = append(parts, "NOT NULL")
	}

	if col.Comment != "" {
		parts = append(parts, "注释: "+col.Comment)
	} else {
		parts = append(parts, "无注释")
	}

	if col.Position != "" {
		parts = append(parts, "位置: "+col.Position)
	}

	return strings.Join(parts, ", ")
}

// buildModifyColumnDesc 构建修改字段描述
func buildModifyColumnDesc(col *models.ColumnDef) string {
	var parts []string
	parts = append(parts, "修改字段 "+col.Name+" 为 "+col.Type)

	if col.Comment != "" {
		parts = append(parts, "注释: "+col.Comment)
	}

	return strings.Join(parts, ", ")
}

// buildChangeColumnDesc 构建CHANGE字段描述
func buildChangeColumnDesc(oldName string, col *models.ColumnDef) string {
	var parts []string

	if oldName != col.Name {
		parts = append(parts, "字段 "+oldName+" 改名为 "+col.Name)
	}

	parts = append(parts, "类型: "+col.Type)

	if col.Comment != "" {
		parts = append(parts, "注释: "+col.Comment)
	}

	return strings.Join(parts, ", ")
}

// buildAddIndexDesc 构建添加索引描述
func buildAddIndexDesc(idx *models.IndexDef) string {
	typeName := "索引"
	switch idx.Type {
	case "PRIMARY":
		typeName = "主键"
	case "UNIQUE":
		typeName = "唯一索引"
	case "FULLTEXT":
		typeName = "全文索引"
	}

	desc := "添加" + typeName
	if idx.Name != "" && idx.Name != "PRIMARY" {
		desc += " " + idx.Name
	}
	if len(idx.Columns) > 0 {
		desc += " (" + strings.Join(idx.Columns, ", ") + ")"
	}

	return desc
}
