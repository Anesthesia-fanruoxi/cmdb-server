package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 请求结构体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username"` // 用户名（必填）
	Password string `json:"password"` // 密码（必填）
	Nickname string `json:"nickname"` // 昵称（选填）
	Phone    string `json:"phone"`    // 手机号（选填）
	Email    string `json:"email"`    // 邮箱（选填）
	RoleID   uint   `json:"role_id"`  // 角色ID（选填，默认为普通用户）
	DeptID   uint   `json:"dept_id"`  // 部门ID（必填）
}

// 请求参数结构体
type UserListRequest struct {
	Page     int    `json:"page"`      // 页码，从1开始
	PageSize int    `json:"page_size"` // 每页数量
	Username string `json:"username"`  // 用户名模糊搜索
	Email    string `json:"email"`     // 邮箱模糊搜索
}

// 结果体
type UserListResponse struct {
	Total int64      `json:"total"`
	List  []UserInfo `json:"list"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	IsEnabled bool   `json:"is_enabled"`
	Role      struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"role"`
	Department struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"department"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID       uint   `json:"id"`                 // 用户ID（必填）
	Nickname string `json:"nickname,omitempty"` // 昵称（选填）
	Email    string `json:"email,omitempty"`    // 邮箱（选填）
	Phone    string `json:"phone,omitempty"`    // 手机号（选填）
	Password string `json:"password,omitempty"` // 密码（选填）
	RoleID   uint   `json:"role_id,omitempty"`  // 角色ID（选填）
}

// CreateUser 创建用户
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填参数
	if req.Username == "" || req.Password == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户名和密码不能为空")
		return
	}

	// 验证用户名格式
	if !utils.ValidateUsername(req.Username) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户名格式错误：必须以字母开头，只能包含字母、数字和下划线，长度4-32位")
		return
	}

	// 验证密码格式
	if !utils.ValidatePassword(req.Password) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "密码格式错误：必须包含大小写字母和数字，长度8-32位")
		return
	}

	// 验证昵称格式
	if req.Nickname != "" && !utils.ValidateNickname(req.Nickname) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "昵称格式错误：长度必须在2-32位之间")
		return
	}

	// 验证手机号格式
	if req.Phone != "" && !utils.Validatephone(req.Phone) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "手机号格式错误")
		return
	}

	// 验证邮箱格式
	if req.Email != "" && !utils.ValidateEmail(req.Email) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "邮箱格式错误")
		return
	}

	// 如果未指定角色，默认为普通用户
	if req.RoleID == 0 {
		req.RoleID = 2 // 普通用户角色ID为2
	}

	// 如果未指定部门，设置为默认部门
	if req.DeptID == 0 {
		req.DeptID = 99 // 未分类ID为99
	}

	gormDB := initData.GetDB()

	// 开启事务
	tx := gormDB.Begin()
	if tx.Error != nil {
		log.Printf("开启事务失败: %v", tx.Error)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建用户失败")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查部门是否存在
	var dept models.Department
	if err := tx.First(&dept, req.DeptID).Error; err != nil {
		tx.Rollback()
		log.Printf("查询部门失败: %v", err)
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门不存在")
		return
	}

	// 检查角色是否存在
	var role models.Role
	if err := tx.First(&role, req.RoleID).Error; err != nil {
		tx.Rollback()
		log.Printf("查询角色失败: %v", err)
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色不存在")
		return
	}

	// 检查用户名是否已存在
	var count int64
	if err := tx.Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		tx.Rollback()
		log.Printf("检查用户名失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建用户失败")
		return
	}

	if count > 0 {
		tx.Rollback()
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户名已存在")
		return
	}

	// 创建用户
	user := &models.User{
		Username:  req.Username,
		Nickname:  req.Nickname,
		Email:     req.Email,
		Phone:     req.Phone,
		RoleID:    req.RoleID,
		IsEnabled: true, // 默认启用
		DeptID:    req.DeptID,
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		tx.Rollback()
		log.Printf("设置密码失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建用户失败")
		return
	}

	// 保存到数据库
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		log.Printf("保存用户失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建用户失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("提交���务失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建用户失败")
		return
	}

	// 返回成功响应
	utils.Success(w, map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"role": map[string]interface{}{
			"id":   role.ID,
			"name": role.Name,
			"code": role.Code,
		},
		"department": map[string]interface{}{
			"id":   dept.ID,
			"name": dept.Name,
			"code": dept.Code,
		},
	})
}

// Login 用户登录
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析登录请求失败: %v", err)
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证请求参数
	if req.Username == "" || req.Password == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户名和密码不能为空")
		return
	}

	gormDB := initData.GetDB()

	// 查询用户
	var user models.User
	result := gormDB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		log.Printf("查询用户失败: %v", result.Error)
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "用户名或密码错误")
		return
	}

	// 检查用户是否被禁用
	if !user.IsEnabled {
		log.Printf("用户 %s 已被禁用", user.Username)
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "用户已被禁用，请联系管理员")
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		log.Printf("用户 %s 密码验证失败", user.Username)
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "用户名或密码错误")
		return
	}

	// 查询用户的角色和部门信息
	var userInfo struct {
		models.User
		RoleName string
		RoleCode string
		DeptName string
		DeptCode string
	}

	err := gormDB.Table("users").
		Select("users.*, roles.name as role_name, roles.code as role_code, departments.name as dept_name, departments.code as dept_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Joins("LEFT JOIN departments ON users.dept_id = departments.id").
		Where("users.id = ?", user.ID).
		First(&userInfo).Error

	if err != nil {
		log.Printf("查询用户详细信息失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户信息失���")
		return
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("生成token失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "生成token失败")
		return
	}

	// 返回登录成功信息
	utils.Success(w, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       userInfo.ID,
			"username": userInfo.Username,
			"nickname": userInfo.Nickname,
			"email":    userInfo.Email,
			"phone":    userInfo.Phone,
			"role": map[string]interface{}{
				"id":   userInfo.RoleID,
				"name": userInfo.RoleName,
				"code": userInfo.RoleCode,
			},
			"department": map[string]interface{}{
				"id":   userInfo.DeptID,
				"name": userInfo.DeptName,
				"code": userInfo.DeptCode,
			},
		},
	})
}

// Logout 用户登出
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 获取Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "未提供认证token")
		return
	}

	// 检查Bearer前缀
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "token格式错误")
		return
	}

	// 解析token获取用户信息
	claims, err := utils.ParseToken(parts[1])
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "无效的token")
		return
	}

	// 从Redis删除token
	err = initData.DeleteToken(claims.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "登出失败")
		return
	}

	utils.Success(w, nil)
}

// ListUsers 获取用户列表
func ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析查询参数
	params := r.URL.Query()
	req := UserListRequest{
		Page:     1,  // 默认第1页
		PageSize: 10, // 默认每页10条
	}

	// 获取分页参数
	if page := params.Get("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			req.Page = pageNum
		}
	}
	if pageSize := params.Get("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size > 0 {
			req.PageSize = size
		}
	}

	// 获取搜索参数
	req.Username = params.Get("username")
	req.Email = params.Get("email")

	// 构建查询条件
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if req.Username != "" {
		conditions = append(conditions, "users.username LIKE ?")
		args = append(args, "%"+req.Username+"%")
	}
	if req.Email != "" {
		conditions = append(conditions, "users.email LIKE ?")
		args = append(args, "%"+req.Email+"%")
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	gormDB := initData.GetDB()

	// 构建基础查询
	baseQuery := gormDB.Table("users").
		Select("users.*, roles.name as role_name, roles.code as role_code, departments.name as dept_name, departments.code as dept_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Joins("LEFT JOIN departments ON users.dept_id = departments.id")

	// 添加条件查询
	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		baseQuery = baseQuery.Where(whereClause, args...)
	}

	// 获取总数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		log.Printf("获取总数失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户列表���败")
		return
	}

	// 查询用户列表（包含角色和部门信息）
	var users []struct {
		models.User
		RoleName string
		RoleCode string
		DeptName string
		DeptCode string
	}

	// 分页查询
	if err := baseQuery.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		log.Printf("查询用户列表失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户列表失败")
		return
	}

	// 转换为响应格式
	userList := make([]UserInfo, 0, len(users))
	for _, user := range users {
		userList = append(userList, UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Phone:     user.Phone,
			IsEnabled: user.IsEnabled,
			Role: struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
				Code string `json:"code"`
			}{
				ID:   user.RoleID,
				Name: user.RoleName,
				Code: user.RoleCode,
			},
			Department: struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
				Code string `json:"code"`
			}{
				ID:   user.DeptID,
				Name: user.DeptName,
				Code: user.DeptCode,
			},
		})
	}

	// 返回响应
	utils.Success(w, map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"list":      userList,
	})
}

// DeleteUser 删除用户
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if req.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户ID不能为空")
		return
	}

	gormDB := initData.GetDB()

	// 查询要删除的用户及其角色信息
	type UserWithRole struct {
		models.User
		RoleCode string
	}
	var targetUser UserWithRole

	err := gormDB.Table("users").
		Select("users.*, roles.code as role_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Where("users.id = ?", req.ID).
		First(&targetUser).Error

	if err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "用户不存在")
		return
	}

	// 检查是否管理员账号
	if targetUser.RoleCode == "admin" {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "不能删除管理员账号")
		return
	}

	// 获取当前操作用户的信息（从token中获取）
	currentUserID := r.Header.Get("UserID")
	if currentUserID == "" {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "未授权的操作")
		return
	}

	// 检查是否在删除自己
	if currentUserID == strconv.FormatUint(uint64(targetUser.ID), 10) {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "不能删除自己的账号")
		return
	}

	// 执行删除操作（真删除）
	if err := gormDB.Unscoped().Delete(&models.User{}, targetUser.ID).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除用户失败")
		return
	}

	// 删除用户的token
	initData.DeleteToken(targetUser.ID)

	utils.Success(w, map[string]interface{}{
		"user_id": targetUser.ID,
	})
}

// UpdateUser 更新用户信息
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析请求体为map，以支持动态字段更新
	var reqMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqMap); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "��效的请求参数")
		return
	}

	// 获取要修改的用户ID
	userID, ok := reqMap["id"].(float64)
	if !ok || userID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户ID不能为空")
		return
	}

	// 从token中获取当前用户信息
	claims, err := utils.GetUserFromToken(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, err.Error())
		return
	}

	gormDB := initData.GetDB()

	// 获取当前用户的完整信息��包括角色）
	var currentUser struct {
		models.User
		RoleCode string
	}
	if err := gormDB.Table("users").
		Select("users.*, roles.code as role_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Where("users.id = ?", claims.UserID).
		First(&currentUser).Error; err != nil {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "获取用户信息失败")
		return
	}

	// 判断是否是管理员
	isAdmin := currentUser.RoleCode == "admin"

	// 如果不是管理员且不是修改自己的信息
	if !isAdmin && claims.UserID != uint(userID) {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "只能修改自己的信息")
		return
	}

	// 查询要更新的用户
	var targetUser models.User
	if err := gormDB.First(&targetUser, uint(userID)).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "用户不存在")
		return
	}

	// 创建更新map
	updates := make(map[string]interface{})

	// 定义普通用户可以修改的字段
	normalUserFields := map[string]bool{
		"nickname": true,
		"email":    true,
		"phone":    true,
		"password": true,
	}

	// 遍历请求体中的字段
	for key, value := range reqMap {
		// 跳过ID字段
		if key == "id" {
			continue
		}

		// 如果不是管理员，只允许修改特定字段
		if !isAdmin && !normalUserFields[key] {
			utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "没有权限修改该字段: "+key)
			return
		}

		// 根据字段类型进行验证和处理
		switch key {
		case "password":
			if password, ok := value.(string); ok {
				if !utils.ValidatePassword(password) {
					utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "密码格式错误")
					return
				}
				// 使用 SetPassword 方法来设置密码，确保使用正确的盐值
				if err := targetUser.SetPassword(password); err != nil {
					utils.Error(w, http.StatusInternalServerError, utils.ERROR, "密码加密失败")
					return
				}
				updates[key] = targetUser.Password
			}
		case "nickname":
			if nickname, ok := value.(string); ok {
				if !utils.ValidateNickname(nickname) {
					utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "昵称格式错误")
					return
				}
				updates[key] = nickname
			}
		case "email":
			if email, ok := value.(string); ok {
				if !utils.ValidateEmail(email) {
					utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "邮箱格式错误")
					return
				}
				updates[key] = email
			}
		case "phone":
			if phone, ok := value.(string); ok {
				if !utils.Validatephone(phone) {
					utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "手机号格式错误")
					return
				}
				updates[key] = phone
			}
		default:
			// 管理员可以修改其他字段
			if isAdmin {
				updates[key] = value
			}
		}
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "没有要更新的字段")
		return
	}

	// 添加更新时间
	updates["updated_at"] = time.Now()

	// 执行更新
	if err := gormDB.Model(&targetUser).Updates(updates).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新用户信息失败")
		return
	}

	// 查询更新后的完整用户信息
	var userInfo struct {
		models.User
		RoleName string
		RoleCode string
		DeptName string
		DeptCode string
	}

	err = gormDB.Table("users").
		Select("users.*, roles.name as role_name, roles.code as role_code, departments.name as dept_name, departments.code as dept_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Joins("LEFT JOIN departments ON users.dept_id = departments.id").
		Where("users.id = ?", targetUser.ID).
		First(&userInfo).Error

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取更新后的用户信息失败")
		return
	}

	// 返回更新后的用户信息
	utils.Success(w, UserInfo{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Phone:     userInfo.Phone,
		IsEnabled: userInfo.IsEnabled,
		Role: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   userInfo.RoleID,
			Name: userInfo.RoleName,
			Code: userInfo.RoleCode,
		},
		Department: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   userInfo.DeptID,
			Name: userInfo.DeptName,
			Code: userInfo.DeptCode,
		},
	})
}

// GetUserInfo 获取当前用户信息
func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 请求头获取当前用户ID
	currentUserID := r.Header.Get("UserID")
	if currentUserID == "" {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "未授权的操作")
		return
	}

	// 将用户ID转换为uint
	uid, err := strconv.ParseUint(currentUserID, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "无效的用户ID")
		return
	}

	gormDB := initData.GetDB()

	// 查询用户信息（包含角色和部门信息）
	var userInfo struct {
		models.User
		RoleName string
		RoleCode string
		DeptName string
		DeptCode string
	}

	err = gormDB.Table("users").
		Select("users.*, roles.name as role_name, roles.code as role_code, departments.name as dept_name, departments.code as dept_code").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Joins("LEFT JOIN departments ON users.dept_id = departments.id").
		Where("users.id = ?", uid).
		First(&userInfo).Error

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取用户信息失败")
		return
	}

	// 转换为响应格式
	response := UserInfo{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Phone:     userInfo.Phone,
		IsEnabled: userInfo.IsEnabled,
		Role: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   userInfo.RoleID,
			Name: userInfo.RoleName,
			Code: userInfo.RoleCode,
		},
		Department: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   userInfo.DeptID,
			Name: userInfo.DeptName,
			Code: userInfo.DeptCode,
		},
	}

	utils.Success(w, response)
}
