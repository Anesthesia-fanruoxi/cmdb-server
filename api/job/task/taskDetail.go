package task

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// JobDetailResponse 任务详情响应
type JobDetailResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	TaskKey       string `json:"task_key"`
	Cron          string `json:"cron"`
	CronDesc      string `json:"cron_desc"` // cron表达式说明
	Description   string `json:"description"`
	ScriptID      string `json:"script_id"`      // 脚本ID
	ScriptType    string `json:"script_type"`    // 脚本类型
	ScriptContent string `json:"script_content"` // 脚本内容
	Status        int    `json:"status"`
	Creator       string `json:"creator"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// JobDetail 获取任务详情
func JobDetail(w http.ResponseWriter, r *http.Request) {
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

	// 查询任务信息
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 构建响应数据
	response := JobDetailResponse{
		ID:            jobInfo.ID,
		Name:          jobInfo.Name,
		TaskKey:       jobInfo.TaskKey,
		Cron:          jobInfo.Cron,
		CronDesc:      utils.TranslateCron(jobInfo.Cron),
		Description:   jobInfo.Description,
		ScriptID:      jobInfo.ScriptID,
		ScriptType:    jobInfo.ScriptType,
		ScriptContent: jobInfo.ScriptContent,
		Status:        jobInfo.Status,
		Creator:       jobInfo.Creator,
		CreatedAt:     jobInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     jobInfo.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
