package role

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MenuPermMap 菜单权限映射
type MenuPermMap map[int64][]int

// UpdateRoleMenuRequest 更新角色菜单权限请求
type UpdateRoleMenuRequest struct {
	RoleID      int64       `json:"role_id" validate:"required"`
	Permissions MenuPermMap `json:"permissions" validate:"required"`
}

// Role 角色实体
type Role struct {
	ID          int64      `json:"id" gorm:"primaryKey;column:id;type:bigint"`
	Name        string     `json:"name" gorm:"column:name;type:varchar(32);not null"`
	Code        string     `json:"code" gorm:"column:code;type:varchar(32);not null"`
	Description string     `json:"description" gorm:"column:description;type:varchar(128)"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at"`
	Level       int64      `json:"level" gorm:"column:level;type:bigint;not null"`
}

// RoleMenu 角色菜单关联
type RoleMenu struct {
	RoleID    int64      `json:"role_id" gorm:"column:role_id;type:bigint;not null"`
	MenuID    int64      `json:"menu_id" gorm:"column:menu_id;type:bigint;not null"`
	HasMenu   bool       `json:"has_menu" gorm:"column:has_menu;type:tinyint(1);not null;default:0"`
	HasRead   bool       `json:"has_read" gorm:"column:has_read;type:tinyint(1);not null;default:0"`
	HasWrite  bool       `json:"has_write" gorm:"column:has_write;type:tinyint(1);not null;default:0"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName 指定表名
func (Role) TableName() string {
	return middleware.TableRoles
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return middleware.TableRoleMenus
}

// RoleService 角色服务接口
type RoleService interface {
	Create(role *Role) error
	Update(role *Role) error
	Delete(id int64) error
	GetByID(id int64) (*Role, error)
	List(page, size int) ([]*Role, int64, error)
	GetByCode(code string) (*Role, error)
}

// RoleMenuDetail 获取角色菜单权限
func RoleMenuDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取角色ID
	roleIDStr := r.URL.Query().Get("role_id")
	if roleIDStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "角色ID不能为空")
		return
	}

	// 查询角色的菜单权限
	var roleMenus []RoleMenu
	if err := database.DB.Where("role_id = ?", roleIDStr).Find(&roleMenus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询角色菜单权限失败")
		return
	}

	// 构建权限映射
	permissions := make(MenuPermMap)
	for _, rm := range roleMenus {
		perms := []int{0, 0, 0}
		if rm.HasMenu {
			perms[0] = 1
		}
		if rm.HasRead {
			perms[1] = 1
		}
		if rm.HasWrite {
			perms[2] = 1
		}
		permissions[rm.MenuID] = perms
	}

	middleware.ResponseSuccess(w, "查询成功", permissions)
}

// RoleMenuUpdate 更新角色菜单权限
func RoleMenuUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req UpdateRoleMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 获取角色信息
	var role Role
	if err := database.DB.First(&role, req.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "角色不存在")
		return
	}

	// 如果是超级管理员角色，不允许修改权限
	if role.Level == 0 {
		middleware.ResponseError(w, http.StatusForbidden, "超级管理员角色权限不能修改")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 删除原有的角色菜单关联
	if err := tx.Where("role_id = ?", req.RoleID).Delete(&RoleMenu{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除原有权限失败")
		return
	}

	// 创建新的角色菜单关联
	for menuID, perms := range req.Permissions {
		if len(perms) != 3 {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusBadRequest, "权限格式错误")
			return
		}

		roleMenu := RoleMenu{
			RoleID:   req.RoleID,
			MenuID:   menuID,
			HasMenu:  perms[0] == 1,
			HasRead:  perms[1] == 1,
			HasWrite: perms[2] == 1,
		}
		if err := tx.Create(&roleMenu).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "创建权限失败")
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新权限失败")
		return
	}

	// 刷新角色权限缓存
	if err := utils.RefreshRolePermissions(middleware.FromInt64[uint](req.RoleID)); err != nil {
		fmt.Printf("刷新角色(%d)权限缓存失败: %v\n", req.RoleID, err)
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
