package models

import "time"

// ScaleRequest 扩缩容请求
type ScaleRequest struct {
	Namespace string `json:"namespace"` // 命名空间
	Action    string `json:"action"`    // 操作类型：scale_up 或 scale_down
}

// NamespaceStatusResponse 表示命名空间的状态响应
type NamespaceStatusResponse struct {
	Project            string  `json:"project"`             // 项目 key
	ProjectName        string  `json:"project_name"`        // 项目名称
	Namespace          string  `json:"namespace"`           // 命名空间名称
	HasPods            bool    `json:"has_pods"`            // 是否有 Pods
	SubDomain          string  `json:"sub_domain"`          // 二级域名
	ActionUserName     *string `json:"action_user_name"`    // 操作用户名称
	ActionTime         *string `json:"action_time"`         // 操作时间
	ActionTimestamp    *string `json:"action_timestamp"`    // 操作时间戳
	OperatUserName     *string `json:"operat_user_name"`    // 操作者用户名
	OperationTime      *string `json:"operation_time"`      // 操作时间
	OperationTimestamp *string `json:"operation_timestamp"` // 操作时间戳
}

// BatchScaleRequest 批量扩缩容请求
type BatchScaleRequest struct {
	Namespaces []string `json:"namespaces"` // 命名空间列表
	Action     string   `json:"action"`     // 操作类型：scale_up 或 scale_down
}

// BatchScaleResponse 批量扩缩容响应
type BatchScaleResponse struct {
	Success []string `json:"success"` // 成功的命名空间列表
	Failed  []struct {
		Namespace string `json:"namespace"` // 失败命名空间
		Error     string `json:"error"`     // 失败原因
	} `json:"failed"`
	TotalCount   int    `json:"total_count"`   // 总数
	SuccessCount int    `json:"success_count"` // 成功数量
	FailedCount  int    `json:"failed_count"`  // 失败数量
	TimeElapsed  string `json:"time_elapsed"`  // 耗时
}

// BatchScaleConfig 批量扩缩容配置
type BatchScaleConfig struct {
	MaxConcurrent  int           // 最大并发数
	IntervalDelay  time.Duration // 操作间隔
	RequestTimeout time.Duration // 单个请求超时时间
}

// SafetyChecks 安全检查配置
type SafetyChecks struct {
	MaxErrorRate    float64       // 最大允许错误率
	MaxResponseTime time.Duration // 最大响应时间
	MinHealthyNodes int           // 最小健康节点数
}

// ScaleResult 缩容结果
type ScaleResult struct {
	Namespace string // 命名空间
	Error     error  // 错误信息
}
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	NodePort   int32  `json:"node_port"`
	Protocol   string `json:"protocol"`
	TargetPort int32  `json:"target_port"`
}

type ServicePortResponse struct {
	Project     string        `json:"project"`
	ProjectName string        `json:"project_name"`
	Namespace   string        `json:"namespace"`
	ServiceName string        `json:"service_name"`
	Ports       []ServicePort `json:"ports"`
}
