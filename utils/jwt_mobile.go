package utils

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/routers/path"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"strings"
	"time"
)

// MobileWhiteList 定义移动端不需要验证token的路径
var MobileWhiteList = map[string]bool{
	path.UserLoginPath: true, // 登录接口不需要token
}

// GenerateMobileToken 生成移动端JWT token
func GenerateMobileToken(userID interface{}, userName string, roleID interface{}, deptID interface{}, permissions string, isDefaultPass bool) (string, error) {
	// 获取过期时间 - 移动端使用固定的过期时间
	expireTime := time.Now().Add(10 * time.Minute) // 移动端设置为10分钟

	// 将userID转换为int64
	userIDInt := middleware.ToInt64(userID)
	roleIDInt := middleware.ToInt64(roleID)
	deptIDInt := middleware.ToInt64(deptID)

	// 基础claims
	claims := jwt.MapClaims{
		"user_id":         userIDInt,
		"user_name":       userName,
		"role_id":         roleIDInt,
		"dept_id":         deptIDInt,
		"permissions":     permissions,
		"is_default_pass": isDefaultPass,
		"device_type":     "mobile",
		"iat":             time.Now().Unix(),
		"exp":             expireTime.Unix(),
		"iss":             "cmdb-mobile",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetConfig().JWT.Secret))
	if err != nil {
		return "", err
	}

	// 使用新的方法存储token
	expiration := 10 * time.Minute
	if err := StoreMobileTokenWithRedis(userIDInt, tokenString, expiration); err != nil {
		middleware.Logger.Errorf("存储移动端token失败: %v", err)
		return "", fmt.Errorf("存储token失败: %v", err)
	}

	return tokenString, nil
}

// ParseMobileToken 解析移动端JWT token并获取完整用户信息
func ParseMobileToken(tokenString string, secret string) (UserInfo, error) {
	// 验证token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		middleware.Logger.Errorf("移动端token解析失败: %v", err)
		switch {
		case err.Error() == "Token is expired":
			return UserInfo{}, fmt.Errorf("token已过期")
		default:
			return UserInfo{}, fmt.Errorf("token无效: %v", err)
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok1 := claims["user_id"].(float64)
		if !ok1 {
			middleware.Logger.Errorf("token中用户ID格式错误")
			return UserInfo{}, fmt.Errorf("token中的用户ID格式错误")
		}

		// 使用新的方法从Redis获取token
		userIDInt := int64(userID)
		storedToken, err := GetMobileTokenFromRedis(userIDInt)
		if err != nil {
			middleware.Logger.Errorf("从Redis获取移动端token失败, 键: %s, 错误: %v",
				MobileTokenKey(userIDInt), err)
			return UserInfo{}, fmt.Errorf("token已失效,请重新登录")
		}

		// 检查token是否一致
		if storedToken != tokenString {
			middleware.Logger.Errorf("移动端token不一致，预期: %s，实际: %s",
				tokenString[:10]+"...", storedToken[:10]+"...")
			return UserInfo{}, fmt.Errorf("token已被更新,请重新登录")
		}

		middleware.Logger.Infof("移动端token验证成功，用户ID: %d", userIDInt)

		// 获取完整的用户信息
		return GetUserInfo(int64(userID))
	}

	return UserInfo{}, fmt.Errorf("无效的token")
}

// RefreshMobileTokenExpiration 刷新移动端token的过期时间
func RefreshMobileTokenExpiration(userID int64, tokenString string) error {
	// 获取当前存储的token
	storedToken, err := GetMobileTokenFromRedis(userID)
	if err != nil {
		return fmt.Errorf("token不存在: %v", err)
	}

	// 验证token是否匹配
	if storedToken != tokenString {
		return fmt.Errorf("token不匹配")
	}

	// 仅刷新过期时间，而不是重新存储token
	ctx := context.Background()
	key := middleware.GetMobileUserTokenKey(userID)
	expiration := 10 * time.Minute

	// 使用EXPIRE命令只更新过期时间，不重新写入数据
	success, err := database.RDB.Expire(ctx, key, expiration).Result()
	if err != nil {
		middleware.Logger.Errorf("刷新移动端token过期时间失败: %v", err)
		return err
	}

	if success {
		middleware.Logger.Infof("移动端token过期时间刷新成功 - 用户ID: %d, 键: %s, 新过期时间: %v", userID, key, expiration)
	} else {
		middleware.Logger.Warnf("移动端token不存在，无法刷新过期时间 - 用户ID: %d, 键: %s", userID, key)
	}

	return nil
}

// StoreMobileTokenWithRedis 在Redis中存储移动端token (使用更简单的方式)
func StoreMobileTokenWithRedis(userID int64, tokenString string, expiration time.Duration) error {
	ctx := context.Background()
	key := middleware.GetMobileUserTokenKey(userID)

	// 尝试使用SET命令存储token
	err := database.RDB.Set(ctx, key, tokenString, expiration).Err()
	middleware.Logger.Infof("移动端Token存储 - 用户ID: %d, 键: %s, 过期时间: %v, 结果: %v",
		userID, key, expiration, err)

	// 再次尝试获取以确认存储成功
	stored, getErr := database.RDB.Get(ctx, key).Result()
	if getErr != nil {
		middleware.Logger.Errorf("存储后获取token失败: %v", getErr)
	} else if stored != tokenString {
		middleware.Logger.Errorf("存储的token与预期不符")
	} else {
		middleware.Logger.Infof("移动端token存储成功")
	}

	return err
}

// GetMobileTokenFromRedis 从Redis获取移动端token (使用更简单的方式)
func GetMobileTokenFromRedis(userID int64) (string, error) {
	ctx := context.Background()
	key := middleware.GetMobileUserTokenKey(userID)

	// 尝试使用GET命令获取token
	token, err := database.RDB.Get(ctx, key).Result()
	// 仅记录是否失败，不记录具体内容以避免日志过多
	if err != nil {
		middleware.Logger.Infof("移动端Token获取失败 - 用户ID: %d, 键: %s", userID, key)
	}
	return token, err
}

// MobileJWTAuth 移动端JWT认证中间件
func MobileJWTAuth(next http.HandlerFunc, permission string, logType string, module string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取请求开始时间
		startTime := time.Now()

		// 记录标准日志（控制台输出）
		if logType == "s" || logType == "all" {
			middleware.Logger.Infof("[移动API请求] %s %s", r.Method, r.URL.Path)
		}

		// 处理跨域预检请求
		if r.Method == "OPTIONS" {
			HandleCORS(w, r)
			return
		}

		// 设置跨域响应头
		SetCORSHeaders(w)

		// 检查是否在移动端白名单中
		if MobileWhiteList[r.URL.Path] {
			// 如果是白名单路径，直接执行处理函数
			next(w, r)
			return
		}

		// 获取token
		token := r.Header.Get("Authorization")
		if token == "" {
			middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
			return
		}

		// 记录当前验证的token（截断以避免日志过长）
		if len(token) > 50 {
			middleware.Logger.Infof("移动端Token验证 Token前缀: %s...", token[:50])
		} else {
			middleware.Logger.Infof("移动端Token验证 Token: %s", token)
		}

		// 去掉Bearer前缀
		tokenString := strings.TrimPrefix(token, "Bearer ")

		// 解析token
		jwtConfig := config.GetConfig().JWT
		userInfo, err := ParseMobileToken(tokenString, jwtConfig.Secret)
		if err != nil {
			middleware.ResponseError(w, http.StatusUnauthorized, err.Error())
			return
		}

		// 刷新token过期时间
		if err := RefreshMobileTokenExpiration(userInfo.UserID, tokenString); err != nil {
			middleware.Logger.Warnf("刷新移动端token过期时间失败: %v", err)
		}

		// 检查用户是否已启用双因子认证
		var user system.User
		if err := database.DB.First(&user, userInfo.UserID).Error; err == nil {
			if !user.OtpEnabled || !user.OtpSetupCompleted {
				middleware.ResponseError(w, http.StatusForbidden, "移动端只允许使用已启用双因子认证的账户")
				return
			}
		}

		// 检查是否使用默认密码
		if userInfo.IsDefaultPass {
			// 如果是默认密码，只允许访问修改密码的接口
			if r.URL.Path != path.UserFactorAuthPath && r.URL.Path != path.UserFactorVerifyPath {
				middleware.ResponseError(w, http.StatusUnauthorized, "请先修改默认密码")
				return
			}
		}

		// 权限验证
		if permission != "" {
			// 记录权限验证日志
			middleware.Logger.Infof("[移动端权限验证] 接口: %s, 需要权限: %s",
				r.URL.Path, permission)

			if hasPermission(userInfo.Permissions, permission) {
				middleware.Logger.Infof("[移动端权限验证] ✅ 权限匹配 - 用户: %s, 接口: %s",
					userInfo.UserName, r.URL.Path)
			} else {
				middleware.Logger.Warnf("[移动端权限验证] ❌ 权限不匹配 - 用户: %s, 接口: %s, 需要权限: %s",
					userInfo.UserName, r.URL.Path, permission)
				middleware.ResponseError(w, http.StatusForbidden, "没有访问权限")
				return
			}
		}

		// 将用户信息添加到上下文
		ctx := context.WithValue(r.Context(), UserInfoKey, userInfo)
		r = r.WithContext(ctx)

		// 如果需要记录操作日志，使用移动端专用的日志记录
		if logType == "o" || logType == "all" {
			LogMobileOperation(w, r, next, userInfo, module, startTime)
			return
		}

		// 执行后续处理函数
		next(w, r)
	}
}

// LogMobileOperation 记录移动端操作日志
func LogMobileOperation(w http.ResponseWriter, r *http.Request, next http.HandlerFunc, userInfo UserInfo, module string, startTime time.Time) {
	// 使用与PC端类似的日志记录逻辑，但可以添加移动端标识
	rw := &ResponseWriter{
		ResponseWriter: w,
		Body:           &bytes.Buffer{},
		Status:         http.StatusOK,
	}

	var requestBody []byte
	if r.Body != nil {
		requestBody, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	next.ServeHTTP(rw, r)

	action := ParseModuleAndAction(r.URL.Path)

	log := system.OperationLog{
		UserID:    userInfo.UserID,
		UserName:  userInfo.UserName,
		Action:    action,
		Module:    module,
		Method:    r.Method,
		URL:       r.URL.String(),
		IP:        GetClientIP(r),
		Request:   FilterSensitiveInfo(string(requestBody)),
		Response:  FilterSensitiveInfo(rw.Body.String()),
		Status:    rw.Status,
		Duration:  int(time.Since(startTime).Milliseconds()),
		CreatedAt: time.Now(),
	}

	go SaveOperationLog(&log)
}

// MobileTokenKey 获取移动端token存储的Redis键名
// 已废弃，请使用middleware.GetMobileUserTokenKey
func MobileTokenKey(userID int64) string {
	return middleware.GetMobileUserTokenKey(userID)
}
