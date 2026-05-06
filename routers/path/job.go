package path

// 任务管理路径
const (
	// 任务基础操作
	JobListPath          = "/api/job/task/list"           // 任务列表
	JobCreatePath        = "/api/job/task/create"         // 创建任务
	JobUpdatePath        = "/api/job/task/update"         // 更新任务
	JobDeletePath        = "/api/job/task/delete"         // 删除任务
	JobDetailPath        = "/api/job/task/detail"         // 任务详情（包含脚本内容）
	JobProjectDetailPath = "/api/job/task/project/detail" // 任务项目详情（含参数+执行计划）
	JobBindPath          = "/api/job/task/bind"           // 绑定任务（含参数配置）
	JobUnbindPath        = "/api/job/task/unbind"         // 解绑任务

	// 脚本预览
	PreviewScriptPath        = "/api/job/task/script/preview"         // 预览脚本（通用）
	PreviewProjectScriptPath = "/api/job/task/project/script/preview" // 预览项目脚本

	// 执行记录
	JobExecProjectsPath = "/api/job/exec/projects" // 获取可访问的项目列表
	JobExecListPath     = "/api/job/exec/list"     // 执行记录列表
	JobExecDetailPath   = "/api/job/exec/detail"   // 执行记录详情

	// 项目管理接口
	ProjectTasksPath      = "/api/job/agent/project/tasks"      // 查询项目关联的任务列表
	ProjectExecutionsPath = "/api/job/agent/project/executions" // 查询项目的执行记录
	AgentReportPath       = "/api/job/agent/report"             // Agent上报执行结果（保留）

	// 调度器控制接口
	SchedulerStartPath  = "/api/job/scheduler/start"  // 启动任务调度
	SchedulerStopPath   = "/api/job/scheduler/stop"   // 停止任务调度
	SchedulerReloadPath = "/api/job/scheduler/reload" // 重新加载调度器

	// Cron验证
	JobCronPath = "/api/job/task/cron" // 验证cron表达式
)
