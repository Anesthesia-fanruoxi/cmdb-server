package utils

import (
	"cmdb/pkg/middleware"
	"net"
	"net/http"
	"strings"
)

// WhitelistConfig IP白名单配置
type WhitelistConfig struct {
	AllowedIPs []string `json:"allowed_ips"` // 允许的IP地址列表
}

// DefaultWhitelistConfig 默认白名单配置
var DefaultWhitelistConfig = WhitelistConfig{
	AllowedIPs: []string{
		"220.243.131.172",
		"220.243.131.173",
		"101.126.59.7",
		"101.126.59.8",
		"101.126.59.9",
	},
}

// IsIPInWhitelist 检查IP是否在白名单中
func IsIPInWhitelist(clientIP string, whitelist []string) bool {
	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		middleware.Logger.Warnf("无法解析客户端IP: %s", clientIP)
		return false
	}

	for _, allowedIP := range whitelist {
		// 检查是否是CIDR格式
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err != nil {
				middleware.Logger.Warnf("无法解析CIDR: %s", allowedIP)
				continue
			}
			if ipNet.Contains(clientIPAddr) {
				return true
			}
		} else {
			// 单个IP地址比较
			allowedIPAddr := net.ParseIP(allowedIP)
			if allowedIPAddr != nil && allowedIPAddr.Equal(clientIPAddr) {
				return true
			}
		}
	}
	return false
}

// IPWhitelistMiddleware IP白名单验证中间件
func IPWhitelistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端IP
		clientIP := GetClientIP(r)

		middleware.Logger.Infof("IP白名单验证 - 客户端IP: %s", clientIP)

		// 检查IP是否在白名单中
		if !IsIPInWhitelist(clientIP, DefaultWhitelistConfig.AllowedIPs) {
			middleware.Logger.Warnf("IP白名单验证失败 - 客户端IP: %s 不在白名单中", clientIP)
			middleware.ResponseError(w, http.StatusForbidden, "IP地址不在白名单中")
			return
		}

		middleware.Logger.Infof("IP白名单验证成功 - 客户端IP: %s", clientIP)

		// 通过验证，继续处理请求
		next.ServeHTTP(w, r)
	})
}
