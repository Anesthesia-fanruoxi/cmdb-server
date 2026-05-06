package exec

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
	"time"
)

// ExecListRequest 执行记录查询参数
type ExecListRequest struct {
	JobID       int64  `json:"job_id"`       // 任务ID
	ProjectName string `json:"project_name"` // 项目名称
	ProjectEnv  string `json:"project_env"`  // 项目环境
	ExecStatus  *int   `json:"exec_status"`  // 执行状态
	Page        int    `json:"page"`         // 页码
	PageSize    int    `json:"page_size"`    // 每页数量
}

// ExecListResponse 执行记录响应
type ExecListResponse struct {
	Total int64          `json:"total"` // 总数
	List  []ExecListItem `json:"list"`  // 列表
}

// ExecListItem 执行记录项
type ExecListItem struct {
	ID          int64  `json:"id"`
	JobID       int64  `json:"job_id"`
	TaskKey     string `json:"task_key"`
	ProjectName string `json:"project_name"`
	ProjectEnv  string `json:"project_env"`
	ExecStatus  int    `json:"exec_status"`
	Result      string `json:"result"`
	ErrorMsg    string `json:"error_msg"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Duration    int    `json:"duration"`
	CreatedAt   string `json:"created_at"`
}

// ExecList 获取执行记录列表
func ExecList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	jobID := r.URL.Query().Get("job_id")
	project := r.URL.Query().Get("project")
	taskName := r.URL.Query().Get("task_name")
	status := r.URL.Query().Get("exec_status")

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件
	query := database.DB.Model(&job.JobExecution{})

	// 应用过滤条件
	if jobID != "" {
		jobIDInt, _ := strconv.ParseInt(jobID, 10, 64)
		query = query.Where("job_id = ?", jobIDInt)
	}
	if project != "" {
		query = query.Where("project = ?", project)
	}
	if taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("exec_status = ?", statusInt)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询执行记录总数失败")
		return
	}

	// 查询数据，按ID倒序排列，最新的记录在前面
	// 只选择列表需要的字段，不包括大文本字段
	var executions []struct {
		ID         int64      `json:"id"`
		JobID      int64      `json:"job_id"`
		Project    string     `json:"project"`
		TaskKey    string     `json:"task_key"`
		TaskName   string     `json:"task_name"`
		ExecStatus int        `json:"exec_status"`
		StartTime  time.Time  `json:"start_time"`
		EndTime    *time.Time `json:"end_time"`
		Duration   int        `json:"duration"`
		CreatedAt  time.Time  `json:"created_at"`
	}

	if err := query.Order("id DESC").
		Select("id, job_id, project, task_key, task_name, exec_status, start_time, end_time, duration, created_at").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&executions).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询执行记录列表失败")
		return
	}

	// 查询任务信息
	var jobIDs []int64
	for _, e := range executions {
		jobIDs = append(jobIDs, e.JobID)
	}

	var jobInfos []job.JobInfo
	if len(jobIDs) > 0 {
		database.DB.Where("id IN ?", jobIDs).Find(&jobInfos)
	}

	// 构建任务ID到名称的映射
	jobMap := make(map[int64]string)
	for _, j := range jobInfos {
		jobMap[j.ID] = j.Name
	}

	// 构建响应数据
	list := make([]map[string]interface{}, 0, len(executions))
	for _, e := range executions {
		jobName := jobMap[e.JobID]
		if jobName == "" {
			jobName = e.TaskName // 如果找不到任务名称，使用记录中的任务名称
		}

		startTime := e.StartTime.Format("2006-01-02 15:04:05")
		var endTime string
		if e.EndTime != nil {
			endTime = e.EndTime.Format("2006-01-02 15:04:05")
		}

		// 构建每条记录的数据，不包含大文本字段
		item := map[string]interface{}{
			"id":          e.ID,
			"job_id":      e.JobID,
			"job_name":    jobName,
			"project":     e.Project,
			"task_key":    e.TaskKey,
			"task_name":   e.TaskName,
			"exec_status": e.ExecStatus,
			"start_time":  startTime,
			"end_time":    endTime,
			"duration":    e.Duration,
			"created_at":  e.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		list = append(list, item)
	}

	// 构建最终响应
	response := map[string]interface{}{
		"total": total,
		"list":  list,
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
