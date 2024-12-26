package routers

import (
	"cmdb/api/system"
	"cmdb/utils"
	"net/http"
)

// InitSystemRoutes 初始化系统管理相关路由
func InitSystemRoutes(mux *http.ServeMux) {
	// 用户相关路由
	mux.HandleFunc(UserLoginPath, system.Login)   // 登录不需要认证
	mux.HandleFunc(UserLogoutPath, system.Logout) // 登出不需要认证
	mux.HandleFunc(UserCreatePath, utils.JWTAuth(system.CreateUser))
	mux.HandleFunc(UserListPath, utils.JWTAuth(system.ListUsers))
	mux.HandleFunc(UserDeletePath, utils.JWTAuth(system.DeleteUser))
	mux.HandleFunc(UserUpdatePath, utils.JWTAuth(system.UpdateUser))
	mux.HandleFunc(UserDetailPath, utils.JWTAuth(system.GetUserDetail))

	// 角色相关路由
	mux.HandleFunc(RoleListPath, utils.JWTAuth(system.ListRoles))
	mux.HandleFunc(RoleCreatePath, utils.JWTAuth(system.CreateRole))
	mux.HandleFunc(RoleDeletePath, utils.JWTAuth(system.DeleteRole))
	mux.HandleFunc(RoleUpdatePath, utils.JWTAuth(system.UpdateRole))

	// 部门相关路由
	mux.HandleFunc(DeptCreatePath, utils.JWTAuth(system.CreateDepartments))
	mux.HandleFunc(DeptUpdatePath, utils.JWTAuth(system.UpdateDepartments))
	mux.HandleFunc(DeptDeletePath, utils.JWTAuth(system.DeleteDepartments))
	mux.HandleFunc(DeptListPath, utils.JWTAuth(system.ListDepartments))

	// 字典相关路由
	mux.HandleFunc(DictCreatePath, utils.JWTAuth(system.CreateDict))
	mux.HandleFunc(DictDeletePath, utils.JWTAuth(system.DeleteDict))
	mux.HandleFunc(DictListPath, utils.JWTAuth(system.ListDicts))
	mux.HandleFunc(DictQueryPath, utils.JWTAuth(system.QueryDict))
	mux.HandleFunc(DictItemCreatePath, utils.JWTAuth(system.CreateDictItem))
	mux.HandleFunc(DictItemDeletePath, utils.JWTAuth(system.DeleteDictItem))

	// 菜单相关路由
	mux.HandleFunc(MenuTreePath, utils.JWTAuth(system.GetMenuTree))
	mux.HandleFunc(MenuUserPath, utils.JWTAuth(system.GetUserMenus))
	mux.HandleFunc(MenuCreatePath, utils.JWTAuth(system.CreateMenu))
	mux.HandleFunc(MenuUpdatePath, utils.JWTAuth(system.UpdateMenu))
	mux.HandleFunc(MenuDeletePath, utils.JWTAuth(system.DeleteMenu))
}
