package routers

import (
	"cmdb/api/task"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitTaskRoutes 初始化任务管理相关路由
func InitTaskRoutes(mux *http.ServeMux) {
	mux.HandleFunc(path.TaskListPath, utils.JWTAuth(task.TaskList, "", "", "任务列表"))
	mux.HandleFunc(path.TaskDetailPath, utils.JWTAuth(task.TaskDetail, "", "", "任务详情"))
	mux.HandleFunc(path.TaskPreviewPath, utils.JWTAuth(task.TaskPreview, "", "", "数据预览"))
	mux.HandleFunc(path.TaskCancelPath, utils.JWTAuth(task.TaskCancel, "", "", "取消任务"))
	mux.HandleFunc(path.TaskExportPath, utils.JWTAuth(task.TaskExport, "", "", "导出任务"))
	mux.HandleFunc(path.TaskGenerateDownloadPath, utils.JWTAuth(task.TaskGenerateDownload, "", "", "生成临时下载链接"))
	mux.HandleFunc(path.TaskDownloadPath, utils.JWTAuth(task.TaskDownload, "", "", "下载任务文件")) // 走JWT但在白名单中

	// Pod相关API
	mux.HandleFunc(path.TaskK8sScalePodPath, utils.JWTAuth(task.PodScaleTask, "assets:kubePod:w", "o", "停止Pod"))
	mux.HandleFunc(path.TaskK8sBatchScalePodsPath, utils.JWTAuth(task.PodBatchScaleTask, "assets:kubePod:w", "o", "批量停止Pod"))
}
