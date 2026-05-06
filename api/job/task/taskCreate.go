package task

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// JobCreateRequest 创建任务请求
type JobCreateRequest struct {
	Name          string `json:"name"`           // 任务名称
	TaskKey       string `json:"task_key"`       // 任务标识符
	Cron          string `json:"cron"`           // cron表达式
	Description   string `json:"description"`    // 任务描述
	ScriptID      string `json:"script_id"`      // 脚本ID（Agent端执行的标识）
	ScriptType    string `json:"script_type"`    // 脚本类型（仅用于界面展示）
	ScriptContent string `json:"script_content"` // 脚本内容（仅用于界面编辑和展示）
}

// JobCreate 创建任务
func JobCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req JobCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 参数验证
	if req.Name == "" || req.TaskKey == "" || req.Cron == "" || req.ScriptID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要参数(name, task_key, cron, script_id)")
		return
	}

	// 检查任务标识符是否已存在
	var count int64
	if err := database.DB.Model(&job.JobInfo{}).Where("task_key = ?", req.TaskKey).Count(&count).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "任务标识符已存在")
		return
	}

	// 创建任务
	jobInfo := &job.JobInfo{
		Name:          req.Name,
		TaskKey:       req.TaskKey,
		Cron:          req.Cron,
		Description:   req.Description,
		ScriptID:      req.ScriptID,      // 脚本ID（用于执行）
		ScriptType:    req.ScriptType,    // 脚本类型（仅展示）
		ScriptContent: req.ScriptContent, // 脚本内容（仅展示）
		Status:        0,                 // 默认停止状态
		Creator:       userInfo.UserName, // 从请求头获取创建者
	}

	if err := database.DB.Create(jobInfo).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建任务失败")
		return
	}

	middleware.ResponseSuccess(w, "创建成功", jobInfo)
}
