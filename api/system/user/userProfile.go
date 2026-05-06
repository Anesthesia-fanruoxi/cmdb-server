package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"net/http"
)

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	NickName string `json:"nick_name" validate:"omitempty,max=32"`    // 可选字段，最大32字符
	Email    string `json:"email" validate:"omitempty,email,max=128"` // 可选字段，需要是有效的邮箱
	Phone    string `json:"phone" validate:"omitempty,len=11"`        // 可选字段，11位手机号
	Password string `json:"password" validate:"omitempty,password"`   // 可选字段，需要符合密码规则
}

// ProfileUser 个人信息用户结构（排除敏感字段）
type ProfileUser struct {
	ID            int64       `json:"id"`
	RoleID        int64       `json:"role_id"`
	IsEnabled     bool        `json:"is_enabled"`
	IsDefaultPass bool        `json:"is_default_pass"`
	OnlineAssets  bool        `json:"online_assets"`
	TestAssets    bool        `json:"test_assets"`
	UserName      string      `json:"user_name"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	NickName      string      `json:"nick_name"`
	Email         string      `json:"email"`
	Phone         string      `json:"phone"`
	DeptID        int64       `json:"dept_id"`
	OtpEnabled    bool        `json:"otp_enabled"`
	Role          system.Role `json:"role"`
}

// ProfileResponse 个人信息响应
type ProfileResponse struct {
	ProfileUser
	Permissions []string `json:"permissions"` // 权限标识列表
}

// GetProfile 获取用户个人信息
func GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "未登录")
		return
	}

	// 查询用户详细信息
	var user system.User
	if err := database.DB.Where("id = ?", userInfo.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
			return
		}
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户失败")
		return
	}

	// 查询角色信息
	var role system.Role
	if user.RoleID > 0 {
		if err := database.DB.Where("id = ?", user.RoleID).First(&role).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询角色失败")
			return
		}
	}

	// 查询用户的权限
	var roleMenus []system.RoleMenu
	if err := database.DB.Where("role_id = ?", user.RoleID).Find(&roleMenus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户权限失败")
		return
	}

	// 获取菜单信息
	var menus []system.Menu
	menuIDs := make([]int64, 0, len(roleMenus))
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuID)
	}
	if err := database.DB.Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询菜单失败")
		return
	}

	// 构建菜单ID到权限的映射
	menuMap := make(map[int64]system.Menu)
	for _, menu := range menus {
		menuMap[menu.ID] = menu
	}

	// 构建权限列表
	var permissions []string
	for _, rm := range roleMenus {
		if menu, ok := menuMap[rm.MenuID]; ok && menu.Permission != "" {
			if rm.HasRead {
				permissions = append(permissions, menu.Permission+":r")
			}
			if rm.HasWrite {
				permissions = append(permissions, menu.Permission+":w")
			}
		}
	}

	// 构建响应（排除敏感字段）
	response := ProfileResponse{
		ProfileUser: ProfileUser{
			ID:            user.ID,
			RoleID:        user.RoleID,
			IsEnabled:     user.IsEnabled,
			IsDefaultPass: user.IsDefaultPass,
			OnlineAssets:  user.OnlineAssets,
			TestAssets:    user.TestAssets,
			UserName:      user.UserName,
			CreatedAt:     user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     user.UpdatedAt.Format("2006-01-02 15:04:05"),
			NickName:      user.NickName,
			Email:         user.Email,
			Phone:         user.Phone,
			DeptID:        user.DeptID,
			OtpEnabled:    user.OtpEnabled,
			Role:          role,
		},
		Permissions: permissions,
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// UpdateProfile 更新个人信息
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求体
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 构建更新数据
	updates := make(map[string]interface{})

	// 只更新提供的字段
	if req.NickName != "" {
		updates["nick_name"] = req.NickName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Password != "" {
		hashedPassword, err := utils.EncryptPassword(req.Password)
		if err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "密码加密失败")
			return
		}
		updates["password"] = hashedPassword
		updates["is_default_pass"] = 0
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 执行更新
	if err := tx.Model(&system.User{}).Where("id = ?", userInfo.UserID).Updates(updates).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新个人信息失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "提交事务失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
