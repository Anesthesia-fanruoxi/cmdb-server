package system

// IPProjectRequest IP项目关联请求（加密）
type IPProjectRequest struct {
	Data string `json:"data"` // 加密数据
}

// IPProjectData IP项目关联数据（解密后）
type IPProjectData struct {
	IP      string `json:"ip"`      // IP地址
	Project string `json:"project"` // 项目标识
}

// IPProjectQueryRequest IP项目查询请求
type IPProjectQueryRequest struct {
	IP      string `json:"ip"`      // 根据IP查询
	Project string `json:"project"` // 根据项目查询
}

// IPProjectQueryResponse IP项目查询响应
type IPProjectQueryResponse struct {
	IP       string   `json:"ip,omitempty"`       // IP地址（查询单个或根据项目返回的IP）
	Project  string   `json:"project,omitempty"`  // 项目标识
	Projects []string `json:"projects,omitempty"` // IP关联的项目列表（一个IP可能被多个项目使用）
}
