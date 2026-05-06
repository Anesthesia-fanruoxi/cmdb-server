package agent

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
)

// ProjectTaskInfo 项目任务信息
type ProjectTaskInfo struct {
	JobID        int64                  `json:"job_id"`
	Name         string                 `json:"name"`
	TaskKey      string                 `json:"task_key"`
	Cron         string                 `json:"cron"`
	Description  string                 `json:"description"`
	ScriptID     string                 `json:"script_id"`
	ScriptType   string                 `json:"script_type"`
	ScriptParams map[string]interface{} `json:"script_params,omitempty"` // 脚本参数
	Status       int                    `json:"status"`
	NextRunTime  string                 `json:"next_run_time,omitempty"`
	CreatedAt    string                 `json:"created_at"`
}

// ProjectTasks 获取项目关联的任务列表
func ProjectTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取项目标识
	project := r.URL.Query().Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少项目标识")
		return
	}

	// 验证项目权限
	accessibleProjects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 检查用户是否有权限访问该项目
	if !isSuperAdmin {
		hasPermission := false
		for _, p := range accessibleProjects {
			if p == project {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
			return
		}
	}

	// 查询该项目关联的任务绑定关系
	var projectBinds []job.JobProject
	if err := database.DB.Where("project = ?", project).Find(&projectBinds).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务绑定失败")
		return
	}

	// 如果没有关联任务，返回空列表
	if len(projectBinds) == 0 {
		middleware.ResponseSuccess(w, "获取成功", []ProjectTaskInfo{})
		return
	}

	// 提取任务ID列表
	jobIDs := make([]int64, 0, len(projectBinds))
	paramsMap := make(map[int64]map[string]interface{}) // job_id -> script_params
	for _, bind := range projectBinds {
		jobIDs = append(jobIDs, bind.JobID)
		paramsMap[bind.JobID] = bind.ScriptParams
	}

	// 查询任务信息
	var jobInfos []job.JobInfo
	if err := database.DB.Where("id IN ?", jobIDs).Find(&jobInfos).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务信息失败")
		return
	}

	// 当前时间，用于计算下一次执行时间
	now := time.Now()

	// 创建cron解析器
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// 构建任务列表
	taskList := make([]ProjectTaskInfo, 0, len(jobInfos))
	for _, jobInfo := range jobInfos {
		taskInfo := ProjectTaskInfo{
			JobID:        jobInfo.ID,
			Name:         jobInfo.Name,
			TaskKey:      jobInfo.TaskKey,
			Cron:         jobInfo.Cron,
			Description:  jobInfo.Description,
			ScriptID:     jobInfo.ScriptID,
			ScriptType:   jobInfo.ScriptType,
			ScriptParams: paramsMap[jobInfo.ID], // 从映射中获取脚本参数
			Status:       jobInfo.Status,
			CreatedAt:    jobInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 如果有cron表达式，计算下一次执行时间
		if jobInfo.Cron != "" {
			schedule, err := parser.Parse(jobInfo.Cron)
			if err == nil {
				nextTime := schedule.Next(now)
				taskInfo.NextRunTime = nextTime.Format("2006-01-02 15:04:05")
			}
		}

		taskList = append(taskList, taskInfo)
	}

	middleware.ResponseSuccess(w, "获取成功", taskList)
}
