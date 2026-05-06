package user

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/pkg/security"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// IPRateLimitRequest 频率限制操作请求
type IPRateLimitRequest struct {
	IP string `json:"ip" validate:"required,ip"`
}

// IPRateLimitResetRequest 重置频率限制请求
type IPRateLimitResetRequest struct {
	IP string `json:"ip" validate:"required,ip"`
}

// IPRateLimitListResponse 频率限制列表响应
type IPRateLimitListResponse struct {
	IP            string `json:"ip"`
	LastAccess    int64  `json:"last_access"`
	LastAccessStr string `json:"last_access_str"`
}

// GetIPRateLimits 获取IP频率限制列表
func GetIPRateLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从Redis中获取所有频率限制记录
	ctx := r.Context()
	pattern := security.IPRateLimitKeyPrefix + "*"
	keys, err := database.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取频率限制列表失败: %v", err))
		return
	}

	var ratelimits []IPRateLimitListResponse
	for _, key := range keys {
		// 从key中提取IP地址
		ip := strings.TrimPrefix(key, security.IPRateLimitKeyPrefix)

		// 获取上次访问时间
		lastAccessStr, err := database.RDB.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		lastAccess, _ := strconv.ParseInt(lastAccessStr, 10, 64)
		lastAccessTime := time.Unix(lastAccess, 0)

		ratelimits = append(ratelimits, IPRateLimitListResponse{
			IP:            ip,
			LastAccess:    lastAccess,
			LastAccessStr: lastAccessTime.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.ResponseSuccess(w, "获取IP频率限制列表成功", ratelimits)
}

// ResetIPRateLimit 重置IP频率限制
func ResetIPRateLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req IPRateLimitResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 重置IP频率限制
	ctx := r.Context()
	key := security.IPRateLimitKeyPrefix + req.IP
	if err := database.RDB.Del(ctx, key).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("重置IP频率限制失败: %v", err))
		return
	}

	middleware.ResponseSuccess(w, "重置IP频率限制成功", nil)
}
