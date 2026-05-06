package assets

import (
	"cmdb/pkg/middleware"
	"time"
)

// ProUpdate 项目发版信息模型
type ProUpdate struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	Project      string    `json:"project" gorm:"type:varchar(64);uniqueIndex;comment:项目名称"`
	Status       string    `json:"status" gorm:"type:varchar(16);default:'success';comment:上次发版状态"`
	TotalUpdates int       `json:"total_updates" gorm:"default:0;comment:总发版次数"`
	GitURL       string    `json:"git_url" gorm:"type:varchar(255);comment:Git仓库地址"`
	Description  string    `json:"description" gorm:"type:text;comment:项目描述"`
	UpdateFeishu string    `json:"update_feishu" gorm:"type:varchar(255);comment:项目通知飞书URL"`
	NotifyFeishu string    `json:"notify_feishu" gorm:"type:varchar(255);comment:步骤通知飞书URL"`
	CreateTime   time.Time `json:"create_time" gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime   time.Time `json:"update_time" gorm:"autoUpdateTime;comment:更新时间"`
}

// TableName 设置表名
func (ProUpdate) TableName() string {
	return middleware.TableProUpdate
}

// ProjectUpdateListResponse 项目发版列表响应结构
type ProjectUpdateListResponse struct {
	Total int                 `json:"total"` // 总数
	Items []ProjectUpdateInfo `json:"items"` // 项目列表
}

// ProjectUpdateListGroupedResponse 按类型分组的项目发版列表响应结构
type ProjectUpdateListGroupedResponse struct {
	Default []ProjectUpdateInfo `json:"default"` // default类型的项目列表
	Web     []ProjectUpdateInfo `json:"web"`     // web类型的项目列表
}

// ProjectUpdateInfo 项目发版信息
type ProjectUpdateInfo struct {
	Project          string `json:"project"`           // 项目标识
	ProjectName      string `json:"project_name"`      // 项目名称
	GitURL           string `json:"git_url"`           // Git仓库地址（根据类型返回前端或后端）
	FeishuURL        string `json:"feishu_url"`        // 飞书通知地址
	BackendTool      string `json:"backend_tool"`      // 后端工具（java17等）
	FrontendTool     string `json:"frontend_tool"`     // 前端工具（node14等）
	EnableSkywalking bool   `json:"enable_skywalking"` // 是否启用skywalking
	LastUpdate       string `json:"last_update"`       // 上次发版时间（按照创建时间降序排列的最新记录）
	Status           string `json:"status"`            // 上次发版状态
	TotalUpdates     int    `json:"total_updates"`     // 总发版次数
}

// CountResult 用于临时存储计数结果
type CountResult struct {
	Name       string `gorm:"column:name"`
	TotalCount int    `gorm:"column:total_count"`
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Project      string `json:"project"`       // 项目标识
	ProjectName  string `json:"project_name"`  // 项目名称
	GitURL       string `json:"git_url"`       // Git仓库地址
	Tool         string `json:"tool"`          // 工具（根据类型返回前端或后端）
	UpdateFeishu string `json:"update_feishu"` // 发版通知地址（开发组内）
	NotifyFeishu string `json:"notify_feishu"` // 通知地址（所有人）
}

// ProUpdateListDetailRequest 请求参数
type ProUpdateListDetailRequest struct {
	Project string `json:"project" form:"project" validate:"required"`
}

// ProUpdateListDetailResponse 响应结构
type ProUpdateListDetailResponse struct {
	ProjectInfo  ProjectInfo        `json:"project_info"`  // 项目基本信息
	TotalUpdates int                `json:"total_updates"` // 总发版次数
	LastUpdate   string             `json:"last_update"`   // 最后更新时间
	Records      []UpdateRecordInfo `json:"records"`       // 所有发版记录
}

// ProUpdateListDetailGroupedResponse 按类型分组的响应结构
type ProUpdateListDetailGroupedResponse struct {
	Default []UpdateRecordInfo `json:"default"` // default类型的发版记录
	Web     []UpdateRecordInfo `json:"web"`     // web类型的发版记录
}

// UpdateRecordInfo 发版记录信息
type UpdateRecordInfo struct {
	ID          uint   `json:"id"`           // 记录ID
	TaskID      string `json:"task_id"`      // 任务ID
	Type        string `json:"type"`         // 发版类型（web/default）
	Status      string `json:"status"`       // 发版状态
	Description string `json:"description"`  // 发版说明
	StartedAt   string `json:"started_at"`   // 开始时间
	CompletedAt string `json:"completed_at"` // 完成时间
}

// RecordRequest 接收的加密请求数据
type RecordRequest struct {
	Data string `json:"data"` // 加密的数据
}

// ReleaseRecordRequest 解密后的发版记录请求数据
type ReleaseRecordRequest struct {
	IsSet bool   `json:"isset"` // 区分任务信息(false)和步骤信息(true)
	ID    string `json:"id"`    // 任务ID

	// 任务信息字段（isset=false时使用）
	Name         string `json:"name"`          // 项目名称
	Description  string `json:"description"`   // 项目描述
	GitURL       string `json:"git_url"`       // Git仓库地址
	UpdateFeishu string `json:"update_feishu"` // 发版通知地址（开发组内）
	NotifyFeishu string `json:"notify_feishu"` // 通知地址（所有人）
	StartedAt    string `json:"started_at"`    // 任务开始时间
	Type         string `json:"type"`          // 发版类型（web/double/single）
	Category     string `json:"category"`      // 分类标签（如manager，仅web类型使用）
	FinishedAt   string `json:"finished_at"`   // 任务结束时间（可为空）
	Status       string `json:"status"`        // 任务状态或步骤状态（共用字段）

	// 步骤信息字段（isset=true时使用）
	Step           int     `json:"step"`             // 步骤编号
	StepType       string  `json:"step_type"`        // 步骤标签
	StepStartedAt  string  `json:"step_started_at"`  // 步骤开始时间
	StepFinishedAt string  `json:"step_finished_at"` // 步骤结束时间
	StepName       string  `json:"step_name"`        // 步骤名称
	StepStatus     string  `json:"step_status"`      // 步骤状态：success/failed
	Remote         string  `json:"remote"`           //来源于server/agent
	Duration       float64 `json:"duration"`         // 步骤耗时（秒，保留2位小数）
	EstimatedEnd   string  `json:"estimated_end"`    // 预计完成时间
	LastDuration   float64 `json:"last_duration"`    // 上次耗时（秒，保留2位小数）
}

// RedisTaskRecord Redis中存储的任务记录结构
type RedisTaskRecord struct {
	ProjectName  string          `json:"project_name"`  // 项目名称
	Description  string          `json:"description"`   // 任务描述
	ErrorMessage string          `json:"error_message"` // 错误信息
	GitURL       string          `json:"git_url"`       // Git仓库地址
	UpdateFeishu string          `json:"update_feishu"` // 发版通知地址（开发组内）
	NotifyFeishu string          `json:"notify_feishu"` // 通知地址（所有人）
	StartedAt    string          `json:"started_at"`    // 开始时间
	FinishedAt   string          `json:"finished_at"`   // 结束时间
	Status       string          `json:"status"`        // 任务状态：pending/running/success/failed
	Step         int             `json:"step"`          // 当前步骤
	StepDetail   []RedisStepInfo `json:"step_detail"`   // 步骤详情数组
	Timestamp    string          `json:"timestamp"`     // 时间戳
	Type         string          `json:"type"`          // 发版类型（web/double/single）
	Category     string          `json:"category"`      // 分类标签（如manager，仅web类型使用）
}

// RedisStepInfo Redis中存储的步骤信息结构
type RedisStepInfo struct {
	Step           int     `json:"step"`             // 步骤编号
	StepType       string  `json:"step_type"`        // 步骤标签
	StepStartedAt  string  `json:"step_started_at"`  // 步骤开始时间
	StepFinishedAt string  `json:"step_finished_at"` // 步骤结束时间
	StepName       string  `json:"step_name"`        // 步骤名称
	StepStatus     string  `json:"step_status"`      // 步骤状态：success/failed
	Duration       float64 `json:"duration"`         // 步骤耗时（秒，保留2位小数）
	Remote         string  `json:"remote"`           // 请求来源：server/agent
	EstimatedEnd   string  `json:"estimated_end"`    // 预计完成时间
	LastDuration   float64 `json:"last_duration"`    // 上次耗时（秒，保留2位小数）
}

// ProUpdateTaskDetailRequest 请求参数
type ProUpdateTaskDetailRequest struct {
	ID uint `json:"id" form:"id" validate:"required"`
}

// TaskDetailSSEResponse SSE返回的任务详情（包含完整步骤列表）
type TaskDetailSSEResponse struct {
	ProjectName  string         `json:"project_name"`  // 项目名称
	Description  string         `json:"description"`   // 任务描述
	ErrorMessage string         `json:"error_message"` // 错误信息
	GitURL       string         `json:"git_url"`       // Git仓库地址
	UpdateFeishu string         `json:"update_feishu"` // 发版通知地址
	NotifyFeishu string         `json:"notify_feishu"` // 通知地址
	StartedAt    string         `json:"started_at"`    // 开始时间
	FinishedAt   string         `json:"finished_at"`   // 结束时间
	Status       string         `json:"status"`        // 任务状态
	CurrentStep  int            `json:"current_step"`  // 当前步骤
	TotalSteps   int            `json:"total_steps"`   // 总步骤数
	Type         string         `json:"type"`          // 发版类型
	Category     string         `json:"category"`      // 分类标签
	Steps        []FullStepInfo `json:"steps"`         // 完整步骤列表
}

// FullStepInfo 完整步骤信息（模板+实际状态）
type FullStepInfo struct {
	Step           int     `json:"step"`             // 步骤编号
	StepType       string  `json:"step_type"`        // 步骤标签
	StepName       string  `json:"step_name"`        // 步骤名称
	StepStatus     string  `json:"step_status"`      // 步骤状态：pending/running/success/failed
	StepStartedAt  string  `json:"step_started_at"`  // 步骤开始时间
	StepFinishedAt string  `json:"step_finished_at"` // 步骤结束时间
	Duration       float64 `json:"duration"`         // 步骤耗时
	Remote         string  `json:"remote"`           // 执行位置：server/agent
	EstimatedEnd   string  `json:"estimated_end"`    // 预计完成时间
	LastDuration   float64 `json:"last_duration"`    // 上次耗时
}

// MobileTaskResponse 手机端精简版响应（不包含敏感信息）
type MobileTaskResponse struct {
	ProjectName  string           `json:"project_name"`  // 项目名称
	ErrorMessage string           `json:"error_message"` // 错误信息
	StartedAt    string           `json:"started_at"`    // 开始时间
	FinishedAt   string           `json:"finished_at"`   // 结束时间
	Status       string           `json:"status"`        // 任务状态
	CurrentStep  int              `json:"current_step"`  // 当前步骤
	TotalSteps   int              `json:"total_steps"`   // 总步骤数
	Type         string           `json:"type"`          // 发版类型
	Steps        []MobileStepInfo `json:"steps"`         // 步骤列表
}

// MobileStepInfo 手机端精简版步骤信息
type MobileStepInfo struct {
	Step           int     `json:"step"`             // 步骤编号
	StepName       string  `json:"step_name"`        // 步骤名称
	StepStatus     string  `json:"step_status"`      // 步骤状态
	StepStartedAt  string  `json:"step_started_at"`  // 步骤开始时间
	StepFinishedAt string  `json:"step_finished_at"` // 步骤结束时间
	Duration       float64 `json:"duration"`         // 步骤耗时
}
