package middleware

import (
	"cmdb/utils"
	"context"
	"net/http"
	"strconv"
	"strings"
)

// 公开路由列表
var publicRoutes = []string{
	"/api/user/login",  // 登录接口
	"/api/user/logout", // 登出接口
	"/api/user/create", // 登出接口
}

// ClaimsKey 用于context的key
var ClaimsKey = struct{}{}

// JWTAuth JWT认证中间件
func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是公开路由
		if isPublicRoute(r.URL.Path) {
			next(w, r)
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

		// 解析token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, utils.UNAUTHORIZED, "无效的token")
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
