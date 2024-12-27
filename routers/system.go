package routers

import (
	"cmdb/api/system"
	"cmdb/utils"
	"net/http"
)

// InitSystemRoutes 初始化系统管理相关路由
func InitSystemRoutes(mux *http.ServeMux) {
	// 用户相关路由
	mux.HandleFunc(UserLoginPath, system.Login)                             // 登录不需要认证
	mux.HandleFunc(UserLogoutPath, system.Logout)                           // 登出不需要认证
	mux.HandleFunc(UserDetailPath, utils.JWTAuth(system.GetUserDetail, "")) // 获取用户信息只需要token
	mux.HandleFunc(MenuUserPath, utils.JWTAuth(system.GetUserMenus, ""))    // 获取用户菜单只需要token

	// 其他需要权限的路由
	mux.HandleFunc(UserCreatePath, utils.JWTAuth(system.CreateUser, "system:user"))
	mux.HandleFunc(UserListPath, utils.JWTAuth(system.ListUsers, "system:user"))
	mux.HandleFunc(UserDeletePath, utils.JWTAuth(system.DeleteUser, "system:user"))
	mux.HandleFunc(UserUpdatePath, utils.JWTAuth(system.UpdateUser, "system:user"))

	// 角色相关路由
	mux.HandleFunc(RoleListPath, utils.JWTAuth(system.ListRoles, "system:role"))
	mux.HandleFunc(RoleCreatePath, utils.JWTAuth(system.CreateRole, "system:role"))
	mux.HandleFunc(RoleDeletePath, utils.JWTAuth(system.DeleteRole, "system:role"))
	mux.HandleFunc(RoleUpdatePath, utils.JWTAuth(system.UpdateRole, "system:role"))
	mux.HandleFunc(RoleGetMenuPath, utils.JWTAuth(system.GetRoleMenus, "system:role"))
	mux.HandleFunc(RoleSetMenuPath, utils.JWTAuth(system.UpdateRoleMenus, "system:role"))

	// 部门相关路由
	mux.HandleFunc(DeptCreatePath, utils.JWTAuth(system.CreateDepartments, "system:dept"))
	mux.HandleFunc(DeptUpdatePath, utils.JWTAuth(system.UpdateDepartments, "system:dept"))
	mux.HandleFunc(DeptDeletePath, utils.JWTAuth(system.DeleteDepartments, "system:dept"))
	mux.HandleFunc(DeptListPath, utils.JWTAuth(system.ListDepartments, "system:dept"))
	mux.HandleFunc(GetDeptProjectsPath, utils.JWTAuth(system.GetDeptProjects, ""))
	mux.HandleFunc(UpdateDeptProjectsPath, utils.JWTAuth(system.UpdateDeptProjects, ""))

	// 字典相关路由
	mux.HandleFunc(DictCreatePath, utils.JWTAuth(system.CreateDict, "system:dict"))
	mux.HandleFunc(DictDeletePath, utils.JWTAuth(system.DeleteDict, "system:dict"))
	mux.HandleFunc(DictListPath, utils.JWTAuth(system.ListDicts, "system:dict"))
	mux.HandleFunc(DictQueryPath, utils.JWTAuth(system.QueryDict, ""))
	mux.HandleFunc(DictItemCreatePath, utils.JWTAuth(system.CreateDictItem, "system:dict"))
	mux.HandleFunc(DictItemDeletePath, utils.JWTAuth(system.DeleteDictItem, "system:dict"))

	// 菜单相关路由
	mux.HandleFunc(MenuTreePath, utils.JWTAuth(system.GetMenuTree, "system:menu"))
	mux.HandleFunc(MenuCreatePath, utils.JWTAuth(system.CreateMenu, "system:menu"))
	mux.HandleFunc(MenuUpdatePath, utils.JWTAuth(system.UpdateMenu, "system:menu"))
	mux.HandleFunc(MenuDeletePath, utils.JWTAuth(system.DeleteMenu, "system:menu"))

}
