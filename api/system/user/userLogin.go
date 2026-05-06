package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/pkg/security"
	"cmdb/utils"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// 登录类型常量
const (
	LoginTypePassword = "password" // 密码登录
	LoginTypeTOTP     = "totp"     // TOTP验证码登录
)

// LoginRequest 登录请求
type LoginRequest struct {
	UserName  string `json:"user_name" validate:"required"`
	LoginType string `json:"login_type" validate:"required,oneof=password totp"`
	Password  string `json:"password,omitempty"`  // 密码登录时必填
	TOTPCode  string `json:"totp_code,omitempty"` // TOTP登录时必填
}

// PasswordRules 密码规则
type PasswordRules struct {
	MinLength   int  `json:"min_length"`   // 最小长度
	MaxLength   int  `json:"max_length"`   // 最大长度
	NeedNumber  bool `json:"need_number"`  // 是否需要数字
	NeedLetter  bool `json:"need_letter"`  // 是否需要字母
	NeedCase    bool `json:"need_case"`    // 是否需要大小写
	NeedSpecial bool `json:"need_special"` // 是否需要特殊字符
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token              string         `json:"token"`
	UserID             int64          `json:"user_id"`
	UserName           string         `json:"user_name"`
	RoleID             int64          `json:"role_id"`
	RoleName           string         `json:"role_name"`
	IsDefaultPass      bool           `json:"is_default_pass"`          // 是否是默认密码
	OtpEnabled         bool           `json:"otp_enabled"`              // 是否已启用双因子认证
	AllowPasswordLogin bool           `json:"allow_password_login"`     // 是否允许密码登录
	PasswordRules      *PasswordRules `json:"password_rules,omitempty"` // 密码规则，仅当是默认密码时返回
}

// UserLogin 用户登录接口
func UserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取客户端IP地址
	clientIP := utils.GetClientIP(r)

	// 检查IP限流
	if err := security.CheckIPRateLimit(r.Context(), clientIP); err != nil {
		middleware.ResponseError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	// 检查IP是否在黑名单中
	if err := security.CheckIPBlacklist(clientIP); err != nil {
		middleware.ResponseError(w, http.StatusForbidden, err.Error())
		return
	}

	// 解析请求体
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 验证根据登录类型进行额外参数验证
	switch req.LoginType {
	case LoginTypePassword:
		if req.Password == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "密码不能为空")
			return
		}
	case LoginTypeTOTP:
		if req.TOTPCode == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "TOTP验证码不能为空")
			return
		}
	default:
		middleware.ResponseError(w, http.StatusBadRequest, "不支持的登录类型")
		return
	}

	// 1. 查询用户
	var user system.User
	var count int64
	if err := database.DB.Table(middleware.TableUsers).
		Where(middleware.TableUsers+".user_name = ?", req.UserName).
		Count(&count).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "系统错误")
		return
	}

	if count == 0 {
		// 记录IP登录失败
		if err := security.IncrementIPLoginFailed(clientIP); err != nil {
			middleware.Logger.Errorf("记录IP登录失败次数失败: %v", err)
		}

		// 记录登录失败日志
		utils.RecordLoginLog(r, 0, req.UserName, false, "用户不存在")
		middleware.ResponseError(w, http.StatusBadRequest, "用户不存在")
		return
	}

	// 查询用户详细信息
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
		Where(middleware.TableUsers+".user_name = ?", req.UserName).
		Scan(&user).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "系统错误")
		return
	}

	// 2. 检查用户状态
	if !user.IsEnabled {
		// 记录IP登录失败
		if err := security.IncrementIPLoginFailed(clientIP); err != nil {
			middleware.Logger.Errorf("记录IP登录失败次数失败: %v", err)
		}

		// 记录登录失败日志
		utils.RecordLoginLog(r, user.ID, user.UserName, false, "用户已被禁用")
		middleware.ResponseError(w, http.StatusForbidden, "用户已被禁用")
		return
	}

	// 3. 检查登录安全（如果开启了登录保护）
	if err := security.CheckLoginSecurity(req.UserName); err != nil {
		// 记录IP登录失败
		if err := security.IncrementIPLoginFailed(clientIP); err != nil {
			middleware.Logger.Errorf("记录IP登录失败次数失败: %v", err)
		}

		// 记录登录失败日志
		utils.RecordLoginLog(r, user.ID, user.UserName, false, err.Error())
		middleware.ResponseError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	// 4. 根据登录类型进行验证
	loginSuccess := false
	loginFailMessage := ""

	switch req.LoginType {
	case LoginTypePassword:
		// 检查用户是否允许使用密码登录
		if !user.AllowPasswordLogin {
			loginFailMessage = "您的账户已启用双因子认证，请使用TOTP验证码登录"
		} else if !utils.ComparePassword(user.Password, req.Password) {
			loginFailMessage = "用户名或密码错误"
		} else {
			loginSuccess = true
		}
	case LoginTypeTOTP:
		// 验证TOTP
		if !user.OtpEnabled {
			loginFailMessage = "该账户未启用TOTP验证"
		} else if !validateTOTP(user.OtpSecret, req.TOTPCode) {
			loginFailMessage = "TOTP验证码无效"
		} else {
			loginSuccess = true
		}
	}

	if !loginSuccess {
		// 记录登录失败
		if err := security.IncrementLoginFailed(req.UserName); err != nil {
			middleware.Logger.Errorf("记录登录失败次数失败: %v", err)
		}

		// 记录IP登录失败
		if err := security.IncrementIPLoginFailed(clientIP); err != nil {
			middleware.Logger.Errorf("记录IP登录失败次数失败: %v", err)
		}

		// 记录登录失败日志
		utils.RecordLoginLog(r, user.ID, user.UserName, false, loginFailMessage)
		middleware.ResponseError(w, http.StatusBadRequest, loginFailMessage)
		return
	}

	// 5. 登录成功，重置失败次数
	if err := security.ResetLoginFailed(req.UserName); err != nil {
		middleware.Logger.Errorf("重置登录失败次数失败: %v", err)
	}

	// 重置IP登录失败次数
	if err := security.ResetIPLoginFailed(clientIP); err != nil {
		middleware.Logger.Errorf("重置IP登录失败次数失败: %v", err)
	}

	// 获取角色权限
	permissions, err := utils.GetRolePermissions(uint(user.RoleID))
	if err != nil {
		if err := utils.RefreshRolePermissions(uint(user.RoleID)); err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "刷新权限缓存失败")
			return
		}
		permissions, err = utils.GetRolePermissions(uint(user.RoleID))
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取权限失败")
			return
		}
	}
	permissionStr := strings.Join(permissions, ",")

	// 生成token（无论是否是默认密码都生成token）
	token, err := utils.GenerateToken(
		user.ID,
		user.UserName,
		user.RoleID,
		user.DeptID,
		permissionStr,
		user.IsDefaultPass,
	)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成token失败")
		return
	}

	// 将token存入Redis并设置过期时间
	middleware.Logger.Infof("[PC端登录] 正在存储Token，用户ID: %d, 键名: cmdb:user:%d:token", user.ID, user.ID)
	if err := utils.StoreToken(user.ID, token); err != nil {
		middleware.Logger.Errorf("[PC端登录] 存储token到Redis失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "存储token失败")
		return
	}

	// 验证存储是否成功
	ctx := context.Background()
	redisKey := middleware.GetUserTokenKey(user.ID)
	storedToken, redisErr := database.RDB.Get(ctx, redisKey).Result()
	if redisErr != nil {
		middleware.Logger.Errorf("[PC端登录] 验证token存储失败，键: %s，错误: %v", redisKey, redisErr)
	} else if storedToken != token {
		middleware.Logger.Errorf("[PC端登录] 存储的token不一致，预期: %s..., 实际: %s...",
			token[:20], storedToken[:20])
	} else {
		middleware.Logger.Infof("[PC端登录] token验证成功，键: %s", redisKey)
	}

	// 记录登录成功日志
	utils.RecordLoginLog(r, user.ID, user.UserName, true, "登录成功")

	// 返回登录信息
	resp := LoginResponse{
		Token:              token,
		UserID:             middleware.ToInt64(user.ID),
		UserName:           user.UserName,
		RoleID:             middleware.ToInt64(user.RoleID),
		RoleName:           user.Role.Name,
		IsDefaultPass:      user.IsDefaultPass,
		OtpEnabled:         user.OtpEnabled,
		AllowPasswordLogin: user.AllowPasswordLogin,
	}

	if user.IsDefaultPass {
		middleware.ResponseSuccess(w, "首次登录请绑定双因子", resp)
	} else {
		middleware.ResponseSuccess(w, "登录成功", resp)
	}
}

// validateTOTP 验证TOTP代码
func validateTOTP(secret, code string) bool {
	return utils.ValidateOTP(secret, code)
}
