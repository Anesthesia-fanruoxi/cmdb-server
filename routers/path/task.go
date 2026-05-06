package path

const (
	// 任务管理API路径
	TaskListPath             = "/sse/tasks/list"              // 任务列表（SSE）
	TaskDetailPath           = "/sse/tasks/detail"            // 任务详情（SSE）
	TaskCancelPath           = "/api/tasks/cancel"            // 取消任务
	TaskPreviewPath          = "/api/tasks/preview"           // 数据预览
	TaskExportPath           = "/api/tasks/export"            // 导出任务
	TaskDownloadPath         = "/api/tasks/download"          // 下载任务文件
	TaskGenerateDownloadPath = "/api/tasks/download/generate" // 生成临时下载链接

	// Pod操作相关API路径
	TaskK8sScalePodPath       = "/api/tasks/pod/scale"       // 操作namespace下pod
	TaskK8sBatchScalePodsPath = "/api/tasks/pod/batch-scale" // 批量操作namespace下pod
)
