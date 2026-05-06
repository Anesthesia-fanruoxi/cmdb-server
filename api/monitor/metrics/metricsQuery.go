package metrics

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// InstantQueryRequest 即时查询请求结构
type InstantQueryRequest struct {
	Query string `json:"query"`
	Time  int64  `json:"time,omitempty"` // 毫秒时间戳
}

// RangeQueryRequest 范围查询请求结构
type RangeQueryRequest struct {
	Query string `json:"query"`
	Start int64  `json:"start"` // 毫秒时间戳
	End   int64  `json:"end"`   // 毫秒时间戳
	Step  int64  `json:"step"`  // 毫秒
}

// 计算自动步长 (单位：毫秒)
func CalculateAutoStep(startMillis, endMillis int64) int64 {
	// 计算时间范围 (毫秒)
	timeRange := endMillis - startMillis

	// 目标数据点数量 (理想情况下希望返回的数据点数)
	// 200-300个点通常是可视化的良好平衡点
	const targetDataPoints = 250

	// 初步计算步长
	stepMillis := timeRange / targetDataPoints

	// 将步长调整为"漂亮"的数值 (秒级别)
	// 可选值: 1s, 5s, 10s, 30s, 1m, 5m, 10m, 30m, 1h, 3h, 6h, 12h, 1d
	niceSteps := []int64{
		1000,     // 1秒
		5000,     // 5秒
		10000,    // 10秒
		30000,    // 30秒
		60000,    // 1分钟
		300000,   // 5分钟
		600000,   // 10分钟
		1800000,  // 30分钟
		3600000,  // 1小时
		10800000, // 3小时
		21600000, // 6小时
		43200000, // 12小时
		86400000, // 1天
	}

	// 找到最接近的"漂亮"步长
	var bestStep int64 = niceSteps[len(niceSteps)-1] // 默认使用最大步长
	for _, step := range niceSteps {
		if step >= stepMillis {
			bestStep = step
			break
		}
	}

	// 确保至少有几个数据点
	if timeRange/bestStep < 5 {
		// 如果数据点太少，回退到上一个较小的步长
		for i := len(niceSteps) - 2; i >= 0; i-- {
			if timeRange/niceSteps[i] >= 5 {
				bestStep = niceSteps[i]
				break
			}
		}
	}

	// 确保数据点不会过多 (防止查询过大)
	var maxDataPoints int64 = 300 // 设置一个上限
	if timeRange/bestStep > maxDataPoints {
		// 如果数据点太多，选择更大的步长
		for i := 0; i < len(niceSteps); i++ {
			if timeRange/niceSteps[i] <= maxDataPoints {
				bestStep = niceSteps[i]
				break
			}
		}
	}

	return bestStep
}

// MonitorMetricsQuery 处理Prometheus即时查询请求
func MonitorMetricsQuery(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 解析查询参数
	query := r.URL.Query().Get("query")
	timeStr := r.URL.Query().Get("time")

	// 验证请求参数
	if query == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query参数不能为空")
		return
	}

	// 处理时间参数（毫秒时间戳）
	var millisecondTimestamp int64
	if timeStr != "" {
		var err error
		millisecondTimestamp, err = strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, "time参数必须是毫秒时间戳")
			return
		}
	}

	// 执行查询
	result, err := database.QueryInstantWithMilliseconds(r.Context(), query, millisecondTimestamp)
	if err != nil {
		middleware.Logger.Errorf("执行Prometheus即时查询失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "执行查询失败: "+err.Error())
		return
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}

// MonitorMetricsRangeQuery 处理Prometheus范围查询请求
func MonitorMetricsRangeQuery(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 解析查询参数
	query := r.URL.Query().Get("query")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	stepStr := r.URL.Query().Get("step")

	// 验证请求参数
	if query == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query参数不能为空")
		return
	}

	if startStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "start参数不能为空")
		return
	}

	if endStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "end参数不能为空")
		return
	}

	// 解析开始时间（毫秒时间戳）
	startMillis, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "start参数必须是毫秒时间戳")
		return
	}

	// 解析结束时间（毫秒时间戳）
	endMillis, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "end参数必须是毫秒时间戳")
		return
	}

	// 处理步长参数
	var stepMillis int64
	if stepStr != "" {
		// 如果提供了步长参数，则使用提供的值
		stepMillis, err = strconv.ParseInt(stepStr, 10, 64)
		if err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, "step参数必须是毫秒值")
			return
		}
	} else {
		// 自动计算步长
		stepMillis = CalculateAutoStep(startMillis, endMillis)
	}

	// 执行查询
	result, err := database.QueryRangeWithMilliseconds(r.Context(), query, startMillis, endMillis, stepMillis)
	if err != nil {
		middleware.Logger.Errorf("执行Prometheus范围查询失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "执行查询失败: "+err.Error())
		return
	}

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", result)
}
