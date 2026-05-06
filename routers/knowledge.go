package routers

import (
	"cmdb/api/knowledge/document"
	"cmdb/api/knowledge/personal"
	"cmdb/api/knowledge/public"
	"cmdb/api/knowledge/recycle"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitKnowledgeRoutes 初始化知识库路由
func InitKnowledgeRoutes(mux *http.ServeMux) {
	// 文档管理
	mux.HandleFunc(path.DocProjectsPath, utils.JWTAuth(document.DocProjects, "knowledge:document:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.DocListPath, utils.JWTAuth(document.DocList, "knowledge:document:r", "", "文档管理"))
	mux.HandleFunc(path.DocCreatePath, utils.JWTAuth(document.DocCreate, "knowledge:document:w", "o", "创建文档"))
	mux.HandleFunc(path.DocUpdatePath, utils.JWTAuth(document.DocUpdate, "knowledge:document:w", "o", "更新文档"))
	mux.HandleFunc(path.DocDeletePath, utils.JWTAuth(document.DocDelete, "knowledge:document:w", "o", "删除文档"))
	mux.HandleFunc(path.DocDetailPath, utils.JWTAuth(document.DocDetail, "knowledge:document:r", "", "文档详情"))
	mux.HandleFunc(path.DocHistoryListPath, utils.JWTAuth(document.DocHistoryList, "knowledge:document:r", "", "文档历史列表"))
	mux.HandleFunc(path.DocHistoryDetailPath, utils.JWTAuth(document.DocHistoryDetail, "knowledge:document:r", "", "文档历史详情"))
	mux.HandleFunc(path.DocHistoryRestorePath, utils.JWTAuth(document.DocHistoryRestore, "knowledge:document:w", "o", "文档历史恢复"))

	// 公共文档
	mux.HandleFunc(path.PublicListPath, utils.JWTAuth(public.PublicList, "knowledge:public:r", "", "公共文档列表"))
	mux.HandleFunc(path.PublicUserListPath, utils.JWTAuth(public.PublicUserList, "knowledge:public:r", "", "用户公共文档列表"))
	mux.HandleFunc(path.PublicDetailPath, utils.JWTAuth(public.PublicDetail, "knowledge:public:r", "", "公共文档详情"))
	mux.HandleFunc(path.PublicCreatePath, utils.JWTAuth(public.PublicCreate, "knowledge:public:w", "o", "创建公共文档"))
	mux.HandleFunc(path.PublicUpdatePath, utils.JWTAuth(public.PublicUpdate, "knowledge:public:w", "o", "更新公共文档"))
	mux.HandleFunc(path.PublicDeletePath, utils.JWTAuth(public.PublicDelete, "knowledge:public:w", "o", "删除公共文档"))
	mux.HandleFunc(path.PublicSharePath, utils.JWTAuth(public.PublicShare, "knowledge:public:r", "o", "分享公共文档"))
	mux.HandleFunc(path.PublicDownloadPath, utils.JWTAuth(public.PublicDownload, "knowledge:public:r", "", "下载公共文档"))
	mux.HandleFunc(path.PublicAnonymousPath, utils.JWTAuth(public.PublicAnonymous, "", "", ""))
	mux.HandleFunc(path.PublicAnonymousDeletePath, utils.JWTAuth(public.PublicAnonymousDelete, "knowledge:public:r", "o", "删除匿名公共文档"))
	mux.HandleFunc(path.PublicHistoryListPath, utils.JWTAuth(public.PublicHistoryList, "knowledge:public:r", "", "公共文档历史列表"))
	mux.HandleFunc(path.PublicHistoryDetailPath, utils.JWTAuth(public.PublicHistoryDetail, "knowledge:public:r", "", "公共文档历史详情"))
	mux.HandleFunc(path.PublicHistoryRestorePath, utils.JWTAuth(public.PublicHistoryRestore, "knowledge:public:w", "o", "恢复公共文档"))
	mux.HandleFunc(path.PublicShareListPath, utils.JWTAuth(public.PublicShareList, "knowledge:public:r", "", "分享列表"))

	// 个人空间
	mux.HandleFunc(path.PersonalListPath, utils.JWTAuth(personal.PersonalList, "knowledge:personal:r", "", "个人文档列表"))
	mux.HandleFunc(path.PersonalDetailPath, utils.JWTAuth(personal.PersonalDetail, "knowledge:personal:r", "", "个人文档详情"))
	mux.HandleFunc(path.PersonalCreatePath, utils.JWTAuth(personal.PersonalCreate, "knowledge:personal:w", "o", "创建个人文档"))
	mux.HandleFunc(path.PersonalUpdatePath, utils.JWTAuth(personal.PersonalUpdate, "knowledge:personal:w", "o", "更新个人文档"))
	mux.HandleFunc(path.PersonalDeletePath, utils.JWTAuth(personal.PersonalDelete, "knowledge:personal:w", "o", "删除个人文档"))

	//// 公共文档回收站管理
	mux.HandleFunc(path.RecycleDocListPath, utils.JWTAuth(recycle.RecycleList, "knowledge:recycle:r", "", "公共文档回收站列表"))
	mux.HandleFunc(path.RecycleDetailPath, utils.JWTAuth(recycle.RecycleDetail, "knowledge:recycle:r", "", "公共文档回收站详情"))
	mux.HandleFunc(path.RecycleDocRefreshPath, utils.JWTAuth(recycle.RecycleRefresh, "knowledge:recycle:w", "o", "刷新公共文档回收站"))
	mux.HandleFunc(path.RecycleDocRemovePath, utils.JWTAuth(recycle.RecycleRemove, "knowledge:recycle:w", "o", "删除公共文档回收站"))
	mux.HandleFunc(path.RecycleDocBatchRefreshPath, utils.JWTAuth(recycle.RecycleBatchRefresh, "knowledge:recycle:w", "o", "批量刷新公共文档回收站"))
	mux.HandleFunc(path.RecycleDocBatchRemovePath, utils.JWTAuth(recycle.RecycleBatchRemove, "knowledge:recycle:w", "o", "批量删除公共文档回收站"))
}
