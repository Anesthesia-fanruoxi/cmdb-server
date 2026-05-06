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

// IPBlacklistRequest 黑名单操作请求
type IPBlacklistRequest struct {
	IP string `json:"ip" validate:"required,ip"`
}

// IPBlacklistListResponse 黑名单列表响应
type IPBlacklistListResponse struct {
	IP        string `json:"ip"`
	ExpireAt  int64  `json:"expire_at"`
	Remaining int    `json:"remaining_hours"`
}

// IPFailedCountResponse 失败计数响应
type IPFailedCountResponse struct {
	IP           string `json:"ip"`
	FailedCount  int    `json:"failed_count"`
	FailedLimit  int    `json:"failed_limit"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresAtStr string `json:"expires_at_str"`
}

// GetIPBlacklist 获取IP黑名单列表
func GetIPBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从Redis中获取所有黑名单
	ctx := r.Context()
	// 使用变量构建模式
	pattern := fmt.Sprintf(middleware.KeyPathIPBlacklist, "*")
	keys, err := database.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取黑名单失败: %v", err))
		return
	}

	var blacklist []IPBlacklistListResponse
	for _, key := range keys {
		// 从key中提取IP地址，使用变量前缀
		prefix := fmt.Sprintf(middleware.KeyPathIPBlacklist, "")
		prefix = strings.TrimSuffix(prefix, "%s")
		ip := strings.TrimPrefix(key, prefix)

		// 获取TTL
		ttl, err := database.RDB.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		remainingHours := int(ttl.Hours())
		expireAt := int64(0)
		if ttl.Seconds() > 0 {
			expireAt = time.Now().Add(ttl).Unix()
		}

		blacklist = append(blacklist, IPBlacklistListResponse{
			IP:        ip,
			ExpireAt:  expireAt,
			Remaining: remainingHours,
		})
	}

	middleware.ResponseSuccess(w, "获取IP黑名单成功", blacklist)
}

// GetIPFailedCounts 获取当前IP失败计数列表
func GetIPFailedCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取安全配置
	config, err := security.GetConfig(r.Context())
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取安全配置失败: %v", err))
		return
	}

	// 从Redis中获取所有失败计数
	ctx := r.Context()
	// 使用变量构建模式
	pattern := fmt.Sprintf(middleware.KeyPathIPLoginFailed, "*")
	keys, err := database.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取失败计数列表失败: %v", err))
		return
	}

	var failedCounts []IPFailedCountResponse
	for _, key := range keys {
		// 从key中提取IP地址，使用变量前缀
		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			ip := parts[3] // 获取IP部分

			// 获取失败次数
			countStr, err := database.RDB.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			count, _ := strconv.Atoi(countStr)

			// 获取TTL
			ttl, err := database.RDB.TTL(ctx, key).Result()
			if err != nil {
				continue
			}

			expiresAt := int64(0)
			expiresAtStr := ""
			if ttl.Seconds() > 0 {
				expiresTime := time.Now().Add(ttl)
				expiresAt = expiresTime.Unix()
				expiresAtStr = expiresTime.Format("2006-01-02 15:04:05")
			}

			failedCounts = append(failedCounts, IPFailedCountResponse{
				IP:           ip,
				FailedCount:  count,
				FailedLimit:  config.BlacklistCount,
				ExpiresAt:    expiresAt,
				ExpiresAtStr: expiresAtStr,
			})
		}
	}

	middleware.ResponseSuccess(w, "获取IP失败计数列表成功", failedCounts)
}

// RemoveIPFromBlacklist 从黑名单中移除IP
func RemoveIPFromBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req IPBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 从黑名单中移除IP
	if err := security.RemoveIPFromBlacklist(req.IP); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("移除IP失败: %v", err))
		return
	}

	middleware.ResponseSuccess(w, "移除IP成功", nil)
}

// ResetIPFailedCount 重置IP失败计数
func ResetIPFailedCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req IPBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 重置IP失败计数
	if err := security.ResetIPLoginFailed(req.IP); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("重置IP失败计数失败: %v", err))
		return
	}

	middleware.ResponseSuccess(w, "重置IP失败计数成功", nil)
}

// AddIPToBlacklist 手动将IP添加到黑名单
func AddIPToBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req IPBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 获取安全配置
	config, err := security.GetConfig(r.Context())
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取安全配置失败: %v", err))
		return
	}

	// 将IP添加到黑名单
	ctx := r.Context()
	key := middleware.GetIPBlacklistKey(req.IP)
	if err := database.RDB.Set(ctx, key, "blocked", time.Duration(config.BlacklistLock)*time.Minute).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("添加IP失败: %v", err))
		return
	}

	middleware.ResponseSuccess(w, "添加IP到黑名单成功", nil)
}
