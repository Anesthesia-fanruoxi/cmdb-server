package task

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// JobProjectDetailResponse 任务项目详情响应
type JobProjectDetailResponse struct {
	// 任务基本信息
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TaskKey     string `json:"task_key"`
	Cron        string `json:"cron"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	Creator     string `json:"creator"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`

	// 关联的项目信息
	Projects []ProjectInfo `json:"projects"`
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Project      string                 `json:"project"`                 // 项目标识
	NextRun      string                 `json:"next_run"`                // 下次运行时间
	ScriptParams map[string]interface{} `json:"script_params,omitempty"` // 脚本参数
}

// JobProjectDetail 获取任务和项目绑定详情
func JobProjectDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取任务ID
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的任务ID")
		return
	}

	// 查询任务基本信息
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 查询关联的项目信息
	var projects []job.JobProject
	if err := database.DB.Where("job_id = ?", id).Find(&projects).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取项目信息失败")
		return
	}

	// 当前时间，用于计算下一次执行时间
	now := time.Now()

	// 创建cron解析器
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// 解析cron表达式
	var nextRunTime time.Time
	if jobInfo.Cron != "" {
		schedule, err := parser.Parse(jobInfo.Cron)
		if err == nil {
			// 计算下一次执行时间
			nextRunTime = schedule.Next(now)
		}
	}

	// 构建响应数据
	projectInfos := make([]ProjectInfo, 0, len(projects))
	for _, p := range projects {
		projectInfo := ProjectInfo{
			Project:      p.Project,
			ScriptParams: p.ScriptParams,
		}

		// 添加下一次执行时间
		if !nextRunTime.IsZero() {
			projectInfo.NextRun = nextRunTime.Format("2006-01-02 15:04:05")
		}

		projectInfos = append(projectInfos, projectInfo)
	}

	response := JobProjectDetailResponse{
		ID:          jobInfo.ID,
		Name:        jobInfo.Name,
		TaskKey:     jobInfo.TaskKey,
		Cron:        jobInfo.Cron,
		Description: jobInfo.Description,
		Status:      jobInfo.Status,
		Creator:     jobInfo.Creator,
		CreatedAt:   jobInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   jobInfo.UpdatedAt.Format("2006-01-02 15:04:05"),
		Projects:    projectInfos,
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
