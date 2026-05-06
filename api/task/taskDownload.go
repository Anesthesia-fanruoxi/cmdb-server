package task

import (
	"cmdb/models/task"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// TaskCenter 类型别名，方便使用
type TaskCenter = task.TaskCenter

// TaskDownload 下载任务结果文件（通过临时下载码）
func TaskDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 从URL参数获取下载码
	downloadCode := r.URL.Query().Get("code")
	if downloadCode == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "下载码不能为空")
		return
	}

	middleware.Logger.Infof("开始下载: code=%s", downloadCode)

	// 从Redis获取taskID
	redisKey := fmt.Sprintf("%s%s", TaskDownloadCodeKeyPrefix, downloadCode)
	middleware.Logger.Infof("查询Redis: key=%s", redisKey)

	taskID, err := database.RDB.Get(r.Context(), redisKey).Result()
	if err != nil {
		middleware.Logger.Errorf("从Redis获取下载码失败: key=%s, error=%v", redisKey, err)
		middleware.ResponseError(w, http.StatusUnauthorized, "下载链接已过期或无效")
		return
	}

	middleware.Logger.Infof("从Redis获取到taskID: %s", taskID)

	// 获取任务记录
	record, err := GetTaskByTaskID(taskID)
	if err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "任务不存在")
		return
	}

	// 检查任务状态
	if record.Status != "success" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务未完成，无法下载")
		return
	}

	// 检查文件路径
	if record.FilePath == "" {
		middleware.ResponseError(w, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(record.FilePath)
	if os.IsNotExist(err) {
		middleware.ResponseError(w, http.StatusNotFound, "文件已过期或被删除")
		return
	}

	// 打开文件
	file, err := os.Open(record.FilePath)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "打开文件失败")
		return
	}
	defer file.Close()

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(record.FileName)))
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 使用流式传输，避免大文件占用内存
	http.ServeContent(w, r, record.FileName, fileInfo.ModTime(), file)

	middleware.Logger.Infof("下载任务文件: taskID=%s, file=%s, size=%d, code=%s", taskID, record.FileName, fileInfo.Size(), downloadCode)
}
