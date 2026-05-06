package sql

// RuleCheckConfig 规则检查配置（从rule_content JSON解析）
type RuleCheckConfig struct {
	// 通用配置
	MinValue    *int     `json:"min_value,omitempty"`    // 最小值
	MaxValue    *int     `json:"max_value,omitempty"`    // 最大值
	Required    *bool    `json:"required,omitempty"`     // 是否必需
	Forbidden   *bool    `json:"forbidden,omitempty"`    // 是否禁止
	AllowedList []string `json:"allowed_list,omitempty"` // 允许列表
	ForbidList  []string `json:"forbid_list,omitempty"`  // 禁止列表

	// 百分比配置
	MinPercent *int `json:"min_percent,omitempty"` // 最小百分比（如字段注释覆盖率≥80%）

	// 字符串匹配
	Pattern string `json:"pattern,omitempty"` // 正则表达式模式
}

// RuleCheckResult 规则检查结果
type RuleCheckResult struct {
	RuleID           int64             `json:"rule_id"`                     // 规则ID
	RuleName         string            `json:"rule_name"`                   // 规则名称
	RuleType         string            `json:"rule_type"`                   // 规则类型
	RuleCategory     string            `json:"rule_category"`               // 规则分类：security/standard
	Passed           bool              `json:"passed"`                      // 是否通过
	ErrorMessage     string            `json:"error_message,omitempty"`     // 错误信息
	RuleDescription  string            `json:"rule_description,omitempty"`  // 规则描述
	Details          []string          `json:"details,omitempty"`           // 详细信息（简单文本列表）
	ViolationDetails []ViolationDetail `json:"violation_details,omitempty"` // 违规明细（结构化）
}

// ViolationDetail 违规明细
type ViolationDetail struct {
	Target        string `json:"target"`                   // 违规目标：字段名/索引名/表名等
	TargetType    string `json:"target_type,omitempty"`    // 目标类型：column/index/table
	Issue         string `json:"issue"`                    // 问题描述
	CurrentValue  string `json:"current_value,omitempty"`  // 当前值
	ExpectedValue string `json:"expected_value,omitempty"` // 期望值
}

// SQLRuleCheckResult SQL规则检查结果（单条SQL）
type SQLRuleCheckResult struct {
	SQL          string             `json:"sql"`                    // 原始SQL
	SQLType      string             `json:"sql_type"`               // SQL类型
	SQLCategory  string             `json:"sql_category"`           // SQL分类
	TableNames   []string           `json:"table_names,omitempty"`  // 涉及的表名
	RuleResults  []*RuleCheckResult `json:"rule_results"`           // 规则检查结果
	HasViolation bool               `json:"has_violation"`          // 是否有违规
	HasBlocker   bool               `json:"has_blocker"`            // 是否有阻断级别违规
	ParseResult  *AnalyzeResponse   `json:"parse_result,omitempty"` // 完整解析结果（可选）
}

// BatchRuleCheckResult 批量规则检查结果
type BatchRuleCheckResult struct {
	SQLResults   []*SQLRuleCheckResult `json:"sql_results"`   // 每条SQL的检查结果
	HasViolation bool                  `json:"has_violation"` // 是否有违规
	HasBlocker   bool                  `json:"has_blocker"`   // 是否有阻断级别违规
	Summary      *CheckSummary         `json:"summary"`       // 汇总信息
}

// CheckSummary 检查汇总
type CheckSummary struct {
	TotalSQL      int `json:"total_sql"`      // 总SQL数
	PassedSQL     int `json:"passed_sql"`     // 通过的SQL数
	ViolationSQL  int `json:"violation_sql"`  // 违规的SQL数
	BlockerSQL    int `json:"blocker_sql"`    // 阻断的SQL数
	TotalRules    int `json:"total_rules"`    // 总规则数
	ViolatedRules int `json:"violated_rules"` // 违反的规则数
	BlockerRules  int `json:"blocker_rules"`  // 阻断级别规则数
}
