package exec

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// ExecDetailResponse 执行记录详情响应
type ExecDetailResponse struct {
	ID         int64  `json:"id"`
	JobID      int64  `json:"job_id"`
	TaskKey    string `json:"task_key"`
	Project    string `json:"project"`
	ExecStatus int    `json:"exec_status"`
	Result     string `json:"result"`
	ErrorMsg   string `json:"error_msg"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Duration   int    `json:"duration"`
	CreatedAt  string `json:"created_at"`

	// 关联的任务信息
	JobInfo *struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"job_info"`
}

// ExecDetail 获取执行记录详情
func ExecDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取执行记录ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少执行记录ID")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的执行记录ID")
		return
	}

	// 查询执行记录
	var execution job.JobExecution
	if err := database.DB.Where("id = ?", id).First(&execution).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "执行记录不存在")
		return
	}

	// 查询任务信息
	var jobInfo job.JobInfo
	if err := database.DB.Where("id = ?", execution.JobID).First(&jobInfo).Error; err != nil {
		// 任务可能已被删除，不影响返回执行记录
		jobInfo.Name = "未知任务"
	}

	// 格式化时间
	startTime := execution.StartTime.Format("2006-01-02 15:04:05")
	var endTime string
	if execution.EndTime != nil {
		endTime = execution.EndTime.Format("2006-01-02 15:04:05")
	}

	// 构建响应数据，包含所有详细信息
	response := map[string]interface{}{
		"id":          execution.ID,
		"job_id":      execution.JobID,
		"job_name":    jobInfo.Name,
		"project":     execution.Project,
		"task_key":    execution.TaskKey,
		"task_name":   execution.TaskName,
		"exec_log":    execution.ExecLog, // 详情中包含执行日志
		"exec_status": execution.ExecStatus,
		"result":      execution.Result,   // 详情中包含执行结果
		"error_msg":   execution.ErrorMsg, // 详情中包含错误信息
		"start_time":  startTime,
		"end_time":    endTime,
		"duration":    execution.Duration,
		"created_at":  execution.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
