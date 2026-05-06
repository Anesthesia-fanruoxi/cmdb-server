package task

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// JobListResponse 任务列表响应
type JobListResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TaskKey     string `json:"task_key"`
	Cron        string `json:"cron"`
	CronDesc    string `json:"cron_desc"`
	Description string `json:"description"`
	ScriptID    string `json:"script_id"`
	ScriptType  string `json:"script_type"`
	Status      int    `json:"status"`
	UpdatedAt   string `json:"updated_at"`
}

// JobList 获取任务列表
func JobList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 构建查询
	query := database.DB.Model(&job.JobInfo{})

	// 获取列表数据
	var jobs []job.JobInfo
	if err := query.Order("created_at DESC").Find(&jobs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取任务列表失败")
		return
	}

	// 构建响应数据
	list := make([]JobListResponse, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, JobListResponse{
			ID:          j.ID,
			Name:        j.Name,
			TaskKey:     j.TaskKey,
			Cron:        j.Cron,
			CronDesc:    utils.TranslateCron(j.Cron),
			Description: j.Description,
			ScriptID:    j.ScriptID,
			ScriptType:  j.ScriptType,
			Status:      j.Status,
			UpdatedAt:   j.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.ResponseSuccess(w, "获取成功", list)
}
