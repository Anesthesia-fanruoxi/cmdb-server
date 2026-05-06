package agent

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// ProjectExecutionInfo 项目执行记录信息
type ProjectExecutionInfo struct {
	ID         int64  `json:"id"`
	JobID      int64  `json:"job_id"`
	TaskName   string `json:"task_name"`
	TaskKey    string `json:"task_key"`
	ExecStatus int    `json:"exec_status"`
	Result     string `json:"result"`
	ErrorMsg   string `json:"error_msg"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Duration   int    `json:"duration"`
	CreatedAt  string `json:"created_at"`
}

// ProjectExecutions 获取项目的任务执行记录
func ProjectExecutions(w http.ResponseWriter, r *http.Request) {
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

	// 获取分页参数
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	var total int64
	if err := database.DB.Model(&job.JobExecution{}).
		Where("project = ?", project).
		Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询总数失败")
		return
	}

	// 查询执行记录
	var executions []job.JobExecution
	if err := database.DB.Where("project = ?", project).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&executions).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取执行记录失败")
		return
	}

	// 构建返回数据
	executionList := make([]ProjectExecutionInfo, 0, len(executions))
	for _, exec := range executions {
		execInfo := ProjectExecutionInfo{
			ID:         exec.ID,
			JobID:      exec.JobID,
			TaskName:   exec.TaskName,
			TaskKey:    exec.TaskKey,
			ExecStatus: exec.ExecStatus,
			Result:     exec.Result,
			ErrorMsg:   exec.ErrorMsg,
			StartTime:  exec.StartTime.Format("2006-01-02 15:04:05"),
			Duration:   exec.Duration,
			CreatedAt:  exec.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// EndTime可能为null
		if exec.EndTime != nil {
			execInfo.EndTime = exec.EndTime.Format("2006-01-02 15:04:05")
		}

		executionList = append(executionList, execInfo)
	}

	// 返回分页数据
	response := map[string]interface{}{
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"executions": executionList,
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
