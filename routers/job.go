package routers

import (
	"cmdb/api/job/agent"
	"cmdb/api/job/exec"
	"cmdb/api/job/scheduler"
	"cmdb/api/job/task"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitJobRoutes 初始化任务相关路由
func InitJobRoutes(mux *http.ServeMux) {
	// 任务管理接口
	mux.HandleFunc(path.JobListPath, utils.JWTAuth(task.JobList, "job:task:r", "", "任务管理"))                     // 任务列表
	mux.HandleFunc(path.JobCreatePath, utils.JWTAuth(task.JobCreate, "job:task:w", "o", "创建任务"))                // 创建任务
	mux.HandleFunc(path.JobUpdatePath, utils.JWTAuth(task.JobUpdate, "job:task:w", "o", "更新任务"))                // 更新任务
	mux.HandleFunc(path.JobDeletePath, utils.JWTAuth(task.JobDelete, "job:task:w", "o", "删除任务"))                // 删除任务
	mux.HandleFunc(path.JobDetailPath, utils.JWTAuth(task.JobDetail, "job:task:r", "", "任务详情"))                 // 任务详情（含脚本内容）
	mux.HandleFunc(path.JobProjectDetailPath, utils.JWTAuth(task.JobProjectDetail, "job:task:r", "", "任务项目详情")) // 任务项目详情（含参数+执行计划）
	mux.HandleFunc(path.JobBindPath, utils.JWTAuth(task.JobBind, "job:task:w", "o", "绑定任务"))                    // 绑定任务（含参数配置）
	mux.HandleFunc(path.JobUnbindPath, utils.JWTAuth(task.JobUnbind, "job:task:w", "o", "解绑任务"))                // 解绑任务
	mux.HandleFunc(path.JobCronPath, utils.JWTAuth(task.JobCron, "job:task:r", "", "验证cron表达式"))                // 验证cron表达式

	// 脚本预览接口
	mux.HandleFunc(path.PreviewScriptPath, utils.JWTAuth(task.PreviewScript, "job:task:r", "", "预览脚本"))                 // 预览脚本（通用）
	mux.HandleFunc(path.PreviewProjectScriptPath, utils.JWTAuth(task.PreviewProjectScript, "job:task:r", "", "预览项目脚本")) // 预览项目脚本

	// 执行记录接口
	mux.HandleFunc(path.JobExecProjectsPath, utils.JWTAuth(exec.ExecProjects, "job:task:r", "", "获取可访问项目列表")) // 获取可访问项目列表
	mux.HandleFunc(path.JobExecListPath, utils.JWTAuth(exec.ExecList, "job:task:r", "", "执行记录列表"))            // 执行记录列表
	mux.HandleFunc(path.JobExecDetailPath, utils.JWTAuth(exec.ExecDetail, "job:task:r", "", "执行记录详情"))        // 执行记录详情

	// 项目管理接口
	mux.HandleFunc(path.ProjectTasksPath, utils.JWTAuth(agent.ProjectTasks, "job:agent:r", "", "项目任务列表"))           // 查询项目关联的任务列表
	mux.HandleFunc(path.ProjectExecutionsPath, utils.JWTAuth(agent.ProjectExecutions, "job:agent:r", "", "项目执行记录")) // 查询项目的执行记录
	mux.HandleFunc(path.AgentReportPath, agent.AgentReportResult)                                                   // Agent上报执行结果（不需要认证）

	// 调度器控制接口
	mux.HandleFunc(path.SchedulerStartPath, utils.JWTAuth(scheduler.StartTask, "job:task:w", "o", "启动任务调度"))       // 启动任务调度
	mux.HandleFunc(path.SchedulerStopPath, utils.JWTAuth(scheduler.StopTask, "job:task:w", "o", "停止任务调度"))         // 停止任务调度
	mux.HandleFunc(path.SchedulerReloadPath, utils.JWTAuth(scheduler.ReloadScheduler, "job:task:w", "o", "重载调度器")) // 重新加载调度器
}
