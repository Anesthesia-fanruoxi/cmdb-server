package task

import (
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"strconv"
)

// ScriptPreviewRequest 脚本预览请求
type ScriptPreviewRequest struct {
	JobID         int64                  `json:"job_id"`                            // 任务ID（二选一）
	ScriptContent string                 `json:"script_content"`                    // 脚本内容（二选一）
	ScriptParams  map[string]interface{} `json:"script_params" validate:"required"` // 脚本参数
}

// ScriptPreviewResponse 脚本预览响应
type ScriptPreviewResponse struct {
	OriginalScript string                 `json:"original_script"` // 原始脚本
	RenderedScript string                 `json:"rendered_script"` // 渲染后的脚本
	ScriptParams   map[string]interface{} `json:"script_params"`   // 使用的参数
}

// PreviewScript 预览脚本（渲染模板）
func PreviewScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req ScriptPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.ScriptParams == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "script_params不能为空")
		return
	}

	var scriptContent string

	// 获取脚本内容
	if req.JobID > 0 {
		// 从数据库查询任务脚本
		var jobInfo job.JobInfo
		if err := database.DB.First(&jobInfo, req.JobID).Error; err != nil {
			middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
			return
		}
		scriptContent = jobInfo.ScriptContent
	} else if req.ScriptContent != "" {
		// 使用传入的脚本内容
		scriptContent = req.ScriptContent
	} else {
		middleware.ResponseError(w, http.StatusBadRequest, "job_id和script_content至少提供一个")
		return
	}

	// 渲染脚本
	renderedScript, err := RenderScript(scriptContent, req.ScriptParams)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "脚本渲染失败: "+err.Error())
		return
	}

	// 构建响应
	response := ScriptPreviewResponse{
		OriginalScript: scriptContent,
		RenderedScript: renderedScript,
		ScriptParams:   req.ScriptParams,
	}

	middleware.ResponseSuccess(w, "预览成功", response)
}

// PreviewProjectScript 预览项目的脚本（使用已配置的参数）
func PreviewProjectScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取参数
	jobIDStr := r.URL.Query().Get("job_id")
	project := r.URL.Query().Get("project")

	if jobIDStr == "" || project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的任务ID")
		return
	}

	// 查询任务信息
	var jobInfo job.JobInfo
	if err := database.DB.First(&jobInfo, jobID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 查询项目绑定信息
	var projectBind job.JobProject
	if err := database.DB.Where("job_id = ? AND project = ?", jobID, project).
		First(&projectBind).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务项目绑定关系不存在")
		return
	}

	// 渲染脚本
	var renderedScript string
	if len(projectBind.ScriptParams) > 0 {
		renderedScript, err = RenderScript(jobInfo.ScriptContent, projectBind.ScriptParams)
		if err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, "脚本渲染失败: "+err.Error())
			return
		}
	} else {
		// 如果没有参数，直接返回原始脚本
		renderedScript = jobInfo.ScriptContent
	}

	// 构建响应
	response := ScriptPreviewResponse{
		OriginalScript: jobInfo.ScriptContent,
		RenderedScript: renderedScript,
		ScriptParams:   projectBind.ScriptParams,
	}

	middleware.ResponseSuccess(w, "预览成功", response)
}
