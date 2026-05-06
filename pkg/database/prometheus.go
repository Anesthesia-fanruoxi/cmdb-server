package database

import (
	"cmdb/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Prometheus Prometheus客户端全局变量
var Prometheus *http.Client

// PrometheusBaseURL 存储基础URL (用于指标查询)
var PrometheusBaseURL string

// PrometheusAlertsURL 存储告警URL (用于告警查询，必须指向Prometheus)
var PrometheusAlertsURL string

// PrometheusResponse Prometheus API响应的基本结构
type PrometheusResponse struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType string          `json:"errorType,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// InstantQueryResult 即时查询结果
type InstantQueryResult struct {
	ResultType string             `json:"resultType"`
	Result     []PrometheusSeries `json:"result"`
}

// RangeQueryResult 范围查询结果
type RangeQueryResult struct {
	ResultType string                  `json:"resultType"`
	Result     []PrometheusRangeSeries `json:"result"`
}

// PrometheusSeries Prometheus 即时查询序列数据
type PrometheusSeries struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"` // [timestamp, value]
}

// PrometheusRangeSeries Prometheus 范围查询序列数据
type PrometheusRangeSeries struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"` // [[timestamp, value], [timestamp, value], ...]
}

// InitPrometheus 初始化Prometheus客户端
func InitPrometheus(cfg *config.PrometheusConfig) error {
	// 保存基础URL (用于指标查询)
	PrometheusBaseURL = cfg.BaseURL

	// 保存告警URL (用于告警查询)
	if cfg.AlertsURL != "" {
		PrometheusAlertsURL = cfg.AlertsURL
	} else {
		PrometheusAlertsURL = cfg.BaseURL // 默认使用base_url
	}

	// 设置超时时间
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	// 创建HTTP客户端
	Prometheus = &http.Client{
		Timeout: timeout,
	}

	return nil
}

// GetAlertsURL 获取告警查询URL
func GetAlertsURL() string {
	return PrometheusAlertsURL
}

// QueryInstant 执行即时查询
func QueryInstant(ctx context.Context, query string, t time.Time) (*InstantQueryResult, error) {
	// 构建URL
	u, err := url.Parse(PrometheusBaseURL + "/api/v1/query")
	if err != nil {
		return nil, fmt.Errorf("解析Prometheus URL失败: %v", err)
	}

	// 添加查询参数
	q := u.Query()
	q.Set("query", query)
	if !t.IsZero() {
		q.Set("time", fmt.Sprintf("%d", t.Unix()))
	}
	u.RawQuery = q.Encode()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 执行请求
	resp, err := Prometheus.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if promResp.Status != "success" {
		return nil, fmt.Errorf("Prometheus查询失败: %s - %s", promResp.ErrorType, promResp.Error)
	}

	// 解析数据
	var result InstantQueryResult
	if err := json.Unmarshal(promResp.Data, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %v", err)
	}

	return &result, nil
}

// QueryRange 执行范围查询
func QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*RangeQueryResult, error) {
	// 构建URL
	u, err := url.Parse(PrometheusBaseURL + "/api/v1/query_range")
	if err != nil {
		return nil, fmt.Errorf("解析Prometheus URL失败: %v", err)
	}

	// 添加查询参数
	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	q.Set("step", fmt.Sprintf("%ds", int(step.Seconds())))
	u.RawQuery = q.Encode()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 执行请求
	resp, err := Prometheus.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if promResp.Status != "success" {
		return nil, fmt.Errorf("Prometheus查询失败: %s - %s", promResp.ErrorType, promResp.Error)
	}

	// 解析数据
	var result RangeQueryResult
	if err := json.Unmarshal(promResp.Data, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %v", err)
	}

	return &result, nil
}

// QueryInstantWithMilliseconds 使用毫秒时间戳执行Prometheus即时查询
func QueryInstantWithMilliseconds(ctx context.Context, query string, timeMillis int64) (*InstantQueryResult, error) {
	// 如果时间戳有效，转换为time.Time
	var t time.Time
	if timeMillis > 0 {
		// 将毫秒转换为秒和纳秒
		sec := timeMillis / 1000
		nsec := (timeMillis % 1000) * 1000000
		t = time.Unix(sec, nsec)
	}

	// 调用原始函数
	return QueryInstant(ctx, query, t)
}

// QueryRangeWithMilliseconds 使用毫秒时间戳执行Prometheus范围查询
func QueryRangeWithMilliseconds(ctx context.Context, query string, startMillis, endMillis, stepMillis int64) (*RangeQueryResult, error) {
	// 将毫秒时间戳转换为time.Time
	startTime := time.Unix(startMillis/1000, (startMillis%1000)*1000000)
	endTime := time.Unix(endMillis/1000, (endMillis%1000)*1000000)
	step := time.Duration(stepMillis) * time.Millisecond

	// 调用原始函数
	return QueryRange(ctx, query, startTime, endTime, step)
}
