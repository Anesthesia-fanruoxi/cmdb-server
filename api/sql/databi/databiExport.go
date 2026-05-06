package databi

import (
	"cmdb/api/task"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ExportRequest 导出请求结构
type ExportRequest struct {
	Project string `json:"project"` // 项目名称
	Context string `json:"context"` // SQL 语句
	Type    string `json:"type"`    // 固定为 "sql"
}

// DataBIExport 创建 DataBI 异步导出任务（统一走任务中心）
// POST /api/sql/databi/export
func DataBIExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析请求
	var req ExportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求失败")
		return
	}

	// 验证参数
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if req.Context == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "context参数不能为空")
		return
	}

	// 默认 type 为 sql
	if req.Type == "" {
		req.Type = "sql"
	}

	// 验证项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !isSuperAdmin && !utils.Contains(projects, req.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 构建任务参数
	taskParams, err := json.Marshal(map[string]interface{}{
		"project": req.Project,
		"context": req.Context,
		"type":    req.Type,
	})
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化参数失败")
		return
	}

	// 创建异步导出任务
	exportTask, err := task.GlobalTaskManager.CreateTask(
		"databi_export",
		json.RawMessage(taskParams),
		userInfo.UserName,
		userInfo.UserID,
	)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("创建导出任务失败: %v", err))
		return
	}

	middleware.ResponseSuccess(w, "导出任务已创建", map[string]interface{}{
		"task_id": exportTask.ID,
		"message": "导出任务已创建，请在任务中心查看进度",
	})
}
