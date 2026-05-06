package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// UserListResponse 用户列表响应
type UserListResponse struct {
	ID                 int64     `json:"id"`
	UserName           string    `json:"user_name"`
	NickName           string    `json:"nick_name"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone"`
	RoleID             int64     `json:"role_id"`
	RoleName           string    `json:"role_name"`
	DeptID             int64     `json:"dept_id"`
	DeptName           string    `json:"dept_name"`
	AllowPasswordLogin bool      `json:"allow_password_login"`
	OtpEnabled         bool      `json:"otp_enabled"`
	IsEnabled          bool      `json:"is_enabled"`
	OnlineAssets       bool      `json:"online_assets"`
	TestAssets         bool      `json:"test_assets"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UserListPageResponse 带分页的用户列表响应
type UserListPageResponse struct {
	Total      int64              `json:"total"`       // 总记录数
	TotalPages int64              `json:"total_pages"` // 总页数
	Page       int                `json:"page"`        // 当前页码
	PageSize   int                `json:"page_size"`   // 每页大小
	List       []UserListResponse `json:"list"`        // 用户列表
}

// UserList 获取用户列表
func UserList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取当前用户的角色级别
	var role system.Role
	if err := database.DB.Table(middleware.TableRoles).
		Select("level").
		Where("id = ?", userInfo.RoleID).
		First(&role).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	name := query.Get("name")                                   // 姓名搜索（模糊查询user_name或nick_name）
	roleID, _ := strconv.ParseInt(query.Get("role_id"), 10, 64) // 角色ID
	deptID, _ := strconv.ParseInt(query.Get("dept_id"), 10, 64) // 部门ID

	// 获取分页参数，默认第一页
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	// 每页固定显示10条记录
	pageSize := 20

	// 构建查询
	db := database.DB.Table(middleware.TableUsers).
		Select(fmt.Sprintf(`
			%[1]s.id, %[1]s.user_name, %[1]s.nick_name, %[1]s.email, %[1]s.phone,
			%[1]s.role_id, %[2]s.name as role_name, 
			%[1]s.dept_id, %[3]s.name as dept_name,
			%[1]s.allow_password_login,
			%[1]s.otp_enabled,
			%[1]s.is_enabled,%[1]s.online_assets, %[1]s.test_assets, %[1]s.created_at, %[1]s.updated_at
		`, middleware.TableUsers, middleware.TableRoles, middleware.TableDepartments)).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.role_id = %s.id",
			middleware.TableRoles, middleware.TableUsers, middleware.TableRoles)).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.dept_id = %s.id",
			middleware.TableDepartments, middleware.TableUsers, middleware.TableDepartments))

	// 根据角色级别限制查询范围
	if role.Level > 0 {
		// 非超级管理员只能查看同部门用户
		db = db.Where(middleware.TableUsers+".dept_id = ?", userInfo.DeptID)
	}

	// 添加查询条件
	if name != "" {
		// 顺序模糊匹配：将输入字符用.*连接，例如 zs -> z.*s
		pattern := ""
		for i, char := range name {
			if i > 0 {
				pattern += ".*"
			}
			pattern += string(char)
		}
		// 使用REGEXP进行正则匹配user_name或nick_name
		db = db.Where(middleware.TableUsers+".user_name REGEXP ? OR "+middleware.TableUsers+".nick_name REGEXP ?",
			pattern, pattern)
	}
	if roleID > 0 {
		db = db.Where(middleware.TableUsers+".role_id = ?", roleID)
	}
	if deptID > 0 {
		db = db.Where(middleware.TableUsers+".dept_id = ?", deptID)
	}

	// 获取总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户总数失败")
		return
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// 计算分页偏移量
	offset := (page - 1) * pageSize

	// 查询当前页数据
	var users []UserListResponse
	if err := db.Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户列表失败")
		return
	}

	// 构建分页响应
	response := UserListPageResponse{
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
		List:       users,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// UserDetail 用户详情
func UserDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取用户ID
	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	// 查询用户
	var user system.User
	if err := database.DB.Table(middleware.TableUsers).
		Select(middleware.TableUsers+".*",
			middleware.TableRoles+".id as `role.id`",
			middleware.TableRoles+".name as `role.name`",
			middleware.TableRoles+".code as `role.code`",
			middleware.TableRoles+".description as `role.description`",
			middleware.TableRoles+".level as `role.level`",
			middleware.TableRoles+".created_at as `role.created_at`",
			middleware.TableRoles+".updated_at as `role.updated_at`").
		Joins("LEFT JOIN "+middleware.TableRoles+" ON "+middleware.TableUsers+".role_id = "+middleware.TableRoles+".id").
		Where(middleware.TableUsers+".id = ?", userID).
		Scan(&user).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
		return
	}

	middleware.ResponseSuccess(w, "查询成功", user)
}
