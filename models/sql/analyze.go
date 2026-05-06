package sql

// AnalyzeRequest SQL分析请求
type AnalyzeRequest struct {
	SQL string `json:"sql"`
}

// AnalyzeResponse SQL分析响应
type AnalyzeResponse struct {
	BasicInfo *ResponseBasicInfo `json:"basic_info"` // 基础信息（通用）
	Analyze   interface{}        `json:"analyze"`    // 详细分析（根据类型不同）
}

// ResponseBasicInfo 响应基础信息（通用字段）
type ResponseBasicInfo struct {
	SQLType     string `json:"sql_type"`              // SQL类型
	SQLCategory string `json:"sql_category"`          // SQL分类: DQL/DML/DDL
	RiskLevel   string `json:"risk_level"`            // 风险等级: safe/warning/danger
	RiskReason  string `json:"risk_reason,omitempty"` // 风险原因（低风险时不返回）
}

// DDLDetail DDL详细分析
type DDLDetail struct {
	// CREATE TABLE 专用
	TableInfo   *DDLTableInfo `json:"table_info,omitempty"`   // 表信息
	FieldInfo   *DDLFieldInfo `json:"field_info,omitempty"`   // 字段信息汇总
	IndexDetail []IndexDef    `json:"index_detail,omitempty"` // 索引详情（表格展示）
	FieldDetail []ColumnDef   `json:"field_detail,omitempty"` // 字段详情（表格展示）

	// ALTER TABLE 专用
	AlterInfo   *DDLAlterInfo `json:"alter_info,omitempty"`   // ALTER操作统计
	AlterDetail []AlterAction `json:"alter_detail,omitempty"` // ALTER操作详情（表格展示）

	// DROP/TRUNCATE 等简单操作
	SimpleInfo *DDLSimpleInfo `json:"simple_info,omitempty"` // 简单操作信息
}

// DDLAlterInfo ALTER操作统计信息
type DDLAlterInfo struct {
	TableName    string `json:"table_name"`    // 表名
	TotalActions int    `json:"total_actions"` // 总操作数量

	// 字段操作细分
	AddFieldCount    string `json:"add_field_count"`    // 添加字段 x/总数
	ModifyFieldCount string `json:"modify_field_count"` // 修改字段 x/总数 (包含MODIFY和ALTER COLUMN)
	ChangeFieldCount string `json:"change_field_count"` // 改名字段 x/总数
	DropFieldCount   string `json:"drop_field_count"`   // 删除字段 x/总数

	// 索引操作细分
	AddIndexCount  string `json:"add_index_count"`  // 添加索引 x/总数
	DropIndexCount string `json:"drop_index_count"` // 删除索引 x/总数

	// 约束操作
	ConstraintActions string `json:"constraint_actions"` // 约束操作 x/总数

	// 表属性操作
	TableAttrActions string `json:"table_attr_actions"` // 表属性操作 x/总数

	// 风险操作统计
	AddFieldWithPositionCount   string `json:"add_field_with_position_count"`   // 添加字段时指定位置(FIRST/AFTER)的数量 x/总数
	AddFieldWithoutCommentCount string `json:"add_field_without_comment_count"` // 添加字段时缺少注释的数量 x/总数
	FieldWithCollateCount       string `json:"field_with_collate_count"`        // 字段自定义排序规则的数量 x/总数
}

// DDLTableInfo 表信息
type DDLTableInfo struct {
	TableName  string `json:"table_name"`          // 表名
	FieldCount int    `json:"field_count"`         // 字段数量
	IndexCount int    `json:"index_count"`         // 索引数量
	PrimaryKey string `json:"primary_key"`         // 主键字段（多个用逗号分隔，无则为空）
	Comment    string `json:"comment,omitempty"`   // 表注释
	Engine     string `json:"engine,omitempty"`    // 存储引擎
	Charset    string `json:"charset,omitempty"`   // 字符集
	Collation  string `json:"collation,omitempty"` // 排序规则
}

// DDLFieldInfo 字段信息汇总
type DDLFieldInfo struct {
	CommentSummary  string `json:"comment_summary"`  // 有注释字段 x/y (z%)
	DefaultSummary  string `json:"default_summary"`  // 有默认值字段 x/y
	NullableSummary string `json:"nullable_summary"` // 允许为空字段 x/y
	CollateSummary  string `json:"collate_summary"`  // 自定义排序规则字段 x/y
}

// TableCheck 表规范检查（保留用于内部检查）
type TableCheck struct {
	AllColumnsHaveComment bool       `json:"all_columns_have_comment"`
	AllColumnsHaveDefault bool       `json:"all_columns_have_default"`
	AllColumnsNullable    bool       `json:"all_columns_nullable"`
	ColumnsWithoutComment []string   `json:"columns_without_comment,omitempty"`
	ColumnsWithoutDefault []string   `json:"columns_without_default,omitempty"`
	ColumnsNotNull        []string   `json:"columns_not_null,omitempty"`
	HasCharset            bool       `json:"has_charset"`
	HasCollation          bool       `json:"has_collation"`
	HasTableComment       bool       `json:"has_table_comment"`
	Charset               string     `json:"charset,omitempty"`
	Collation             string     `json:"collation,omitempty"`
	TableComment          string     `json:"table_comment,omitempty"`
	HasIndex              bool       `json:"has_index"`
	HasPrimaryKey         bool       `json:"has_primary_key"`
	IndexCount            int        `json:"index_count"`
	IndexList             []IndexDef `json:"index_list,omitempty"`
}

// ColumnDef 字段定义
type ColumnDef struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default,omitempty"`
	Comment    string `json:"comment,omitempty"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
	AutoIncr   bool   `json:"auto_incr,omitempty"`
	Unique     bool   `json:"unique,omitempty"`
	Position   string `json:"position,omitempty"` // FIRST / AFTER xxx
	Collate    string `json:"collate,omitempty"`  // 字段排序规则
}

// IndexDef 索引定义
type IndexDef struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"` // PRIMARY/UNIQUE/INDEX/FULLTEXT
	Columns []string `json:"columns"`
}

// AlterAction ALTER操作
type AlterAction struct {
	Action        string     `json:"action"`                // ADD/DROP/MODIFY/CHANGE/RENAME
	Target        string     `json:"target"`                // COLUMN/INDEX/KEY/COMMENT
	Column        *ColumnDef `json:"column,omitempty"`      // 字段信息
	OldName       string     `json:"old_name,omitempty"`    // 旧名称
	NewName       string     `json:"new_name,omitempty"`    // 新名称
	Index         *IndexDef  `json:"index,omitempty"`       // 索引信息
	Position      string     `json:"position,omitempty"`    // 位置参数：FIRST / AFTER xxx
	Description   string     `json:"description,omitempty"` // 操作描述
	ChangeSummary string     `json:"change_summary"`        // 变更摘要（简洁展示）
}

// TableOptions 表选项（保留用于内部使用）
type TableOptions struct {
	Engine  string `json:"engine,omitempty"`
	Charset string `json:"charset,omitempty"`
	Collate string `json:"collate,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// DMLDetail DML详细分析
type DMLDetail struct {
	Operation     string      `json:"operation"`                // INSERT/UPDATE/DELETE
	Table         string      `json:"table"`                    // 目标表
	Columns       []string    `json:"columns,omitempty"`        // 涉及字段
	HasWhere      bool        `json:"has_where"`                // 是否有WHERE
	WhereClause   string      `json:"where_clause,omitempty"`   // WHERE条件
	HasLimit      bool        `json:"has_limit,omitempty"`      // 是否有LIMIT
	RowCount      int         `json:"row_count,omitempty"`      // 行数（INSERT）
	DataSource    string      `json:"data_source,omitempty"`    // VALUES/SELECT
	Values        [][]string  `json:"values,omitempty"`         // INSERT的值（预览）
	SetClauses    []SetClause `json:"set_clauses,omitempty"`    // UPDATE的SET子句
	EstimatedRows *int64      `json:"estimated_rows,omitempty"` // 预估影响行数（指针类型，nil表示未查询，0表示查询结果为0）
}

// SetClause UPDATE的SET子句
type SetClause struct {
	Column string `json:"column"`  // 字段名
	Value  string `json:"value"`   // 设置的值
	IsExpr bool   `json:"is_expr"` // 是否是表达式
}

// DQLDetail DQL详细分析
type DQLDetail struct {
	SelectType  string     `json:"select_type"` // SIMPLE/SUBQUERY/UNION/CTE
	Tables      []TableRef `json:"tables"`
	Columns     []string   `json:"columns,omitempty"`
	HasWhere    bool       `json:"has_where"`
	HasJoin     bool       `json:"has_join"`
	HasGroup    bool       `json:"has_group"`
	HasOrder    bool       `json:"has_order"`
	HasLimit    bool       `json:"has_limit"`
	HasSubquery bool       `json:"has_subquery"`
	HasUnion    bool       `json:"has_union"`
	HasCTE      bool       `json:"has_cte"`
}

// TableRef 表引用
type TableRef struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
	Type  string `json:"type,omitempty"` // TABLE/SUBQUERY/CTE
}

// DDLSimpleInfo 简单DDL操作信息
type DDLSimpleInfo struct {
	Operation  string `json:"operation"`           // DROP/TRUNCATE/RENAME
	ObjectType string `json:"object_type"`         // TABLE/INDEX/VIEW/DATABASE
	ObjectName string `json:"object_name"`         // 对象名称
	IfExists   bool   `json:"if_exists,omitempty"` // IF EXISTS
	NewName    string `json:"new_name,omitempty"`  // 新名称（RENAME用）
}
