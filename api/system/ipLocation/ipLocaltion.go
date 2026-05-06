package ipLocation

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"time"
)

// IPLocationResponse IP位置信息响应结构
type IPLocationResponse struct {
	IP       string `json:"ip"`       // IP地址
	Location string `json:"location"` // 完整位置信息
	Country  string `json:"country"`  // 国家
	Region   string `json:"region"`   // 省份/地区
	City     string `json:"city"`     // 城市
	District string `json:"district"` // 区县/运营商
}

// GetLocation 获取IP位置信息
func GetLocation(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 仅允许 GET 方法
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从 URL 查询参数中获取 ip
	queryIP := r.URL.Query().Get("ip")
	if queryIP == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少参数: ip")
		return
	}

	// 获取请求来源IP
	requestIP := utils.GetClientIP(r)

	// 获取完整位置信息
	location := utils.GetIPLocation(queryIP)

	// 检查是否为无效IP
	if location == "无效IP" {
		// 异步记录失败审计
		recordIPLocationAudit(AuditParams{
			RequestIP: requestIP,
			QueryIP:   queryIP,
			Status:    http.StatusBadRequest,
		})
		middleware.ResponseError(w, http.StatusBadRequest, "无效的IP地址")
		return
	}

	// 获取详细位置信息
	country, region, city, district := utils.GetIPLocationDetails(queryIP)

	// 构建响应数据
	response := IPLocationResponse{
		IP:       queryIP,
		Location: location,
		Country:  country,
		Region:   region,
		City:     city,
		District: district,
	}

	// 异步记录成功审计
	recordIPLocationAudit(AuditParams{
		RequestIP: requestIP,
		QueryIP:   queryIP,
		Status:    http.StatusOK,
	})

	// 记录日志
	middleware.Logger.Infof("IP位置查询: 请求IP=%s, 查询IP=%s, 结果=%s, 耗时=%dms",
		requestIP, queryIP, location, time.Since(startTime).Milliseconds())

	// 成功响应
	middleware.ResponseSuccess(w, "IP 位置查询成功", response)
}
