package task

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// PodScaleTask 创建单个Pod缩放任务
func PodScaleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var requestBody struct {
		Namespace string `json:"namespace"`
		Action    bool   `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	if requestBody.Namespace == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "命名空间不能为空")
		return
	}

	params, err := json.Marshal(requestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化任务参数失败: "+err.Error())
		return
	}

	task, err := GlobalTaskManager.CreateTask("pod_scale", params, userInfo.UserName, userInfo.UserID)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建任务失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "已创建Pod缩放任务", map[string]interface{}{
		"task_id": task.ID,
		"status":  task.Status,
	})
}

// PodBatchScaleTask 创建批量Pod缩放任务
func PodBatchScaleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var requestBody struct {
		Namespace []string `json:"namespace"`
		Action    bool     `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	if len(requestBody.Namespace) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "命名空间列表不能为空")
		return
	}

	params, err := json.Marshal(requestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化任务参数失败: "+err.Error())
		return
	}

	task, err := GlobalTaskManager.CreateTask("pod_batch_scale", params, userInfo.UserName, userInfo.UserID)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建任务失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "已创建批量Pod缩放任务", map[string]interface{}{
		"task_id": task.ID,
		"status":  task.Status,
	})
}
