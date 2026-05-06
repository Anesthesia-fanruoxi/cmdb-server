package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
)

// FactorAuthResponse 双因子认证响应结构体
type FactorAuthResponse struct {
	Secret      string   `json:"secret"`       // TOTP密钥
	QrCodeURL   string   `json:"qrcode_url"`   // 二维码URL
	BackupCodes []string `json:"backup_codes"` // 备用恢复码
}

// VerifyRequest 验证请求结构体
type VerifyRequest struct {
	Code string `json:"code"` // 验证码
}

// VerifyWithUsernameRequest 使用用户名验证的请求结构体
type VerifyWithUsernameRequest struct {
	UserName string `json:"user_name"` // 用户名
	Code     string `json:"code"`      // 验证码
}

// GenerateFactorAuth 生成双因子认证信息
// @Summary 生成双因子认证信息
// @Description 为当前用户生成双因子认证信息，包括TOTP密钥、二维码URL和备用恢复码
// @Tags 用户-双因子认证
// @Accept json
// @Produce json
// @Success 200 {object} FactorAuthResponse "成功"
// @Failure 500 {object} middleware.Response "内部服务器错误"
// @Router /api/system/user/factor/generate [post]
func GenerateFactorAuth(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 查询用户详细信息
	var user system.User
	if err := database.DB.First(&user, userInfo.UserID).Error; err != nil {
		middleware.Logger.Errorf("查询用户信息失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户信息失败")
		return
	}

	// 调用工具函数生成OTP信息
	username := userInfo.UserName
	secret, qrCodeURL, err := utils.GenerateOTP(username)
	if err != nil {
		middleware.Logger.Errorf("生成OTP失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "生成双因子认证信息失败")
		return
	}

	// 生成备用恢复码
	backupCodes := utils.GenerateBackupCodes(10, 16) // 生成10个16字符长的备用码

	// 保存到用户表
	user.OtpSecret = secret
	user.OtpEnabled = false // 初始为禁用状态
	user.OtpSetupCompleted = false

	// 将备用码转为JSON字符串
	backupCodesJSON, _ := json.Marshal(backupCodes)
	user.OtpBackupCodes = string(backupCodesJSON)

	// 保存用户信息
	if err := database.DB.Save(&user).Error; err != nil {
		middleware.Logger.Errorf("保存用户信息失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "保存双因子认证信息失败")
		return
	}

	// 记录操作日志
	utils.LogFactorAuthActionHTTP(user.ID, "setup", "success", nil, r)

	// 返回响应
	response := FactorAuthResponse{
		Secret:      secret,
		QrCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}

	middleware.ResponseSuccess(w, "成功", response)
}

// VerifyFactorAuth 验证并启用双因子认证
// @Summary 验证并启用双因子认证
// @Description 验证用户输入的验证码，验证成功后启用双因子认证
// @Tags 用户-双因子认证
// @Accept json
// @Produce json
// @Param verifyRequest body VerifyRequest true "验证请求"
// @Success 200 {object} middleware.Response "成功"
// @Failure 400 {object} middleware.Response "验证码无效"
// @Failure 500 {object} middleware.Response "内部服务器错误"
// @Router /api/system/user/factor/verify [post]
func VerifyFactorAuth(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 查询用户详细信息
	var user system.User
	if err := database.DB.First(&user, userInfo.UserID).Error; err != nil {
		middleware.Logger.Errorf("查询用户信息失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询用户信息失败")
		return
	}

	// 获取请求参数
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求失败")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	var req VerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败")
		return
	}

	if req.Code == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "参数错误：验证码不能为空")
		return
	}

	// 验证验证码
	if !utils.ValidateOTP(user.OtpSecret, req.Code) {
		// 记录失败日志
		utils.LogFactorAuthActionHTTP(user.ID, "verify", "failed", nil, r)
		middleware.ResponseError(w, http.StatusBadRequest, "验证码无效")
		return
	}

	// 验证成功，启用双因子认证
	user.OtpEnabled = true
	user.OtpSetupCompleted = true
	user.IsDefaultPass = false      // 将用户标记为非初始密码状态
	user.AllowPasswordLogin = false // 禁用密码登录，强制使用TOTP登录

	// 保存用户信息
	if err := database.DB.Save(&user).Error; err != nil {
		middleware.Logger.Errorf("保存用户信息失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "启用双因子认证失败")
		return
	}

	// 检查当前token存储情况
	middleware.Logger.Infof("[双因子验证] 验证通过后检查token状态，用户ID: %d", user.ID)

	// 检查PC端token
	ctx := database.DB.Statement.Context
	pcKey := middleware.GetUserTokenKey(user.ID)
	pcToken, pcErr := database.RDB.Get(ctx, pcKey).Result()
	if pcErr != nil {
		middleware.Logger.Infof("[双因子验证] PC端token不存在，键: %s，错误: %v", pcKey, pcErr)
	} else {
		middleware.Logger.Infof("[双因子验证] 当前PC端token存在，键: %s，值长度: %d", pcKey, len(pcToken))
	}

	// 检查移动端token
	mobileKey := middleware.GetMobileUserTokenKey(user.ID)
	mobileToken, mobileErr := database.RDB.Get(ctx, mobileKey).Result()
	if mobileErr != nil {
		middleware.Logger.Infof("[双因子验证] 移动端token不存在，键: %s，错误: %v", mobileKey, mobileErr)
	} else {
		middleware.Logger.Infof("[双因子验证] 当前移动端token存在，键: %s，值长度: %d", mobileKey, len(mobileToken))
	}

	// 记录操作日志
	utils.LogFactorAuthActionHTTP(user.ID, "verify", "success", nil, r)

	// 返回成功响应
	middleware.ResponseSuccess(w, "双因子认证已启用", nil)
}

// VerifyFactorAuthWithUsername 使用用户名和验证码进行双因子验证（登录时使用）
// @Summary 使用用户名和验证码进行双因子验证
// @Description 在登录流程中验证用户输入的验证码，不需要token
// @Tags 用户-双因子认证
// @Accept json
// @Produce json
// @Param verifyRequest body VerifyWithUsernameRequest true "验证请求"
// @Success 200 {object} middleware.Response "成功"
// @Failure 400 {object} middleware.Response "验证码无效"
// @Failure 500 {object} middleware.Response "内部服务器错误"
// @Router /api/system/user/factor/verify-login [post]
func VerifyFactorAuthWithUsername(w http.ResponseWriter, r *http.Request) {
	// 获取请求参数
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求失败")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	var req VerifyWithUsernameRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数解析失败")
		return
	}

	// 验证参数
	if req.UserName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "参数错误：用户名不能为空")
		return
	}
	if req.Code == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "参数错误：验证码不能为空")
		return
	}

	// 查询用户信息
	var user system.User
	if err := database.DB.Where("user_name = ?", req.UserName).First(&user).Error; err != nil {
		middleware.Logger.Warnf("用户名不存在或查询失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "用户名或验证码错误")
		return
	}

	// 检查用户是否启用了双因子认证
	if !user.OtpEnabled || !user.OtpSetupCompleted {
		middleware.Logger.Warnf("用户 %s 未启用双因子认证", req.UserName)
		middleware.ResponseError(w, http.StatusBadRequest, "该用户未启用双因子认证")
		return
	}

	// 验证验证码
	if !utils.ValidateOTP(user.OtpSecret, req.Code) {
		// 记录失败日志
		utils.LogFactorAuthActionHTTP(user.ID, "userName-verify", "failed", nil, r)
		middleware.ResponseError(w, http.StatusBadRequest, "验证码无效")
		return
	}

	// 检查当前token存储情况
	middleware.Logger.Infof("[双因子验证] 用户名验证通过后检查token状态，用户ID: %d", user.ID)

	// 检查PC端token
	ctx := database.DB.Statement.Context
	pcKey := middleware.GetUserTokenKey(user.ID)
	pcToken, pcErr := database.RDB.Get(ctx, pcKey).Result()
	if pcErr != nil {
		middleware.Logger.Infof("[双因子验证] PC端token不存在，键: %s，错误: %v", pcKey, pcErr)
	} else {
		middleware.Logger.Infof("[双因子验证] 当前PC端token存在，键: %s，值长度: %d", pcKey, len(pcToken))
	}

	// 检查移动端token
	mobileKey := middleware.GetMobileUserTokenKey(user.ID)
	mobileToken, mobileErr := database.RDB.Get(ctx, mobileKey).Result()
	if mobileErr != nil {
		middleware.Logger.Infof("[双因子验证] 移动端token不存在，键: %s，错误: %v", mobileKey, mobileErr)
	} else {
		middleware.Logger.Infof("[双因子验证] 当前移动端token存在，键: %s，值长度: %d", mobileKey, len(mobileToken))
	}

	// 记录成功日志
	utils.LogFactorAuthActionHTTP(user.ID, "userName-verify", "success", nil, r)

	// 返回成功响应
	middleware.ResponseSuccess(w, "验证成功", nil)
}
