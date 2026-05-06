package alert

import (
	"cmdb/models/monitor"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// startAlertPolling 启动告警轮询
func startAlertPolling(ctx context.Context) {
	middleware.Logger.Infof("告警轮询服务已启动，轮询间隔: %v", monitor.PollingInterval)

	// 创建定时器
	ticker := time.NewTicker(monitor.PollingInterval)
	defer ticker.Stop()

	// 立即执行一次
	pollAlerts(ctx)

	// 持续轮询
	for {
		select {
		case <-ticker.C:
			pollAlerts(ctx)
		case <-ctx.Done():
			middleware.Logger.Infof("告警轮询服务已停止")
			return
		}
	}
}

// pollAlerts 从Prometheus获取当前的告警
func pollAlerts(ctx context.Context) {
	// 构建API URL (使用AlertsURL，确保告警查询指向Prometheus)
	u, err := url.Parse(database.GetAlertsURL() + "/api/v1/alerts")
	if err != nil {
		middleware.Logger.Errorf("解析Prometheus URL失败: %v", err)
		return
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		middleware.Logger.Errorf("创建请求失败: %v", err)
		return
	}

	// 执行请求
	resp, err := database.Prometheus.Do(req)
	if err != nil {
		//middleware.Logger.Errorf("执行请求失败: %v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.Logger.Errorf("读取响应失败: %v", err)
		return
	}

	// 解析响应
	var alertResp monitor.AlertResponse
	if err := json.Unmarshal(body, &alertResp); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v", err)
		return
	}

	// 检查响应状态
	if alertResp.Status != "success" {
		middleware.Logger.Errorf("获取告警失败: 状态不是success")
		return
	}

	// 标准化告警数据（处理字段名称和值的规范化）
	for i := range alertResp.Data.Alerts {
		StandardizeAlert(&alertResp.Data.Alerts[i])
	}

	// 更新告警详情缓存
	updateAlertDetailsCache(alertResp.Data.Alerts)

	// 处理告警数据
	processAlerts(alertResp.Data.Alerts)
}

// InitAlert 初始化并启动告警系统
func InitAlert() {
	middleware.Logger.Infof("初始化告警系统...")

	// 创建一个后台上下文
	ctx := context.Background()

	// 启动轮询（后台运行）
	go startAlertPolling(ctx)

	middleware.Logger.Infof("告警系统初始化完成")
}
