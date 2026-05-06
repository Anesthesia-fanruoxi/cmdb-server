package metrics

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MetricsListRequest 批量查询请求结构
type MetricsListRequest struct {
	Project   string `json:"project"`   // 项目标识
	Category  string `json:"category"`  // 视图分类
	NameSpace string `json:"namespace"` // 命名空间
	Service   string `json:"service"`   // 服务名称
	Start     string `json:"start"`     // 开始时间（"2006-01-02 15:04:05"格式）
	End       string `json:"end"`       // 结束时间（"2006-01-02 15:04:05"格式）
}

// MetricsListResponse 批量查询响应结构
type MetricsListResponse struct {
	ViewID      int         `json:"view_id"`         // 视图ID
	ViewName    string      `json:"view_name"`       // 视图名称
	Description string      `json:"description"`     // 视图描述
	Query       string      `json:"query"`           // 实际执行的查询语句
	Sort        int         `json:"sort"`            // 视图排序
	ChartType   string      `json:"chart_type"`      // 图表类型
	Standard    string      `json:"standard"`        // 标准
	Data        interface{} `json:"data"`            // 查询结果
	Error       string      `json:"error,omitempty"` // 查询错误信息，如果有
}

// MonitorMetricsList 批量查询监控指标
func MonitorMetricsList(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 解析请求体
	var req MetricsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数解析失败: "+err.Error())
		return
	}

	// 参数验证
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	if req.Category == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "category参数不能为空")
		return
	}

	// 解析结束时间（所有类型的图表都需要）
	var endTime time.Time
	var endMillis int64

	if req.End == "" {
		// 如果未提供结束时间，使用当前时间
		endTime = time.Now()
		endMillis = middleware.TimeToMilliseconds(endTime)
	} else {
		// 解析提供的结束时间
		const layout = "2006-01-02 15:04:05"
		var err error
		endTime, err = time.ParseInLocation(layout, req.End, time.Local)
		if err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, "end参数格式错误，应为："+layout+": "+err.Error())
			return
		}
		endMillis = middleware.TimeToMilliseconds(endTime)
	}

	// 查询项目对应的名称
	var projectName string
	if err := database.DB.Table(middleware.TableProjectDict).Where("project = ?", req.Project).Pluck("project_name", &projectName).Error; err != nil {
		middleware.Logger.Errorf("查询项目名称失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目名称失败: "+err.Error())
		return
	}

	// 查询所有监控视图
	type View struct {
		ID          int    `json:"id" gorm:"column:id"`
		ViewName    string `json:"view_name" gorm:"column:view_name"`
		Query       string `json:"query" gorm:"column:query"`
		Category    string `json:"category" gorm:"column:category"`
		Sort        int    `json:"sort" gorm:"column:sort"`
		Standard    string `json:"standard" gorm:"column:standard"`
		Description string `json:"description" gorm:"column:description"`
		ChartType   string `json:"chart_type" gorm:"column:chart_type"`
	}

	var views []View
	if err := database.DB.Table(middleware.TableMonitorViews).
		Select("id, view_name, query, category, sort, standard, description, chart_type").
		Where("category = ? ", req.Category).
		Find(&views).Error; err != nil {
		middleware.Logger.Errorf("查询监控视图失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询视图失败: "+err.Error())
		return
	}

	if len(views) == 0 {
		// 没有找到视图，返回空结果
		middleware.ResponseSuccess(w, "没有监控视图数据", []interface{}{})
		return
	}

	// 解析开始时间和计算步长（仅对line类型图表需要）
	var startMillis int64
	var stepMillis int64

	// 统计需要做范围查询的视图数
	var lineChartCount int
	for _, v := range views {
		if v.ChartType == "line" {
			lineChartCount++
		}
	}

	// 如果有line类型图表，则需要解析开始时间
	if lineChartCount > 0 {
		if req.Start == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "start参数不能为空（需要用于折线图）")
			return
		}

		// 解析开始时间
		const layout = "2006-01-02 15:04:05"
		startTime, err := time.ParseInLocation(layout, req.Start, time.Local)
		if err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, "start参数格式错误，应为："+layout+": "+err.Error())
			return
		}

		startMillis = middleware.TimeToMilliseconds(startTime)

		if endMillis <= startMillis {
			middleware.ResponseError(w, http.StatusBadRequest, "结束时间必须大于开始时间")
			return
		}

		// 计算查询步长
		stepMillis = CalculateAutoStep(startMillis, endMillis)
	}

	// 使用WaitGroup和Mutex保证并发安全
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]MetricsListResponse, 0, len(views))

	// 为每个视图启动一个goroutine进行查询
	for _, view := range views {
		wg.Add(1)
		go func(v View) {
			defer wg.Done()

			// 获取原始查询语句
			queryWithLabels := v.Query

			// 检查查询中是否包含变量（不区分大小写）
			queryLower := strings.ToLower(queryWithLabels)
			needProjectVar := strings.Contains(queryLower, "$project")
			needNamespaceVar := strings.Contains(queryLower, "$namespace")
			needServiceVar := strings.Contains(queryLower, "$service")

			// namespace处理：只有当查询中包含 $namespace 变量时才处理
			var namespaceValue string
			if needNamespaceVar {
				// 查询包含 $namespace 变量，需要处理 namespace
				namespaceValue = req.NameSpace
				if namespaceValue == "" {
					// 没有传递namespace，自动推导
					namespaceValue = req.Project + "-service"
					middleware.Logger.Infof("查询包含$namespace变量但未传递，自动推导: %s", namespaceValue)
				} else {
					// 使用传递的namespace
					middleware.Logger.Infof("使用传递的namespace: %s", namespaceValue)
				}
			}

			// 替换变量
			if needProjectVar && projectName != "" {
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "$project", projectName)
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "'$project'", "'"+projectName+"'")
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "\"$project\"", "\""+projectName+"\"")
			}

			if needNamespaceVar && namespaceValue != "" {
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "$namespace", namespaceValue)
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "'$namespace'", "'"+namespaceValue+"'")
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "\"$namespace\"", "\""+namespaceValue+"\"")
			}

			// 替换 $service 变量
			if needServiceVar && req.Service != "" {
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "$service", req.Service)
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "'$service'", "'"+req.Service+"'")
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "\"$service\"", "\""+req.Service+"\"")
			}

			// 如果查询中没有花括号，则根据需要添加标签
			if !strings.Contains(queryWithLabels, "{") {
				// 没有花括号，需要添加标签
				var labels []string

				// project 是强制的：
				// - 如果查询包含 $project 变量，上面已经替换了，不需要添加标签
				// - 如果查询不包含 $project 变量，则强制添加 project 标签
				if !needProjectVar && projectName != "" {
					labels = append(labels, `project="`+projectName+`"`)
				}

				// namespace 是可选的：
				// - 如果查询包含 $namespace 变量，上面已经替换了，不需要添加标签
				// - 如果查询不包含 $namespace 变量，不添加 namespace 标签

				// 添加标签到查询语句
				if len(labels) > 0 {
					queryWithLabels = queryWithLabels + `{` + strings.Join(labels, ",") + `}`
				}
			}

			// 记录最终查询语句
			//middleware.Logger.Infof("查询语句: %s", queryWithLabels)

			// 根据图表类型选择不同的查询方式
			var queryResult interface{}
			var err error

			if v.ChartType == "line" {
				// 折线图使用范围查询
				//middleware.Logger.Infof("执行范围查询，时间范围: %d-%d, 步长: %d", startMillis, endMillis, stepMillis)
				queryResult, err = database.QueryRangeWithMilliseconds(r.Context(), queryWithLabels, startMillis, endMillis, stepMillis)
			} else {
				// 其他图表类型（如饼图、柱状图等）使用瞬时查询
				//middleware.Logger.Infof("执行瞬时查询，时间: %d", endMillis)
				queryResult, err = database.QueryInstantWithMilliseconds(r.Context(), queryWithLabels, endMillis)
			}

			// 构造响应
			response := MetricsListResponse{
				ViewID:      v.ID,
				ViewName:    v.ViewName,
				Sort:        v.Sort,
				Standard:    v.Standard,
				Description: v.Description,
				Query:       v.Query,
				ChartType:   v.ChartType,
			}

			if err != nil {
				middleware.Logger.Errorf("查询视图[%s]失败: %v", v.ViewName, err)
				response.Error = err.Error()
			} else {
				response.Data = queryResult
			}

			// 将结果添加到结果集
			mu.Lock()
			results = append(results, response)
			mu.Unlock()
		}(view)
	}

	// 等待所有查询完成
	wg.Wait()

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", results)
}
