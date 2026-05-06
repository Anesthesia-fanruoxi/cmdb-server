package sql

// SearchRequest 查询请求结构体
type SearchRequest struct {
	Agent    string `json:"agent"`    // 项目标识符，用于查询代理地址
	Query    string `json:"query"`    // 完整的查询语句
	DBName   string `json:"dbName"`   // 数据库名称
	Page     int    `json:"page"`     // 页码，从1开始
	QueryID  string `json:"query_id"` // 查询ID，用于取消查询（可选，不传则自动生成）
	Platform string `json:"platform"` // 平台标识：desktop(React桌面端) 或 web(Vue Web端)
}

// SearchResult 查询结果结构体
type SearchResult struct {
	DBName  string         `json:"db_name"`
	Results []*QueryResult `json:"results"`
	Took    int64          `json:"took"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`  // 当前页码
	Pages   int            `json:"pages"` // 总页数
}

// QueryResult 单个查询结果
type QueryResult struct {
	QueryID string          `json:"query_id"` // 查询唯一标识符
	Columns []string        `json:"columns"`
	DBName  string          `json:"db_name"`
	Rows    [][]interface{} `json:"rows"` // 使用interface{}数组存储多种类型的值
	Took    int64           `json:"took"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`  // 当前页码
	Pages   int             `json:"pages"` // 总页数
}

// EncryptedResponse 加密响应结构体
type EncryptedResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Encrypted string `json:"encrypted"`
	} `json:"data"`
}

// SearchDBRequest 查询数据库请求结构体
type SearchDBRequest struct {
	Agent string `json:"agent"` // 项目标识符，用于查询代理地址
}

// SearchTBRequest 查询表请求结构体
type SearchTBRequest struct {
	Agent  string `json:"agent"`  // 项目标识符，用于查询代理地址
	DBName string `json:"dbName"` // 数据库名称
}

// PageRequest 分页请求结构体
type PageRequest struct {
	QueryID string `json:"query_id"` // 查询ID
	Page    int    `json:"page"`     // 页码
}

// SqlHistory SQL历史结构体
type SqlHistory struct {
	CreatedAt string `json:"created_at"`
	QuerySql  string `json:"query_sql"`
}
