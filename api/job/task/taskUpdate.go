package task

import (
	"cmdb/api/job/scheduler"
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// JobUpdateRequest 更新任务请求
type JobUpdateRequest struct {
	ID            int64  `json:"id"`             // 任务id
	Name          string `json:"name"`           // 任务名称
	Cron          string `json:"cron"`           // cron表达式
	Description   string `json:"description"`    // 任务描述
	ScriptID      string `json:"script_id"`      // 脚本ID
	ScriptType    string `json:"script_type"`    // 脚本类型
	ScriptContent string `json:"script_content"` // 脚本内容
	Status        *int   `json:"status"`         // 状态 0:停止 1:运行
}

// JobUpdate 更新任务
func JobUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req JobUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Logger.Errorf("请求数据解析失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 记录请求参数
	middleware.Logger.Infof("更新任务请求 [ID: %d, Name: %s, ScriptID: %s, ScriptType: %s]", req.ID, req.Name, req.ScriptID, req.ScriptType)

	// 查询任务是否存在
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 更新任务信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
		middleware.Logger.Infof("准备更新 name: %s", req.Name)
	}
	if req.Cron != "" {
		updates["cron"] = req.Cron
		middleware.Logger.Infof("准备更新 cron: %s", req.Cron)
	}
	if req.Description != "" {
		updates["description"] = req.Description
		middleware.Logger.Infof("准备更新 description: %s", req.Description)
	}
	if req.ScriptID != "" {
		updates["script_id"] = req.ScriptID
		middleware.Logger.Infof("准备更新 script_id: %s", req.ScriptID)
	}
	if req.ScriptType != "" {
		updates["script_type"] = req.ScriptType
		middleware.Logger.Infof("准备更新 script_type: %s", req.ScriptType)
	}
	if req.ScriptContent != "" {
		updates["script_content"] = req.ScriptContent
		middleware.Logger.Infof("准备更新 script_content: %d 字符", len(req.ScriptContent))
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			middleware.ResponseError(w, http.StatusBadRequest, "无效的状态值")
			return
		}
		updates["status"] = *req.Status
		middleware.Logger.Infof("准备更新 status: %d", *req.Status)
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.Logger.Warnf("没有要更新的内容 [ID: %d]", req.ID)
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的内容")
		return
	}

	middleware.Logger.Infof("开始更新任务 [ID: %d, 更新字段数: %d]", req.ID, len(updates))

	// 执行更新
	if err := database.DB.Model(&jobInfo).Updates(updates).Error; err != nil {
		middleware.Logger.Errorf("更新任务失败 [ID: %d, Error: %v]", req.ID, err)
		middleware.ResponseError(w, http.StatusInternalServerError, "更新任务失败: "+err.Error())
		return
	}

	middleware.Logger.Infof("数据库更新成功 [ID: %d]", req.ID)

	// 重新查询最新数据
	if err := database.DB.First(&jobInfo, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的数据失败")
		return
	}

	// 同步到调度器
	jobScheduler := scheduler.GetScheduler()
	if err := jobScheduler.UpdateJob(&jobInfo); err != nil {
		middleware.Logger.Errorf("同步任务到调度器失败: %v", err)
		// 不影响更新操作，仅记录日志
	}

	middleware.ResponseSuccess(w, "更新成功", jobInfo)
}
