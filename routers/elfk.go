package routers

import (
	"cmdb/api/elfk/search"
	"cmdb/api/elfk/view"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitToolRoutes 初始化工具相关路由
func InitEsRoutes(mux *http.ServeMux) {
	// ES搜索相关路由
	mux.HandleFunc(path.SearchProjectsPath, utils.JWTAuth(search.SearchProjects, "elfk:search:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.SearchESPath, utils.JWTAuth(search.ESSearchDataList, "elfk:search:r", "", "es搜索"))
	mux.HandleFunc(path.SearchIndexPath, utils.JWTAuth(search.EsInfo, "elfk:search:r", "", "es索引"))

	// 添加新的ES缓存查询路由
	mux.HandleFunc(path.SearchESDataPath, utils.JWTAuth(search.ESSearchDataList, "elfk:search:r", "", "es数据查询"))
	mux.HandleFunc(path.SearchESPagePath, utils.JWTAuth(search.ESSearchDataPage, "elfk:search:r", "", "es分页查询"))
	mux.HandleFunc(path.SearchAnalyzePath, utils.JWTAuth(search.ESSearchAnalysis, "elfk:search:r", "", "es数据分析"))

	// 添加新的滚动查询测试接口
	mux.HandleFunc(path.SearchESScrollTestPath, utils.JWTAuth(search.EsScrollData, "elfk:search:r", "", "es滚动查询测试"))
	// 添加新的ES数据导出接口
	mux.HandleFunc(path.SearchESExportPath, utils.JWTAuth(search.EsExportDate, "elfk:search:w", "", "es数据导出"))
	// 添加上下文查询接口
	mux.HandleFunc(path.SearchESContextPath, utils.JWTAuth(search.EsSearchContext, "elfk:search:r", "", "es上下文查询"))

	// ES视图相关路由
	mux.HandleFunc(path.ViewProjectsPath, utils.JWTAuth(view.ViewProjects, "elfk:view:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.ViewListPath, utils.JWTAuth(view.ListView, "elfk:view:r", "", "es视图列表"))
	mux.HandleFunc(path.ViewCreatePath, utils.JWTAuth(view.CreateView, "elfk:view:w", "o", "创建es视图"))
	mux.HandleFunc(path.ViewUpdatePath, utils.JWTAuth(view.UpdateView, "elfk:view:w", "o", "更新es视图"))
	mux.HandleFunc(path.ViewDeletePath, utils.JWTAuth(view.DeleteView, "elfk:view:w", "o", "删除es视图"))
	mux.HandleFunc(path.ViewDetailPath, utils.JWTAuth(view.ViewDetail, "elfk:view:r", "", "es视图详情"))

	// 共享关键词相关路由
	mux.HandleFunc(path.SharedKeywordListPath, utils.JWTAuth(search.SharedKeywordList, "elfk:search:r", "", "共享关键词列表"))
	mux.HandleFunc(path.SharedKeywordCreatePath, utils.JWTAuth(search.SharedKeywordCreate, "elfk:search:r", "o", "创建共享关键词"))
	mux.HandleFunc(path.SharedKeywordUpdatePath, utils.JWTAuth(search.SharedKeywordUpdate, "elfk:search:r", "o", "更新共享关键词"))
	mux.HandleFunc(path.SharedKeywordDeletePath, utils.JWTAuth(search.SharedKeywordDelete, "elfk:search:r", "o", "删除共享关键词"))
}
