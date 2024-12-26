package middleware

import "net/http"

// CORS 处理跨域请求的中间件
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 允许所有来源
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// 允许的请求方法
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 允许的请求头
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With")
		// 允许客户端获取哪些头
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		// 允许缓存预检请求结果的时间（秒）
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
