package utils

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"fmt"
	"net/http"
	"strings"
)

// ProjectAuth 项目权限验证结构体
type ProjectAuthRequst struct {
	UserID       uint
	DeptID       uint
	RoleID       uint
	RoleLevel    int64
	OnlineAssets bool   // 生产资产是否忽略权限（1=忽略，0=不忽略）
	TestAssets   bool   // 测试资产是否忽略权限（1=忽略，0=不忽略）
	MenuPath     string // 当前菜单路径（用于过滤用户单独绑定的项目）
}

// CheckProjectPermission 检查项目权限
// 参数env: "prod"表示生产环境检查，"test"表示测试环境检查
// 返回值：是否有权限，错误信息
func (pa *ProjectAuthRequst) CheckProjectPermission(env string) (bool, string) {
	// 超级管理员直接放行
	if pa.RoleLevel == 0 {
		return true, ""
	}

	// 根据环境类型判断是否忽略权限
	switch env {
	case "prod":
		middleware.Logger.Infof("开始检查生产环境的权限")
		if pa.OnlineAssets {
			// 查询所有项目（由页面具体逻辑获取数据）
			var count int64
			if err := database.DB.Model(&system.DeptProject{}).Count(&count).Error; err != nil {
				return false, "查询项目失败"
			}

			if count > 0 {
				return true, ""
			}
			return false, "没有可访问的项目"
		}
	case "test":
		middleware.Logger.Infof("开始检查测试环境的权限")
		if pa.TestAssets {
			// 查询所有项目（由页面具体逻辑获取数据）
			var count int64
			if err := database.DB.Model(&system.DeptProject{}).Count(&count).Error; err != nil {
				return false, "查询项目失败"
			}

			if count > 0 {
				return true, ""
			}
			return false, "没有可访问的项目"
		}
	}

	// 默认情况下需要检查部门项目权限
	var count int64
	if err := database.DB.Model(&system.DeptProject{}).
		Where("dept_id = ?", pa.DeptID).
		Count(&count).Error; err != nil {
		return false, "验证项目权限失败"
	}

	if count == 0 {
		return false, "无权访问该项目"
	}

	return true, ""
}

// GetAccessibleProjects 获取可访问的项目列表
// 参数env: "prod"表示获取生产环境项目，"test"表示获取测试环境项目
// 权限优先级：超级管理员 > 特殊权限 > 用户直接关联项目 ∪ 部门关联项目
func (pa *ProjectAuthRequst) GetAccessibleProjects(env string) ([]string, error) {
	// 超级管理员可以访问所有项目
	if pa.RoleLevel == 0 {
		var projects []string
		if err := database.DB.Model(&system.DeptProject{}).
			Distinct("project").
			Pluck("project", &projects).Error; err != nil {
			return nil, err
		}
		return projects, nil
	}

	// 根据环境类型判断是否可以直接访问所有项目
	switch env {
	case "prod":
		if pa.OnlineAssets {
			// 可以访问所有项目（不管是生产还是测试）
			var projects []string
			if err := database.DB.Model(&system.DeptProject{}).
				Pluck("project", &projects).Error; err != nil {
				return nil, err
			}
			return projects, nil
		}
	case "test":
		if pa.TestAssets {
			// 可以访问所有项目（不管是生产还是测试）
			var projects []string
			if err := database.DB.Model(&system.DeptProject{}).
				Pluck("project", &projects).Error; err != nil {
				return nil, err
			}
			return projects, nil
		}
	default:
		return nil, nil
	}

	// 1. 获取用户直接关联的项目（根据菜单路径过滤）
	var userProjects []string
	var menuID int64
	query := database.DB.Model(&system.UserProject{}).Where("user_id = ?", pa.UserID)
	if pa.MenuPath != "" {
		// 根据路径查询菜单ID
		menuID = GetMenuIDByPath(pa.MenuPath)
		if menuID > 0 {
			query = query.Where("menu_id = ?", menuID)
		}
	}
	if err := query.Pluck("project", &userProjects).Error; err != nil {
		return nil, err
	}
	middleware.Logger.Infof("[权限] 用户直接关联项目: userID=%d, menuPath=%s, menuID=%d, userProjects=%v", pa.UserID, pa.MenuPath, menuID, userProjects)

	// 2. 获取用户部门关联的项目
	var deptProjects []string
	if err := database.DB.Model(&system.DeptProject{}).
		Where("dept_id = ?", pa.DeptID).
		Pluck("project", &deptProjects).Error; err != nil {
		return nil, err
	}
	middleware.Logger.Infof("[权限] 部门关联项目: deptID=%d, deptProjects=%v", pa.DeptID, deptProjects)

	// 3. 合并去重（用户项目 ∪ 部门项目）
	projectMap := make(map[string]bool)
	for _, p := range userProjects {
		projectMap[p] = true
	}
	for _, p := range deptProjects {
		projectMap[p] = true
	}

	// 转换为切片
	var projects []string
	for p := range projectMap {
		projects = append(projects, p)
	}
	middleware.Logger.Infof("[权限] 最终可访问项目: userID=%d, projects=%v", pa.UserID, projects)

	return projects, nil
}

// ProjectAuth 创建项目权限验证实例
// menuPath: 当前请求的路由路径，如 path.KeyEncryptPath
func ProjectAuth(userID, deptID, roleID uint, roleLevel int64, onlineAssets, testAssets bool, menuPath string) *ProjectAuthRequst {
	return &ProjectAuthRequst{
		UserID:       userID,
		DeptID:       deptID,
		RoleID:       roleID,
		RoleLevel:    roleLevel,
		OnlineAssets: onlineAssets,
		TestAssets:   testAssets,
		MenuPath:     menuPath,
	}
}

// GetMenuIDByPath 根据请求路径获取菜单ID
// 路由格式: /api/assets/proUpdate/list -> 匹配菜单 /assets/proUpdate
func GetMenuIDByPath(urlPath string) int64 {
	// 移除 /api 前缀
	path := strings.TrimPrefix(urlPath, "/api")

	// 分割路径，取前两段作为菜单路径
	// 例如: /assets/proUpdate/list -> /assets/proUpdate
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		middleware.Logger.Debugf("[权限] 路径解析失败: urlPath=%s, parts=%v", urlPath, parts)
		return 0
	}
	menuPath := "/" + parts[0] + "/" + parts[1]

	// 查询菜单ID
	var menu system.Menu
	if err := database.DB.Where("path = ?", menuPath).First(&menu).Error; err != nil {
		middleware.Logger.Debugf("[权限] 菜单查询失败: menuPath=%s, err=%v", menuPath, err)
		return 0
	}
	middleware.Logger.Infof("[权限] 路径匹配菜单: urlPath=%s -> menuPath=%s -> menuID=%d", urlPath, menuPath, menu.ID)
	return menu.ID
}

// CheckRequestProjectPermission 从请求上下文获取用户信息并检查项目权限
// 自动根据请求URL匹配菜单ID，用于过滤用户单独绑定的项目
// 参数：r - HTTP请求, env - 环境类型("prod"或"test")
// 返回：可访问的项目列表，是否超级管理员，错误信息
func CheckRequestProjectPermission(r *http.Request, env string) ([]string, bool, error) {
	// 获取当前用户信息
	userInfo, ok := r.Context().Value(UserInfoKey).(UserInfo)
	if !ok {
		return nil, false, fmt.Errorf("获取用户信息失败")
	}

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		return nil, false, fmt.Errorf("获取角色信息失败: %v", err)
	}

	// 创建权限检查实例，传入请求路径自动匹配菜单
	auth := ProjectAuth(
		uint(userInfo.UserID),
		uint(userInfo.DeptID),
		uint(userInfo.RoleID),
		role.Level,
		userInfo.OnlineAssets,
		userInfo.TestAssets,
		r.URL.Path,
	)

	// 获取可访问项目列表
	projects, err := auth.GetAccessibleProjects(env)
	if err != nil {
		return nil, false, fmt.Errorf("查询可访问项目失败: %v", err)
	}

	// 判断是否为超级管理员
	isSuperAdmin := (role.Level == 0)

	return projects, isSuperAdmin, nil
}

// Contains 检查字符串是否在切片中
func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// ContainsInt64 检查int64值是否在切片中
func ContainsInt64(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
