package utils

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

var (
	officialSearcher *xdb.Searcher
	officialOnce     sync.Once
	officialMutex    sync.RWMutex // 添加读写锁保护并发访问

	// IP定位缓存，减少重复查询
	ipLocationCache = make(map[string]string)
	cacheMutex      sync.RWMutex
	maxCacheSize    = 100 // 最大缓存100个IP
)

// initOfficialSearcher 使用官方XDB搜索器
func initOfficialSearcher() *xdb.Searcher {
	officialOnce.Do(func() {
		var err error
		// 使用官方XDB搜索器实现
		officialSearcher, err = xdb.NewWithFileOnly("public/ip2region.xdb")
		if err != nil {
			// XDB文件加载失败，静默处理
			officialSearcher = nil
		}
	})
	return officialSearcher
}

// formatXDBResult 格式化XDB查询结果
func formatXDBResult(region string) string {
	if region == "" {
		return "未知位置"
	}

	// XDB返回格式：国家|省份|城市|运营商
	parts := strings.Split(region, "|")
	var result []string

	// 按照要求的顺序：国家 省份 城市 区县/运营商
	if len(parts) >= 1 && parts[0] != "" && parts[0] != "0" {
		result = append(result, parts[0]) // 国家
	}
	if len(parts) >= 2 && parts[1] != "" && parts[1] != "0" {
		result = append(result, parts[1]) // 省份
	}
	if len(parts) >= 3 && parts[2] != "" && parts[2] != "0" {
		result = append(result, parts[2]) // 城市
	}
	if len(parts) >= 4 && parts[3] != "" && parts[3] != "0" {
		result = append(result, parts[3]) // 区县/运营商
	}

	if len(result) == 0 {
		return "未知位置"
	}

	return strings.Join(result, " ")
}

// ipStringToUint32 将IP字符串转换为uint32
func ipStringToUint32(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("无效的IP地址: %s", ipStr)
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("不支持IPv6: %s", ipStr)
	}

	// 将4个字节组合成uint32 (大端序)
	return uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3]), nil
}

// cacheResult 将结果存入缓存
func cacheResult(ipStr, result string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 缓存满了，清理一半
	if len(ipLocationCache) >= maxCacheSize {
		count := 0
		for k := range ipLocationCache {
			delete(ipLocationCache, k)
			count++
			if count >= maxCacheSize/2 {
				break
			}
		}
	}

	ipLocationCache[ipStr] = result
}

// GetClientIP 获取客户端IP
func GetClientIP(r *http.Request) string {
	// 尝试从X-Real-IP获取
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 尝试从X-Forwarded-For获取
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 直接从RemoteAddr获取
	ip = r.RemoteAddr
	if ip != "" {
		ips := strings.Split(ip, ":")
		if len(ips) > 0 {
			return ips[0]
		}
	}

	return "unknown"
}

// GetIPLocation 获取 IP 地址的地理位置信息
func GetIPLocation(ipStr string) string {
	// 解析 IP 地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "无效IP"
	}

	// 检查是否为内网IP
	if isPrivateIP(ip) {
		return "内网IP"
	}

	// 先检查缓存
	cacheMutex.RLock()
	if cached, exists := ipLocationCache[ipStr]; exists {
		cacheMutex.RUnlock()
		return cached
	}
	cacheMutex.RUnlock()

	// 获取官方XDB搜索器实例
	searcher := initOfficialSearcher()
	if searcher == nil {
		return "未知位置"
	}

	// 使用读锁保护并发访问
	officialMutex.RLock()
	defer officialMutex.RUnlock()

	// 查询地理位置，添加错误恢复
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = "查询异常"
			}
		}()

		// 将IP字符串转换为uint32
		ipUint32, parseErr := ipStringToUint32(ipStr)
		if parseErr != nil {
			result = "无效IP"
			return
		}

		// 使用官方XDB搜索器查询
		region, err := searcher.Search(ipUint32)
		if err != nil {
			result = "未知位置"
			return
		}

		// 格式化结果
		result = formatXDBResult(region)
	}()

	// 如果查询失败，返回默认值
	if result == "" || result == "查询异常" {
		return "未知位置"
	}

	// 将结果存入缓存
	cacheResult(ipStr, result)

	return result
}

// isPrivateIP 检查是否为内网 IP
func isPrivateIP(ip net.IP) bool {
	// 检查 IPv4 私有地址范围
	if ip4 := ip.To4(); ip4 != nil {
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 127.0.0.0/8 (本地回环)
		if ip4[0] == 127 {
			return true
		}
	} else {
		// IPv6 本地回环
		if ip.Equal(net.ParseIP("::1")) {
			return true
		}
		// IPv6 唯一本地地址 (fc00::/7)
		if ip[0] == 0xfc || ip[0] == 0xfd {
			return true
		}
	}
	return false
}

// GetBrowser 获取浏览器类型
func GetBrowser(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "edge") {
		return "Edge"
	} else if strings.Contains(userAgent, "chrome") {
		return "Chrome"
	} else if strings.Contains(userAgent, "firefox") {
		return "Firefox"
	} else if strings.Contains(userAgent, "safari") {
		return "Safari"
	} else if strings.Contains(userAgent, "opera") {
		return "Opera"
	} else if strings.Contains(userAgent, "msie") || strings.Contains(userAgent, "trident") {
		return "IE"
	}

	return "Other"
}

// GetOS 获取操作系统
func GetOS(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "windows") {
		return "Windows"
	} else if strings.Contains(userAgent, "mac") {
		return "MacOS"
	} else if strings.Contains(userAgent, "linux") {
		return "Linux"
	} else if strings.Contains(userAgent, "android") {
		return "Android"
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "iOS"
	}

	return "Other"
}
func GetIPLocationDetails(ip string) (country, region, city, district string) {
	// 使用 utils.GetIPLocation 获取完整位置信息
	location := GetIPLocation(ip)

	// 分割位置信息（格式：国家 省份 城市 区县）
	parts := strings.Split(location, " ")

	switch len(parts) {
	case 4: // 完整信息：国家 省份 城市 区县
		return parts[0], parts[1], parts[2], parts[3]
	case 3: // 完整信息：国家 省份 城市
		return parts[0], parts[1], parts[2], ""
	case 2: // 部分信息：国家 省份
		return parts[0], parts[1], "", ""
	case 1: // 只有国家
		return parts[0], "", "", ""
	default: // 未知或错误格式
		return "未知", "", "", ""
	}
}
