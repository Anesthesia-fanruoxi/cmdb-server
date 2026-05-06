package path

// 工具相关路径
const (
	// 搜索相关
	SearchProjectsPath = "/api/elfk/search/projects" // 获取可访问的项目列表
	SearchESPath       = "/api/elfk/search/list"     // ES搜索
	SearchIndexPath    = "/api/elfk/search/indices"  // 获取索引列表和字段

	// 添加新的ES数据查询和分页路由
	SearchESDataPath       = "/api/elfk/search/data"        // ES数据查询（将结果缓存到Redis）
	SearchESPagePath       = "/api/elfk/search/page"        // ES数据分页查询（从Redis缓存获取数据）
	SearchESExportPath     = "/api/elfk/search/export"      // ES数据导出为Excel
	SearchESScrollTestPath = "/api/elfk/search/scroll/test" // ES滚动查询测试
	SearchESContextPath    = "/api/elfk/search/context"     // ES上下文查询
	SearchAnalyzePath      = "/api/elfk/search/analyze"     // ES数据分析

	ViewProjectsPath = "/api/elfk/view/projects" // 获取可访问的项目列表
	ViewListPath     = "/api/elfk/view/list"     // 获取视图列表
	ViewCreatePath   = "/api/elfk/view/create"   // 创建视图
	ViewUpdatePath   = "/api/elfk/view/update"   // 更新视图
	ViewDeletePath   = "/api/elfk/view/delete"   // 删除视图
	ViewDetailPath   = "/api/elfk/view/detail"   // 获取视图详情

	// 共享关键词
	SharedKeywordListPath   = "/api/elfk/keyword/list"   // 共享关键词列表
	SharedKeywordCreatePath = "/api/elfk/keyword/create" // 创建共享关键词
	SharedKeywordUpdatePath = "/api/elfk/keyword/update" // 更新共享关键词
	SharedKeywordDeletePath = "/api/elfk/keyword/delete" // 删除共享关键词
)
