package initialize

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"fmt"
	"strings"
	"time"
)

const (
	// SystemDeptID 系统部门ID
	SystemDeptID = 1
)

// InitSuperAdminMenus 为超级管理员角色绑定所有菜单权限
func InitSuperAdminMenus(roleID int64) error {
	// 获取所有菜单ID
	var menus []system.Menu
	if err := database.DB.Find(&menus).Error; err != nil {
		return fmt.Errorf("获取菜单列表失败: %v", err)
	}

	// 开启事务
	tx := database.DB.Begin()

	// 1. 删除该角色之前的所有权限
	if err := tx.Where("role_id = ?", roleID).Delete(&system.RoleMenu{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除原有权限失败: %v", err)
	}

	// 2. 构建批量插入的SQL
	if len(menus) > 0 {
		var values []string
		var params []interface{}
		now := time.Now().Format("2006-01-02 15:04:05")

		for _, menu := range menus {
			values = append(values, "(?, ?, 1, 1, 1, ?)")
			params = append(params, roleID, menu.ID, now)
		}

		// 构建完整的插入SQL
		insertSQL := fmt.Sprintf(
			"INSERT INTO %s (role_id, menu_id, has_menu, has_read, has_write, created_at) VALUES %s",
			middleware.TableRoleMenus,
			strings.Join(values, ","),
		)

		// 执行批量插入
		if err := tx.Exec(insertSQL, params...).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("批量创建权限失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// InitSystemDeptProjects 为系统部门绑定所有项目
func InitSystemDeptProjects() error {
	// 获取所有项目
	var projects []system.ProjectDict
	if err := database.DB.Find(&projects).Error; err != nil {
		return fmt.Errorf("获取项目列表失败: %v", err)
	}

	// 开启事务
	tx := database.DB.Begin()

	// 1. 删除系统部门之前的所有项目关联
	if err := tx.Where("dept_id = ?", SystemDeptID).Delete(&system.DeptProject{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除原有项目关联失败: %v", err)
	}

	// 2. 构建批量插入的SQL
	if len(projects) > 0 {
		var values []string
		var params []interface{}

		for _, project := range projects {
			values = append(values, "(?, ?)")
			params = append(params, SystemDeptID, project.Project)
		}

		// 构建完整的插入SQL
		insertSQL := fmt.Sprintf(
			"INSERT INTO %s (dept_id, project) VALUES %s",
			middleware.TableDeptProjects,
			strings.Join(values, ","),
		)

		// 执行批量插入
		if err := tx.Exec(insertSQL, params...).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("批量创建项目关联失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// InitSystemPrivileges 初始化系统权限
func InitSystemPrivileges() error {
	// 查找超级管理员角色（Level = 0）
	var superAdminRole system.Role
	if err := database.DB.Where("level = ?", 0).First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("获取超级管理员角色失败: %v", err)
	}

	// 初始化超级管理员菜单权限
	if err := InitSuperAdminMenus(superAdminRole.ID); err != nil {
		return err
	}

	// 初始化系统部门项目关联
	if err := InitSystemDeptProjects(); err != nil {
		return err
	}

	return nil
}
