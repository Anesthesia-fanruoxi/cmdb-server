package utils

import (
	"cmdb/config"
	"cmdb/initData"
	"context"
	"errors"
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

const TokenExpiration = 24 * time.Hour

// ClaimsKey 用于context的key
var ClaimsKey = struct{}{}

// 公开路由列表
var publicRoutes = []string{
	"/api/system/user/login",  // 登录接口
	"/api/system/user/logout", // 登出接口
	"/api/system/user/create", // 创建用户接口
}

// GenerateToken 生成JWT token并存入Redis
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
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
	err = initData.SetToken(userID, tokenString, TokenExpiration)
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

// JWTAuth JWT认证中间件
func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ��查是否是公开路由
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

		// 将用户信息添加到请求头中
		r.Header.Set("UserID", strconv.FormatUint(uint64(claims.UserID), 10))
		r.Header.Set("Username", claims.Username)

		// 将claims添加到context中
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)

		next(w, r)
	}
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
