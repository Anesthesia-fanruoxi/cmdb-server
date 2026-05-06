package scheduler

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"

	"github.com/robfig/cron/v3"
)

// StartTaskRequest 启动任务请求
type StartTaskRequest struct {
	JobID int64 `json:"job_id"`
}

// StopTaskRequest 停止任务请求
type StopTaskRequest struct {
	JobID int64 `json:"job_id"`
}

// StartTask 启动任务调度
func StartTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req StartTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 查询任务
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 更新任务状态为启用
	if err := database.DB.Model(&jobInfo).Update("status", 1).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新任务状态失败")
		return
	}

	// 重新查询任务
	if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询任务失败")
		return
	}

	// 添加到调度器
	scheduler := GetScheduler()
	if err := scheduler.AddJob(&jobInfo); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "添加任务到调度器失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "任务已启动", nil)
}

// StopTask 停止任务调度
func StopTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req StopTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 查询任务
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 更新任务状态为停止
	if err := database.DB.Model(&jobInfo).Update("status", 0).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新任务状态失败")
		return
	}

	// 从调度器移除
	scheduler := GetScheduler()
	scheduler.RemoveJob(req.JobID)

	middleware.ResponseSuccess(w, "任务已停止", nil)
}

// ReloadScheduler 重新加载调度器
func ReloadScheduler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	scheduler := GetScheduler()

	// 清空现有任务
	scheduler.mu.Lock()
	for jobID := range scheduler.jobs {
		scheduler.cron.Remove(scheduler.jobs[jobID])
	}
	scheduler.jobs = make(map[int64]cron.EntryID)
	scheduler.mu.Unlock()

	// 重新加载所有任务
	if err := scheduler.LoadJobs(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "重新加载任务失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "调度器已重新加载", nil)
}
