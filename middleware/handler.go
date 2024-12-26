package middleware

import (
	"cmdb/utils"
	"net/http"
	"strings"
)

// RequestHandler 统一的请求处理中间件
func RequestHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理OPTIONS请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 检查是否在白名单中
		if CheckWhiteList(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Token验证
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

		// 将用户信息存储在请求上下文中
		r.Header.Set("UserID", string(claims.UserID))
		r.Header.Set("Username", claims.Username)

		// 继续处理请求
		next.ServeHTTP(w, r)
	}
}
