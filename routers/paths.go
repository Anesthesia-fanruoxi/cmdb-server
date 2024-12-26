package routers

// 资产管理 - 测试环境
const (

	// 集群管理
	ClusterTestStatusPath     = "/api/asset/test/cluster/status"      // 集群状态
	ClusterTestScalePath      = "/api/asset/test/cluster/scale"       // 扩缩容
	ClusterTestServicePath    = "/api/asset/test/cluster/service"     // 服务列表
	ClusterTestNamespacePath  = "/api/asset/test/cluster/namespace"   // 命名空间列表
	ClusterTestBatchScalePath = "/api/asset/test/cluster/scale/batch" // 批量扩缩容

	// 迭代管理
	IterationStartPath = "/api/test/iteration/start" // 开始迭代l
)

// 资产管理 - 生产环境
const (
	// 服务器管理
	ServerProListPath   = "/api/pro/server/list"   // 服务器列表
	ServerProCreatePath = "/api/pro/server/create" // 创建服务器
	ServerProUpdatePath = "/api/pro/server/update" // 更新服务器
	ServerProDeletePath = "/api/pro/server/delete" // 删除服务器
	ServerProDetailPath = "/api/pro/server/detail" // 服务器详情

	// 集群管理
	ClusterProStatusPath     = "/api/pro/cluster/status"      // 集群状态
	ClusterProScalePath      = "/api/pro/cluster/scale"       // 扩缩容
	ClusterProServicePath    = "/api/pro/cluster/service"     // 服务列表
	ClusterProNamespacePath  = "/api/pro/cluster/namespace"   // 命名空间列表
	ClusterProBatchScalePath = "/api/pro/cluster/scale/batch" // 批量扩缩容
)

// 系统管理
const (
	// 用户管理
	UserListPath   = "/api/system/user/list"   // 用户列表
	UserCreatePath = "/api/system/user/create" // 创建用户
	UserUpdatePath = "/api/system/user/update" // 更新用户
	UserDeletePath = "/api/system/user/delete" // 删除用户
	UserDetailPath = "/api/system/user/detail" // 用户详情
	UserLoginPath  = "/api/system/user/login"  // 用户登录
	UserLogoutPath = "/api/system/user/logout" // 用户登录

	// 角色管理
	RoleListPath   = "/api/system/role/list"   // 角色列表
	RoleCreatePath = "/api/system/role/create" // 创建角色
	RoleUpdatePath = "/api/system/role/update" // 更新角色
	RoleDeletePath = "/api/system/role/delete" // 删除角色
	RoleDetailPath = "/api/system/role/detail" // 角色详情

	// 部门管理
	DeptListPath   = "/api/system/dept/list"   // 部门列表
	DeptCreatePath = "/api/system/dept/create" // 创建部门
	DeptUpdatePath = "/api/system/dept/update" // 更新部门
	DeptDeletePath = "/api/system/dept/delete" // 删除部门
	DeptDetailPath = "/api/system/dept/detail" // 部门详情

	// 字典管理
	DictListPath   = "/api/system/dict/list"   // 字典列表
	DictCreatePath = "/api/system/dict/create" // 创建字典
	DictDeletePath = "/api/system/dict/delete" // 删除字典
	DictQueryPath  = "/api/system/dict/query"  // 字典详情

	// 删除键值
	DictItemCreatePath = "/api/system/dict/item/create" // 创建键值
	DictItemDeletePath = "/api/system/dict/item/delete" // 删除键值

	// 菜单管理
	MenuTreePath   = "/api/system/menu/tree"   // 获取菜单树
	MenuUserPath   = "/api/system/menu/user"   // 获取用户菜单
	MenuCreatePath = "/api/system/menu/create" // 创建菜单
	MenuUpdatePath = "/api/system/menu/update" // 更新菜单
	MenuDeletePath = "/api/system/menu/delete" // 删除菜单
)

// 监控中心
const (
	// 服务监控
	ServiceMonitorPath = "/api/monitor/service" // 服务监控
	// 主机监控
	HostMonitorPath = "/api/monitor/host" // 主机监控
	// 容器监控
	ContainerMonitorPath = "/api/monitor/container" // 容器监控
)

// 知识库
const (
	// 文档管理
	DocListPath   = "/api/knowledge/doc/list"   // 文档列表
	DocCreatePath = "/api/knowledge/doc/create" // 创建文档
	DocUpdatePath = "/api/knowledge/doc/update" // 更新文档
	DocDeletePath = "/api/knowledge/doc/delete" // 删除文档
	DocDetailPath = "/api/knowledge/doc/detail" // 文档详情
)
