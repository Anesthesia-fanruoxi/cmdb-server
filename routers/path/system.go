package path

// 系统管理
const (
	// 个人信息
	GetProfilePath    = "/api/user/profile/detail" //获取个人信息
	UpdateProfilePath = "/api/user/profile/update" //修改个人信息
	// 用户管理
	UserListPath                     = "/api/system/user/list"                   // 用户列表
	UserCreatePath                   = "/api/system/user/create"                 // 创建用户
	UserUpdatePath                   = "/api/system/user/update"                 // 更新用户
	UserBatchUpdatePath              = "/api/system/user/batch-update"           // 更新用户
	UserDeletePath                   = "/api/system/user/delete"                 // 删除用户
	UserBatchDeletePath              = "/api/system/user/batch-delete"           // 删除用户
	UserDetailPath                   = "/api/system/user/detail"                 // 用户详情
	UserLoginPath                    = "/api/system/user/login"                  // 用户登录
	UserLogoutPath                   = "/api/system/user/logout"                 // 用户登出
	UserFactorAuthPath               = "/api/system/user/factor/generate"        // 用户创建双因子认证
	UserFactorVerifyPath             = "/api/system/user/factor/verify"          // 用户验证双因子认证
	UserFactorVerifyWithUsernamePath = "/api/system/user/factor/verify-username" // 用户验证双因子认证
	UserMobileLoginPath              = "/api/system/user/mobile/login"           // 新增移动端登录路径
	UserProjectDetailPath            = "/api/system/user/project/detail"         // 用户项目关联详情
	UserProjectUpdatePath            = "/api/system/user/project/update"         // 更新用户项目关联

	// 角色管理
	RoleListPath   = "/api/system/role/list"         // 角色列表
	RoleCreatePath = "/api/system/role/create"       // 创建角色
	RoleUpdatePath = "/api/system/role/update"       // 更新角色
	RoleDeletePath = "/api/system/role/delete"       // 删除角色
	RoleDetailPath = "/api/system/role/detail"       // 角色详情
	RoleMenuDetail = "/api/system/role/menus/detail" // 获取角色菜单权限
	RoleMenuUpdate = "/api/system/role/menus/update" // 更新角色菜单权限

	// 部门管理
	DeptListPath           = "/api/system/dept/list"           // 部门列表
	DeptCreatePath         = "/api/system/dept/create"         // 创建部门
	DeptUpdatePath         = "/api/system/dept/update"         // 更新部门
	DeptDeletePath         = "/api/system/dept/delete"         // 删除部门
	DeptDetailPath         = "/api/system/dept/detail"         // 部门详情
	DeptProjectsDetailPath = "/api/system/dept/project/detail" // 获取部门项目关联
	DeptProjectsUpdatePath = "/api/system/dept/project/update" // 更新部门项目关联

	// 字典管理
	DictListPath   = "/api/system/dict/list"   // 字典列表
	CreateItemPath = "/api/system/dict/create" // 添加键值对
	DeleteItemPath = "/api/system/dict/delete" // 删除键值对
	UpdateItemPath = "/api/system/dict/update" // 更新键值对
	DictDetailPath = "/api/system/dict/detail" // 字典详情

	// 项目管理
	ProjectListPath       = "/api/system/project/list"   // 项目列表
	ProjectDetailListPath = "/api/system/project/detail" // 项目详情
	ProjectCreatePath     = "/api/system/project/create" // 创建项目
	ProjectUpdatePath     = "/api/system/project/update" // 更新项目
	ProjectDeletePath     = "/api/system/project/delete" // 删除项目

	// 菜单管理
	MenuTreePath   = "/api/system/menu/tree"   // 获取菜单树
	MenuUserPath   = "/api/system/menu/user"   // 获取用户菜单
	MenuListPath   = "/api/system/menu/list"   // 获取用户菜单
	MenuCreatePath = "/api/system/menu/create" // 创建菜单
	MenuUpdatePath = "/api/system/menu/update" // 更新菜单
	MenuDeletePath = "/api/system/menu/delete" // 删除菜单

	// IP黑名单管理
	IPBlacklistListPath    = "/api/system/blacklist/list"   // IP黑名单列表
	IPBlacklistAddPath     = "/api/system/blacklist/add"    // 添加IP到黑名单
	IPBlacklistRemovePath  = "/api/system/blacklist/remove" // 从黑名单中移除IP
	IPFailedCountListPath  = "/api/system/blacklist/count"  // 获取IP失败计数列表
	IPFailedCountResetPath = "/api/system/blacklist/reset"  // 重置IP失败计数

	// IP限流管理
	IPRateLimitListPath  = "/api/system/ratelimit/list"  // IP限流列表
	IPRateLimitResetPath = "/api/system/ratelimit/reset" // 重置IP限流

	// 安全配置
	SecurityConfigListPath   = "/api/system/security/list"   // 获取安全配置列表
	SecurityConfigUpdatePath = "/api/system/security/update" // 更新安全配置

	// 系统设置
	BasicSettingPath       = "/api/system/setting/basic" // 获取系统设置
	PasswdSettingPath      = "/api/system/setting/passwd"
	BasicSettingUpdatePath = "/api/system/setting/basic/update" // 更新系统设置

	// IP位置查询
	IPLocationPath = "/api/system/ip/location" // IP位置查询

	// IP项目关联
	IPProjectRecordPath = "/api/system/ip/project/record" // Agent上报IP项目关联（使用Agent Secret加密）
	IPProjectQueryPath  = "/api/system/ip/project/query"  // 查询IP项目关联（使用Agent Secret加密）

	// 设备自动登录
	DeviceBindPath      = "/api/system/user/device/bind"      // 绑定设备
	DeviceUnbindPath    = "/api/system/user/device/unbind"    // 解绑设备
	DeviceChallengePath = "/api/system/user/device/challenge" // 获取挑战码
	DeviceLoginPath     = "/api/system/user/device/login"     // 设备登录

	// 应用更新
	AppUploadPath        = "/api/app/upload"         // 上传安装包
	AppVersionGetPath    = "/api/app/version/list"   // 获取版本信息
	AppVersionUpdatePath = "/api/app/version/update" // 更新版本信息
)
