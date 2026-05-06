package routers

import (
	"cmdb/api/system/appupdate"
	"cmdb/api/system/dept"
	"cmdb/api/system/dict"
	"cmdb/api/system/ipLocation"
	"cmdb/api/system/menu"
	"cmdb/api/system/project"
	"cmdb/api/system/role"
	"cmdb/api/system/security"
	"cmdb/api/system/setting"
	"cmdb/api/system/user"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitSystemRoutes 初始化系统路由
func InitSystemRoutes(mux *http.ServeMux) {
	// 用户管理
	mux.HandleFunc(path.UserLoginPath, utils.JWTAuth(user.UserLogin, "", "", "登录"))                                            // 登录不需要权限
	mux.HandleFunc(path.UserLogoutPath, utils.JWTAuth(user.UserLogout, "", "", "登出"))                                          // 登出不需要权限
	mux.HandleFunc(path.GetProfilePath, utils.JWTAuth(user.GetProfile, "", "", "获取个人信息"))                                      // 获取个人信息不需要权限
	mux.HandleFunc(path.UpdateProfilePath, utils.JWTAuth(user.UpdateProfile, "", "", "修改个人信息"))                                // 修改个人信息不需要权限
	mux.HandleFunc(path.UserFactorAuthPath, utils.JWTAuth(user.GenerateFactorAuth, "", "", "生成双因子认证"))                         // 生成双因子认证不需要权限
	mux.HandleFunc(path.UserFactorVerifyPath, utils.JWTAuth(user.VerifyFactorAuth, "", "", "验证双因子认证"))                         // 验证双因子认证不需要权限
	mux.HandleFunc(path.UserFactorVerifyWithUsernamePath, utils.JWTAuth(user.VerifyFactorAuthWithUsername, "", "", "验证双因子认证")) // 验证双因子认证不需要权限
	mux.HandleFunc(path.UserMobileLoginPath, user.UserMobileLogin)                                                             // 移动端登录接口
	//mux.HandleFunc(path.UserMobileInfoPath, utils.MobileJWTAuth(user.GetMobileUserInfo, "", "", "获取移动端用户信息"))                // 移动端用户信息接口

	mux.HandleFunc(path.UserCreatePath, utils.JWTAuth(user.UserCreate, "system:user:w", "o", "创建用户"))
	mux.HandleFunc(path.UserUpdatePath, utils.JWTAuth(user.UserUpdate, "system:user:w", "o", "更新用户"))
	mux.HandleFunc(path.UserBatchUpdatePath, utils.JWTAuth(user.UserBatchUpdate, "system:user:w", "o", "批量更新用户"))
	mux.HandleFunc(path.UserDeletePath, utils.JWTAuth(user.UserDelete, "system:user:w", "o", "删除用户"))
	mux.HandleFunc(path.UserBatchDeletePath, utils.JWTAuth(user.UserBatchDelete, "system:user:w", "o", "批量删除用户"))
	mux.HandleFunc(path.UserListPath, utils.JWTAuth(user.UserList, "system:user:r", "", "用户列表"))
	mux.HandleFunc(path.UserDetailPath, utils.JWTAuth(user.UserDetail, "system:user:r", "", "用户详情"))
	mux.HandleFunc(path.UserProjectDetailPath, utils.JWTAuth(user.UserProjectDetail, "system:user:r", "", "用户项目详情"))
	mux.HandleFunc(path.UserProjectUpdatePath, utils.JWTAuth(user.UserProjectUpdate, "system:user:w", "o", "更新用户项目"))

	// 角色管理
	mux.HandleFunc(path.RoleListPath, utils.JWTAuth(role.RoleList, "system:role:r", "", "角色列表"))
	mux.HandleFunc(path.RoleCreatePath, utils.JWTAuth(role.RoleCreate, "system:role:w", "o", "创建角色"))
	mux.HandleFunc(path.RoleUpdatePath, utils.JWTAuth(role.RoleUpdate, "system:role:w", "o", "更新角色"))
	mux.HandleFunc(path.RoleDeletePath, utils.JWTAuth(role.RoleDelete, "system:role:w", "o", "删除角色"))
	mux.HandleFunc(path.RoleDetailPath, utils.JWTAuth(role.RoleDetail, "system:role:r", "", "角色详情"))
	// 角色菜单权限
	mux.HandleFunc(path.RoleMenuDetail, utils.JWTAuth(role.RoleMenuDetail, "", "", "角色菜单权限详情"))
	mux.HandleFunc(path.RoleMenuUpdate, utils.JWTAuth(role.RoleMenuUpdate, "", "o", "更新角色菜单权限"))

	// 部门管理
	mux.HandleFunc(path.DeptListPath, utils.JWTAuth(dept.DeptList, "system:dept:r", "", "部门列表"))
	mux.HandleFunc(path.DeptCreatePath, utils.JWTAuth(dept.DeptCreate, "system:dept:w", "o", "创建部门"))
	mux.HandleFunc(path.DeptUpdatePath, utils.JWTAuth(dept.DeptUpdate, "system:dept:w", "o", "更新部门"))
	mux.HandleFunc(path.DeptDeletePath, utils.JWTAuth(dept.DeptDelete, "system:dept:w", "o", "删除部门"))
	mux.HandleFunc(path.DeptDetailPath, utils.JWTAuth(dept.DeptDetail, "system:dept:r", "", "部门详情"))
	mux.HandleFunc(path.DeptProjectsDetailPath, utils.JWTAuth(dept.DeptProjectDetail, "system:dept:r", "", "部门项目详情"))
	mux.HandleFunc(path.DeptProjectsUpdatePath, utils.JWTAuth(dept.DeptProjectUpdate, "system:dept:w", "o", "更新部门项目"))

	// 菜单管理
	mux.HandleFunc(path.MenuTreePath, utils.JWTAuth(menu.MenuTree, "system:menu:r", "", "菜单树"))
	mux.HandleFunc(path.MenuListPath, utils.JWTAuth(menu.MenuList, "system:menu:r", "", "菜单列表"))
	mux.HandleFunc(path.MenuCreatePath, utils.JWTAuth(menu.MenuCreate, "system:menu:w", "o", "创建菜单"))
	mux.HandleFunc(path.MenuUpdatePath, utils.JWTAuth(menu.MenuUpdate, "system:menu:w", "o", "更新菜单"))
	mux.HandleFunc(path.MenuDeletePath, utils.JWTAuth(menu.MenuDelete, "system:menu:w", "o", "删除菜单"))
	mux.HandleFunc(path.MenuUserPath, utils.JWTAuth(menu.MenuUser, "", "", "获取用户菜单权限")) // 获取用户菜单权限不需要权限

	// 字典管理
	mux.HandleFunc(path.CreateItemPath, utils.JWTAuth(dict.CreateDictItem, "system:dict:w", "o", "创建字典项"))
	mux.HandleFunc(path.DeleteItemPath, utils.JWTAuth(dict.DeleteDictItem, "system:dict:w", "o", "删除字典项"))
	mux.HandleFunc(path.UpdateItemPath, utils.JWTAuth(dict.UpdateDictItem, "system:dict:w", "o", "更新字典项"))
	mux.HandleFunc(path.DictListPath, utils.JWTAuth(dict.DictList, "system:dict:r", "", "字典列表"))
	mux.HandleFunc(path.DictDetailPath, utils.JWTAuth(dict.DictDetail, "system:dict:r", "", "字典详情"))

	// 项目管理
	mux.HandleFunc(path.ProjectListPath, utils.JWTAuth(project.ProjectList, "system:project:r", "", "项目列表"))
	mux.HandleFunc(path.ProjectDetailListPath, utils.JWTAuth(project.ProjectDetailList, "system:project:r", "", "项目详情"))
	mux.HandleFunc(path.ProjectCreatePath, utils.JWTAuth(project.ProjectCreate, "system:project:w", "o", "创建项目"))
	mux.HandleFunc(path.ProjectUpdatePath, utils.JWTAuth(project.ProjectUpdate, "system:project:w", "o", "更新项目"))
	mux.HandleFunc(path.ProjectDeletePath, utils.JWTAuth(project.ProjectDelete, "system:project:w", "o", "删除项目"))

	// IP黑名单管理
	mux.HandleFunc(path.IPBlacklistListPath, utils.JWTAuth(user.GetIPBlacklist, "system:security:r", "", "查看IP黑名单"))
	mux.HandleFunc(path.IPBlacklistAddPath, utils.JWTAuth(user.AddIPToBlacklist, "system:security:w", "o", "添加IP到黑名单"))
	mux.HandleFunc(path.IPBlacklistRemovePath, utils.JWTAuth(user.RemoveIPFromBlacklist, "system:security:w", "o", "从黑名单移除IP"))
	mux.HandleFunc(path.IPFailedCountListPath, utils.JWTAuth(user.GetIPFailedCounts, "system:security:r", "", "查看IP失败计数"))
	mux.HandleFunc(path.IPFailedCountResetPath, utils.JWTAuth(user.ResetIPFailedCount, "system:security:w", "o", "重置IP失败计数"))

	// IP限流管理
	mux.HandleFunc(path.IPRateLimitListPath, utils.JWTAuth(user.GetIPRateLimits, "system:security:r", "", "查看IP限流列表"))
	mux.HandleFunc(path.IPRateLimitResetPath, utils.JWTAuth(user.ResetIPRateLimit, "system:security:w", "o", "重置IP限流"))

	// 安全配置
	mux.HandleFunc(path.SecurityConfigListPath, utils.JWTAuth(security.SecurityGetConfig, "system:security:r", "", "获取安全配置"))
	mux.HandleFunc(path.SecurityConfigUpdatePath, utils.JWTAuth(security.SecurityUpdateConfig, "system:security:w", "o", "更新安全配置"))

	// 系统设置
	mux.HandleFunc(path.BasicSettingPath, utils.JWTAuth(setting.GetBasicSetting, "system:setting:r", "", "获取基础设置"))
	mux.HandleFunc(path.PasswdSettingPath, utils.JWTAuth(setting.GetPasswdSetting, "", "", "获取密码规则")) // 获取密码规则不需要权限
	mux.HandleFunc(path.BasicSettingUpdatePath, utils.JWTAuth(setting.UpdateBasicSetting, "system:setting:w", "o", "更新基础设置"))

	// IP位置查询
	mux.HandleFunc(path.IPLocationPath, ipLocation.GetLocation) // IP位置查询无需认证

	// IP项目关联（无需认证，使用Agent Secret加密）
	mux.HandleFunc(path.IPProjectRecordPath, ipLocation.RecordIPProject) // Agent上报
	mux.HandleFunc(path.IPProjectQueryPath, ipLocation.QueryIPProject)   // 查询

	// 设备自动登录
	mux.HandleFunc(path.DeviceBindPath, utils.JWTAuth(user.DeviceBind, "", "", "绑定设备"))            // 绑定设备需要登录
	mux.HandleFunc(path.DeviceUnbindPath, utils.JWTAuth(user.DeviceUnbind, "", "", "解绑设备"))        // 解绑设备需要登录
	mux.HandleFunc(path.DeviceChallengePath, utils.JWTAuth(user.DeviceChallenge, "", "", "获取挑战码")) // 白名单，无需token
	mux.HandleFunc(path.DeviceLoginPath, utils.JWTAuth(user.DeviceLogin, "", "", "设备登录"))          // 白名单，无需token

	// 应用更新
	mux.HandleFunc(path.AppUploadPath, utils.JWTAuth(appupdate.Upload, "", "", "上传安装包"))                // 白名单，无需token
	mux.HandleFunc(path.AppVersionGetPath, utils.JWTAuth(appupdate.GetVersion, "", "", "获取版本信息"))       // 白名单，无需token
	mux.HandleFunc(path.AppVersionUpdatePath, utils.JWTAuth(appupdate.UpdateVersion, "", "", "更新版本信息")) // 白名单，无需token
}
