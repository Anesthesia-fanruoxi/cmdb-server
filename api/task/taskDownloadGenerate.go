package task

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TaskGenerateDownload 生成临时下载链接
func TaskGenerateDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req struct {
		TaskID string `json:"taskId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败")
		return
	}
	defer r.Body.Close()

	if req.TaskID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 判断是否为管理员
	isAdmin := userInfo.RoleID == 1

	// 验证任务权限
	var record *TaskCenter
	var err error

	if isAdmin {
		record, err = GetTaskByTaskID(req.TaskID)
	} else {
		record, err = GetTaskByTaskIDAndUser(req.TaskID, userInfo.UserID)
	}

	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 检查任务状态
	if record.Status != "success" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务未完成，无法生成下载链接")
		return
	}

	// 检查文件路径
	if record.FilePath == "" {
		middleware.ResponseError(w, http.StatusNotFound, "文件不存在")
		return
	}

	// 生成随机下载码
	downloadCode := uuid.New().String()

	// 存储到Redis（只存储taskID）
	redisKey := fmt.Sprintf("%s%s", TaskDownloadCodeKeyPrefix, downloadCode)
	expireTime := 600 * time.Second // 600秒过期

	middleware.Logger.Infof("准备存储下载码到Redis: key=%s, taskID=%s", redisKey, req.TaskID)

	if err := database.RDB.Set(r.Context(), redisKey, req.TaskID, expireTime).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成下载链接失败")
		middleware.Logger.Errorf("存储下载码到Redis失败: %v", err)
		return
	}

	middleware.Logger.Infof("成功存储下载码到Redis: key=%s", redisKey)

	// 从数据库获取平台地址
	var setting system.Setting
	downloadURL := "/api/tasks/download?code=" + downloadCode
	if err := database.DB.First(&setting).Error; err == nil && setting.SystemUrl != "" {
		downloadURL = setting.SystemUrl + "/api/tasks/download?code=" + downloadCode
	}

	expireAt := time.Now().Add(expireTime)

	middleware.ResponseSuccess(w, "生成下载链接成功", map[string]interface{}{
		"downloadUrl":  downloadURL,
		"downloadCode": downloadCode,
		"expireIn":     60,
		"expireAt":     expireAt.Format("2006-01-02 15:04:05"),
	})

	middleware.Logger.Infof("生成临时下载链接: taskID=%s, code=%s, user=%s", req.TaskID, downloadCode, userInfo.UserName)
}
