package utils

import (
	"cmdb/config"
	"cmdb/initData"
	"cmdb/types"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Claims 自定义的JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var cfg *config.Config
var jwtSecret []byte

// SetConfig 设置配置
func SetConfig(c *config.Config) {
	cfg = c
	jwtSecret = []byte(cfg.Security.JWTSecret)
}

// ClaimsKey 用于context的key
var ClaimsKey = struct{}{}

// 公开路由列表
var publicRoutes = []string{
	"/api/system/user/logout", // 登出接口
	"/api/system/user/create", // 创建用户接口
}

// GenerateToken 生成JWT token并存入Redis
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(types.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// 将token存入Redis
	err = initData.SetToken(userID, tokenString, types.TokenExpiration)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT token并验证是否在Redis中存在
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, errors.New("token格式错误或已过期")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 从Redis中获取token并验证
		storedToken, err := initData.GetToken(claims.UserID)
		if err != nil {
			return nil, errors.New("登录已过期，请重新登录")
		}
		if storedToken != tokenString {
			return nil, errors.New("账号已在其他地方登录，请重新登录")
		}
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// RefreshRolePermissions 刷新角色的权限到Redis
func RefreshRolePermissions(roleID uint) error {
	permissions, err := getRolePermissions(roleID)
	if err != nil {
		return err
	}

	// 将权限列表存入Redis
	key := fmt.Sprintf("role:%d:permissions", roleID)
	return initData.GetRedis().Set(context.Background(), key, strings.Join(permissions, ","), types.TokenExpiration).Err()
}

// getRolePermissions 获取角色的所有权限标识
func getRolePermissions(roleID uint) ([]string, error) {
	db := initData.GetDB()

	// 如果是超级管理员角色，返回所有权限
	if roleID == 1 { // 超级管理员
		var allPermissions []string
		err := db.Table("menus").
			Where("permission IS NOT NULL").
			Pluck("permission", &allPermissions).Error
		if err != nil {
			return nil, err
		}
		return allPermissions, nil
	}

	// 查询角色对应的权限
	var permissions []string
	err := db.Table("role_menus").
		Joins("JOIN menus ON menus.id = role_menus.menu_id").
		Where("role_menus.role_id = ? AND menus.permission IS NOT NULL AND menus.is_enabled = 1", roleID).
		Pluck("menus.permission", &permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetRolePermissions 从Redis获取角色权限
func GetRolePermissions(roleID uint) ([]string, error) {
	key := fmt.Sprintf("role:%d:permissions", roleID)
	result, err := initData.GetRedis().Get(context.Background(), key).Result()
	if err != nil {
		// 如果Redis中没有，重新获取并存储
		return refreshAndGetRolePermissions(roleID)
	}

	if result == "" {
		return []string{}, nil
	}

	return strings.Split(result, ","), nil
}

// refreshAndGetRolePermissions 刷新并获取角色权限
func refreshAndGetRolePermissions(roleID uint) ([]string, error) {
	permissions, err := getRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	// 更新到Redis
	err = RefreshRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetUserPermissions 获取用户的所有权限
func GetUserPermissions(userID uint) ([]string, error) {
	db := initData.GetDB()

	// 从users表中直接获取role_id
	var roleID uint
	err := db.Table("users").
		Select("role_id").
		Where("id = ?", userID).
		Scan(&roleID).Error
	if err != nil {
		return nil, err
	}

	// 获取角色的权限
	key := fmt.Sprintf("role:%d:permissions", roleID)
	result, err := initData.GetRedis().Get(context.Background(), key).Result()
	if err != nil {
		// 如果Redis中没有，重新获取并存储
		err = initData.RefreshRolePermissions(roleID)
		if err != nil {
			return nil, err
		}
		// 再次从Redis获取
		result, err = initData.GetRedis().Get(context.Background(), key).Result()
		if err != nil {
			return nil, err
		}
	}

	if result == "" {
		return []string{}, nil
	}

	return strings.Split(result, ","), nil
}

// JWTAuth JWT认证中间件
func JWTAuth(next http.HandlerFunc, permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是公开路由
		if isPublicRoute(r.URL.Path) {
			next(w, r)
			return
		}

		// 获取Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			Error(w, http.StatusUnauthorized, UNAUTHORIZED, "未提供认证token")
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			Error(w, http.StatusUnauthorized, UNAUTHORIZED, "token格式错误")
			return
		}

		// 解析token
		claims, err := ParseToken(parts[1])
		if err != nil {
			Error(w, http.StatusUnauthorized, UNAUTHORIZED, "无效的token")
			return
		}

		// 权限检查
		if permission != "" {
			permissions, err := GetUserPermissions(claims.UserID)
			if err != nil {
				Error(w, http.StatusInternalServerError, ERROR, "获取权限失败")
				return
			}
			if !hasPermission(permissions, permission) {
				Error(w, http.StatusForbidden, FORBIDDEN, "没有访问权限")
				return
			}
		}

		// 将用户信息添加到请求头中
		r.Header.Set("UserID", strconv.FormatUint(uint64(claims.UserID), 10))
		r.Header.Set("Username", claims.Username)

		// 将claims添加到context中
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// hasPermission 检查是否有指定权限
func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}

// isPublicRoute 检查是否是公开路由
func isPublicRoute(path string) bool {
	for _, route := range publicRoutes {
		if route == path {
			return true
		}
	}
	return false
}

// GetUserFromToken 从请求的 Authorization 头中解析用户信息
func GetUserFromToken(r *http.Request) (*Claims, error) {
	// 获取Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("未提供认证token")
	}

	// 检查Bearer前缀
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("token格式错误")
	}

	// 解析token
	claims, err := ParseToken(parts[1])
	if err != nil {
		return nil, errors.New("无效的token")
	}

	return claims, nil
}
