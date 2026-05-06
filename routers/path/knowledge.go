package path

// 知识库
const (
	// 文档管理
	DocProjectsPath       = "/api/knowledge/doc/projects"        // 获取可访问的项目列表
	DocListPath           = "/api/knowledge/doc/list"            // 文档列表
	DocCreatePath         = "/api/knowledge/doc/create"          // 创建文档
	DocUpdatePath         = "/api/knowledge/doc/update"          // 更新文档
	DocDeletePath         = "/api/knowledge/doc/delete"          // 删除文档
	DocDetailPath         = "/api/knowledge/doc/detail"          // 文档详情
	DocHistoryListPath    = "/api/knowledge/doc/history/list"    // 历史版本列表
	DocHistoryDetailPath  = "/api/knowledge/doc/history/detail"  // 历史版本详情
	DocHistoryRestorePath = "/api/knowledge/doc/history/restore" // 恢复历史版本

	// 公共文档
	PublicListPath            = "/api/knowledge/public/list"             // 公共文档列表
	PublicUserListPath        = "/api/knowledge/public/user-list"        // 公共个人文档列表
	PublicDetailPath          = "/api/knowledge/public/detail"           // 公共文档详情列表
	PublicCreatePath          = "/api/knowledge/public/create"           // 创建公共文档
	PublicUpdatePath          = "/api/knowledge/public/update"           // 更新公共文档
	PublicDeletePath          = "/api/knowledge/public/delete"           // 删除公共文档
	PublicSharePath           = "/api/knowledge/public/share"            // 分享公共文档
	PublicDownloadPath        = "/api/knowledge/public/download"         // 下载公共文档
	PublicAnonymousPath       = "/api/share"                             // 匿名访问
	PublicAnonymousDeletePath = "/api/knowledge/public/anonymous/delete" // 关闭分享
	PublicHistoryListPath     = "/api/knowledge/public/history/list"     // 历史版本列表
	PublicHistoryDetailPath   = "/api/knowledge/public/history/detail"   // 历史版本详情
	PublicHistoryRestorePath  = "/api/knowledge/public/history/restore"  // 恢复历史版本
	PublicShareListPath       = "/api/knowledge/public/share/list"       // 分享列表

	// 个人空间
	PersonalListPath   = "/api/knowledge/personal/list"   // 个人文档列表
	PersonalDetailPath = "/api/knowledge/personal/detail" // 个人文档详情列表
	PersonalCreatePath = "/api/knowledge/personal/create" // 创建个人文档
	PersonalUpdatePath = "/api/knowledge/personal/update" // 更新个人文档
	PersonalDeletePath = "/api/knowledge/personal/delete" // 删除个人文档

	// 回收站管理
	RecycleDocListPath         = "/api/knowledge/recycle/list"          // 回收站文档列表
	RecycleDetailPath          = "/api/knowledge/recycle/detail"        // 回收站文档详情
	RecycleDocRefreshPath      = "/api/knowledge/recycle/refresh"       // 从回收站恢复文档
	RecycleDocRemovePath       = "/api/knowledge/recycle/remove"        // 彻底删除文档
	RecycleDocBatchRefreshPath = "/api/knowledge/recycle/batch-refresh" // 批量恢复文档
	RecycleDocBatchRemovePath  = "/api/knowledge/recycle/batch-remove"  // 批量彻底删除
)
