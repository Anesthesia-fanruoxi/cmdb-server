package utils

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/routers/path"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TokenInfo 存储在Redis中的token信息
type TokenInfo struct {
	UserID        int64    `json:"user_id"`
	UserName      string   `json:"user_name"`
	RoleID        int64    `json:"role_id"`
	DeptID        int64    `json:"dept_id"`
	Permissions   []string `json:"permissions"`
	IsDefaultPass bool     `json:"is_default_pass"`
	DeviceType    string   `json:"device_type"`
	CreatedAt     int64    `json:"created_at"`
}

// WhiteList 定义不需要验证token的路径
var WhiteList = map[string]bool{
	path.UserLoginPath:                    true, // 登录接口不需要token
	path.PublicAnonymousPath:              true,
	path.UserFactorVerifyWithUsernamePath: true,
	path.PasswdSettingPath:                true,
	path.SqlAnalyzePath:                   true,
	path.SqlApplyListPath:                 true,
	path.SqlExportListPath:                true, // SSE接口，内部自己验证token
	path.TaskListPath:                     true, // 任务中心SSE接口，内部验证token
	path.TaskDetailPath:                   true, // 任务详情，内部验证token
	path.DeviceChallengePath:              true, // 获取挑战码无需token
	path.DeviceLoginPath:                  true, // 设备登录无需token
	path.AppUploadPath:                    true, // 上传安装包无需token
	path.AppVersionGetPath:                true, // 获取版本信息无需token
	path.AppVersionUpdatePath:             true, // 更新版本信息无需token
	path.TaskDownloadPath:                 true, // 下载文件无需token
	path.SqlExportDownloadPath:            true, // 导出下载（支持code参数，无需token）
	path.MonitorMetricsListSSEPath:        true, // 监控查询SSE接口，内部验证token
	path.DataBITablesPath:                 true, // DataBI SSE接口，内部验证token
}

// UserInfoKey 用户信息上下文键
type userInfoKey struct{}

var UserInfoKey = userInfoKey{}

// UserInfo 用户信息
type UserInfo struct {
	UserID        int64    `json:"user_id"`
	UserName      string   `json:"user_name"`
	NickName      string   `json:"nick_name"`
	RoleID        int64    `json:"role_id"`
	DeptID        int64    `json:"dept_id"`
	OnlineAssets  bool     `json:"online_assets"`
	TestAssets    bool     `json:"test_assets"`
	Permissions   []string `json:"permissions"`
	IsDefaultPass bool     `json:"is_default_pass"` // 是否是默认密码
}

// GetTokenExpireTime 获取token过期时间
func GetTokenExpireTime() (int, error) {
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		return 0, fmt.Errorf("获取系统设置失败: %v", err)
	}
	return setting.TokenExpireTime, nil
}

// GenerateRandomToken 生成随机token
func GenerateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken 生成token
func GenerateToken(userID interface{}, userName string, args ...interface{}) (string, error) {
	// 获取过期时间
	expireHours, err := GetTokenExpireTime()
	if err != nil {
		return "", err
	}

	// 将userID转换为int64
	userIDInt := middleware.ToInt64(userID)

	// 生成随机token
	token, err := GenerateRandomToken(128) // 生成128字节的随机token
	if err != nil {
		return "", fmt.Errorf("生成随机token失败: %v", err)
	}

	// 创建token信息
	tokenInfo := TokenInfo{
		UserID:     userIDInt,
		UserName:   userName,
		DeviceType: "pc",
		CreatedAt:  time.Now().Unix(),
	}

	// 兼容旧的调用方式
	if len(args) >= 5 {
		tokenInfo.RoleID = middleware.ToInt64(args[1])
		tokenInfo.DeptID = middleware.ToInt64(args[2])

		// 处理权限字符串
		if permStr, ok := args[3].(string); ok && permStr != "" {
			tokenInfo.Permissions = strings.Split(permStr, ",")
		}

		if isDefaultPass, ok := args[4].(bool); ok {
			tokenInfo.IsDefaultPass = isDefaultPass
		}
	}

	// 存储token信息
	if err := StoreTokenInfo(userIDInt, token, tokenInfo, expireHours); err != nil {
		return "", err
	}

	return token, nil
}

// GetUserInfo 根据用户ID查询完整的用户信息
func GetUserInfo(userID int64) (UserInfo, error) {
	var user system.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return UserInfo{}, fmt.Errorf("查询用户信息失败: %v", err)
	}

	// 获取用户权限
	permissions, err := GetRolePermissions(uint(user.RoleID))
	if err != nil {
		return UserInfo{}, fmt.Errorf("获取用户权限失败: %v", err)
	}

	return UserInfo{
		UserID:        userID,
		UserName:      user.UserName,
		NickName:      user.NickName,
		RoleID:        user.RoleID,
		DeptID:        user.DeptID,
		OnlineAssets:  user.OnlineAssets,
		TestAssets:    user.TestAssets,
		Permissions:   permissions,
		IsDefaultPass: user.IsDefaultPass,
	}, nil
}

// ParseToken 解析token并获取完整用户信息
func ParseToken(tokenString string, _ string) (UserInfo, error) {
	// 根据token获取用户信息
	tokenInfo, err := GetTokenInfo(tokenString)
	if err != nil {
		return UserInfo{}, err
	}

	// 根据用户ID获取完整的用户信息
	return GetUserInfo(tokenInfo.UserID)
}

// GetTokenInfo 根据token从Redis获取用户信息
func GetTokenInfo(token string) (TokenInfo, error) {
	ctx := context.Background()
	key := middleware.GetTokenInfoKey(token)

	// 从Redis获取token信息
	infoJSON, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		return TokenInfo{}, fmt.Errorf("token已失效,请重新登录")
	}

	// 解析token信息
	var tokenInfo TokenInfo
	if err := json.Unmarshal([]byte(infoJSON), &tokenInfo); err != nil {
		return TokenInfo{}, fmt.Errorf("解析token信息失败: %v", err)
	}

	return tokenInfo, nil
}

// RefreshTokenExpiration 刷新token的过期时间
func RefreshTokenExpiration(userID int64, tokenString string) error {
	expireHours, err := GetTokenExpireTime()
	if err != nil {
		return err
	}
	expireTime := time.Hour * time.Duration(expireHours)

	ctx := context.Background()
	tokenKey := middleware.GetTokenInfoKey(tokenString)
	userKey := middleware.GetUserTokenKey(userID)

	// 获取当前存储的token信息
	infoJSON, err := database.RDB.Get(ctx, tokenKey).Result()
	if err != nil {
		return fmt.Errorf("token不存在: %v", err)
	}

	// 刷新token信息的过期时间
	if err := database.RDB.Set(ctx, tokenKey, infoJSON, expireTime).Err(); err != nil {
		return err
	}

	// 检查用户token映射是否存在，不存在则重新创建
	exists, _ := database.RDB.Exists(ctx, userKey).Result()
	if exists == 0 {
		// 用户token映射丢失，重新创建
		if err := database.RDB.Set(ctx, userKey, tokenString, expireTime).Err(); err != nil {
			return err
		}
	} else {
		// 刷新用户token映射的过期时间
		if err := database.RDB.Expire(ctx, userKey, expireTime).Err(); err != nil {
			return err
		}
	}

	return nil
}

// StoreTokenInfo 在Redis中存储token信息并设置过期时间
func StoreTokenInfo(userID int64, token string, tokenInfo TokenInfo, expireHours int) error {
	expireTime := time.Hour * time.Duration(expireHours)
	ctx := context.Background()

	// 将token信息转为JSON
	infoJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("序列化token信息失败: %v", err)
	}

	// 存储token信息与用户ID的映射
	tokenKey := middleware.GetTokenInfoKey(token)
	if err := database.RDB.Set(ctx, tokenKey, string(infoJSON), expireTime).Err(); err != nil {
		return err
	}

	// 存储用户ID与token的映射，方便管理同一用户的多个token
	userKey := middleware.GetUserTokenKey(userID)
	return database.RDB.Set(ctx, userKey, token, expireTime).Err()
}

// StoreToken 在Redis中存储token并设置过期时间 (保留兼容性)
func StoreToken(userID int64, tokenString string) error {
	// 获取用户信息
	user, err := GetUserInfo(userID)
	if err != nil {
		return err
	}

	// 获取token过期时间
	expireHours, err := GetTokenExpireTime()
	if err != nil {
		return err
	}

	// 创建token信息
	tokenInfo := TokenInfo{
		UserID:        userID,
		UserName:      user.UserName,
		RoleID:        user.RoleID,
		DeptID:        user.DeptID,
		Permissions:   user.Permissions,
		IsDefaultPass: user.IsDefaultPass,
		DeviceType:    "pc",
		CreatedAt:     time.Now().Unix(),
	}

	// 存储token信息
	return StoreTokenInfo(userID, tokenString, tokenInfo, expireHours)
}

// InvalidateToken 使token失效
func InvalidateToken(userID int64) error {
	ctx := context.Background()

	// 获取用户的token
	userKey := middleware.GetUserTokenKey(userID)
	token, err := database.RDB.Get(ctx, userKey).Result()
	if err != nil {
		return nil // 如果token已不存在，不返回错误
	}

	// 删除token信息
	tokenKey := middleware.GetTokenInfoKey(token)
	if err := database.RDB.Del(ctx, tokenKey).Err(); err != nil {
		return err
	}

	// 删除用户与token的映射
	return database.RDB.Del(ctx, userKey).Err()
}

// JWTAuth 认证中间件 (保留名称以兼容现有代码)
func JWTAuth(next http.HandlerFunc, permission string, logType string, module string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取请求开始时间
		startTime := time.Now()

		// 记录标准日志（控制台输出）
		if logType == "s" || logType == "all" {
			middleware.Logger.Infof("[API请求] %s %s", r.Method, r.URL.Path)
		}

		// 处理跨域预检请求
		if r.Method == "OPTIONS" {
			HandleCORS(w, r)
			return
		}

		// 设置跨域响应头
		SetCORSHeaders(w)

		// 检查是否在白名单中
		if WhiteList[r.URL.Path] {
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

		// 去掉Bearer前缀
		tokenString := strings.TrimPrefix(token, "Bearer ")

		// 解析token获取用户信息
		jwtConfig := config.GetConfig().JWT
		userInfo, err := ParseToken(tokenString, jwtConfig.Secret)
		if err != nil {
			middleware.ResponseError(w, http.StatusUnauthorized, err.Error())
			return
		}

		// 刷新token过期时间
		if err := RefreshTokenExpiration(userInfo.UserID, tokenString); err != nil {
			middleware.Logger.Errorf("刷新token过期时间失败: %v", err)
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
			middleware.Logger.Infof("[权限验证] 接口: %s, 需要权限: %s",
				r.URL.Path, permission)

			if hasPermission(userInfo.Permissions, permission) {
				middleware.Logger.Infof("[权限验证] ✅ 权限匹配 - 用户: %s, 接口: %s",
					userInfo.UserName, r.URL.Path)
			} else {
				middleware.Logger.Warnf("[权限验证] ❌ 权限不匹配 - 用户: %s, 接口: %s, 需要权限: %s",
					userInfo.UserName, r.URL.Path, permission)
				middleware.ResponseError(w, http.StatusForbidden, "没有访问权限")
				return
			}
		}

		// 将用户信息添加到上下文
		ctx := context.WithValue(r.Context(), UserInfoKey, userInfo)
		r = r.WithContext(ctx)

		// 如果需要记录操作日志，使用审计中间件包装处理函数
		if logType == "o" || logType == "all" {
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
			return
		}

		next(w, r)
	}
}

// HandleCORS 处理跨域预检请求
func HandleCORS(w http.ResponseWriter, _ *http.Request) {
	SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// SetCORSHeaders 设置跨域响应头
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// CacheRolePermissions 缓存角色权限
func CacheRolePermissions(roleID uint, permissions []string) error {
	ctx := context.Background()
	key := middleware.GetRolePermissionsKey(roleID)

	permJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("解析权限失败: %v", err)
	}

	// 永久存储角色权限，只在权限变更时更新
	err = database.RDB.Set(ctx, key, string(permJSON), 0).Err()
	if err != nil {
		return fmt.Errorf("缓存权限失败: %v", err)
	}

	return nil
}

// GetRolePermissions 从缓存获取角色权限
func GetRolePermissions(roleID uint) ([]string, error) {
	ctx := context.Background()
	key := middleware.GetRolePermissionsKey(roleID)

	permJSON, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("从缓存获取权限失败: %v", err)
	}

	var permissions []string
	if err := json.Unmarshal([]byte(permJSON), &permissions); err != nil {
		return nil, fmt.Errorf("解析权限失败: %v", err)
	}

	return permissions, nil
}

// hasPermission 检查是否有指定权限
func hasPermission(permissions []string, requiredPermission string) bool {
	if requiredPermission == "" {
		return true
	}

	for _, permission := range permissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

// RefreshRolePermissions 刷新角色权限缓存
func RefreshRolePermissions(roleID uint) error {
	type MenuPermission struct {
		Permission string
		HasRead    bool
		HasWrite   bool
	}

	var menuPerms []MenuPermission
	if err := database.DB.Table(middleware.TableMenus).
		Select(middleware.TableMenus+".permission, "+middleware.TableRoleMenus+".has_read, "+middleware.TableRoleMenus+".has_write").
		Joins("INNER JOIN "+middleware.TableRoleMenus+" ON "+middleware.TableRoleMenus+".menu_id = "+middleware.TableMenus+".id").
		Where(middleware.TableRoleMenus+".role_id = ?", roleID).
		Find(&menuPerms).Error; err != nil {
		return fmt.Errorf("查询权限失败: %v", err)
	}

	permissions := make([]string, 0)
	for _, mp := range menuPerms {
		if mp.HasRead {
			permissions = append(permissions, mp.Permission+":r")
		}
		if mp.HasWrite {
			permissions = append(permissions, mp.Permission+":w")
		}
	}

	if err := CacheRolePermissions(roleID, permissions); err != nil {
		return fmt.Errorf("缓存权限失败: %v", err)
	}

	return nil
}

// InitRolePermissions 初始化所有角色的权限缓存
func InitRolePermissions() error {
	var roleIDs []int64
	if err := database.DB.Table(middleware.TableRoles).Select("id").Find(&roleIDs).Error; err != nil {
		return fmt.Errorf("查询角色失败: %v", err)
	}

	for _, roleID := range roleIDs {
		if err := RefreshRolePermissions(uint(roleID)); err != nil {
			middleware.Logger.Errorf("初始化角色(%d)权限失败: %v", roleID, err)
		}
	}

	return nil
}
