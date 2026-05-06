package task

import (
	"cmdb/api/job/scheduler"
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// JobBindRequest 任务绑定请求
type JobBindRequest struct {
	JobID        int64                  `json:"job_id"`                  // 任务ID
	Project      string                 `json:"project"`                 // 项目标识
	ScriptParams map[string]interface{} `json:"script_params,omitempty"` // 脚本参数（可选）
}

// JobBind 绑定任务
func JobBind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req JobBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Logger.Errorf("请求数据解析失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败: "+err.Error())
		return
	}

	// 参数验证
	if req.JobID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "job_id不能为空")
		return
	}
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project不能为空")
		return
	}

	// 记录请求日志
	middleware.Logger.Infof("绑定任务请求 [JobID: %d, Project: %s]", req.JobID, req.Project)

	// 检查任务是否存在
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 查询是否已存在绑定
	var existingBind job.JobProject
	err := database.DB.Where("job_id = ? AND project = ?", req.JobID, req.Project).First(&existingBind).Error

	if err == nil {
		// 已存在，更新配置
		existingBind.ScriptParams = req.ScriptParams
		if err := database.DB.Save(&existingBind).Error; err != nil {
			middleware.Logger.Errorf("更新项目绑定失败 [JobID: %d, Project: %s, Error: %v]", req.JobID, req.Project, err)
			middleware.ResponseError(w, http.StatusInternalServerError, "更新项目绑定失败: "+err.Error())
			return
		}
		middleware.Logger.Infof("更新项目绑定 [JobID: %d, Project: %s, 参数数量: %d]", req.JobID, req.Project, len(req.ScriptParams))
	} else {
		// 不存在，创建新绑定
		projectBind := &job.JobProject{
			JobID:        req.JobID,
			Project:      req.Project,
			ScriptParams: req.ScriptParams,
		}
		if err := database.DB.Create(projectBind).Error; err != nil {
			middleware.Logger.Errorf("创建项目绑定失败 [JobID: %d, Project: %s, Error: %v]", req.JobID, req.Project, err)
			middleware.ResponseError(w, http.StatusInternalServerError, "创建项目绑定失败: "+err.Error())
			return
		}
		middleware.Logger.Infof("创建项目绑定 [JobID: %d, Project: %s, 参数数量: %d]", req.JobID, req.Project, len(req.ScriptParams))
	}

	middleware.Logger.Infof("任务绑定成功 [JobID: %d, Project: %s]", req.JobID, req.Project)

	// 同步到调度器 - 项目绑定变化需要重新加载任务
	jobScheduler := scheduler.GetScheduler()
	if err := jobScheduler.UpdateJob(&jobInfo); err != nil {
		middleware.Logger.Warnf("同步任务到调度器失败: %v", err)
		// 不影响绑定操作，仅记录日志
	}

	middleware.ResponseSuccess(w, "绑定成功", nil)
}

// JobUnbind 解绑项目
func JobUnbind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req JobBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Logger.Errorf("请求数据解析失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败: "+err.Error())
		return
	}

	// 参数验证
	if req.JobID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "job_id不能为空")
		return
	}
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project不能为空")
		return
	}

	// 记录请求日志
	middleware.Logger.Infof("解绑任务请求 [JobID: %d, Project: %s]", req.JobID, req.Project)

	// 检查任务是否存在
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 检查绑定是否存在
	var existingBind job.JobProject
	if err := database.DB.Where("job_id = ? AND project = ?", req.JobID, req.Project).First(&existingBind).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "项目绑定不存在")
		return
	}

	// 删除绑定
	if err := database.DB.Where("job_id = ? AND project = ?", req.JobID, req.Project).Delete(&job.JobProject{}).Error; err != nil {
		middleware.Logger.Errorf("删除项目绑定失败 [JobID: %d, Project: %s, Error: %v]", req.JobID, req.Project, err)
		middleware.ResponseError(w, http.StatusInternalServerError, "删除项目绑定失败: "+err.Error())
		return
	}

	middleware.Logger.Infof("解绑成功 [JobID: %d, Project: %s]", req.JobID, req.Project)

	// 同步到调度器
	jobScheduler := scheduler.GetScheduler()
	if err := jobScheduler.UpdateJob(&jobInfo); err != nil {
		middleware.Logger.Warnf("同步任务到调度器失败: %v", err)
	}

	middleware.ResponseSuccess(w, "解绑成功", nil)
}
