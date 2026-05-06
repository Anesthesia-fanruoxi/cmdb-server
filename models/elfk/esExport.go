package elfk

// ScrollRequest 滚动查询请求结构体
type ScrollRequest struct {
	IndexPattern   string                 `json:"index_pattern"`            // 索引名称
	QueryId        string                 `json:"query_id"`                 // 查询ID
	Project        string                 `json:"project"`                  // 项目标识
	ViewName       string                 `json:"view_name"`                // 视图名称
	StartTime      string                 `json:"start_time"`               // 开始时间
	EndTime        string                 `json:"end_time"`                 // 结束时间
	TimeField      string                 `json:"time_field"`               // 时间字段
	TimeFormat     string                 `json:"time_format"`              // 时间格式
	Keyword        string                 `json:"keyword"`                  // 搜索关键词
	Size           int                    `json:"size"`                     // 批次大小，默认1000
	Query          map[string]interface{} `json:"query,omitempty"`          // 自定义查询（可选）
	UseFieldFilter bool                   `json:"use_field_filter"`         // 是否启用字段过滤
	IncludeFields  []string               `json:"include_fields,omitempty"` // 指定要包含的字段
	ExcludeFields  []string               `json:"exclude_fields,omitempty"` // 指定要排除的字段
}
