package types

// ServiceInfo Service信息
type ServiceInfo struct {
	Project     string        `json:"project"`      // 项目代码
	ProjectName string        `json:"project_name"` // 项目中文名称
	Namespace   string        `json:"namespace"`    // 命名空间
	ServiceName string        `json:"service_name"` // 服务名称
	Type        string        `json:"type"`         // 服务类型
	Ports       []PortMapping `json:"ports"`        // 端口映射
}

// PortMapping 端口映射信息
type PortMapping struct {
	Port       int32  `json:"port"`        // 服务端口
	TargetPort int32  `json:"target_port"` // 目标端口
	NodePort   int32  `json:"node_port"`   // 节点端口（NodePort类型服务特有）
	Protocol   string `json:"protocol"`    // 协议
}
