package metrics

import (
	"cmdb/config"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MonitorMetricsListSSE SSE批量查询监控指标（每5秒推送一次）
// 参数说明：
// - token: JWT令牌（必选）
// - project: 项目标识（必选）- 用于传递项目信息
// - category: 视图分类（必选）- 用于查询对应的监控视图
// - namespace: 命名空间（可选）- 主要用于容器监控，未传递时自动推导为 {project}-service
// - service: 服务名称（可选）- 主要用于流量监控
func MonitorMetricsListSSE(w http.ResponseWriter, r *http.Request) {
	// 从URL参数获取token
	token := r.URL.Query().Get("token")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	// 验证token
	jwtConfig := config.GetConfig().JWT
	_, err := utils.ParseToken(token, jwtConfig.Secret)
	if err != nil {
		middleware.ResponseError(w, http.StatusUnauthorized, "token无效: "+err.Error())
		return
	}

	// 从URL参数获取查询参数
	project := r.URL.Query().Get("project")
	category := r.URL.Query().Get("category")
	namespace := r.URL.Query().Get("namespace")
	service := r.URL.Query().Get("service")

	// 参数验证
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	if category == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "category参数不能为空")
		return
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// 发送连接成功消息
	sendSSEMetricsEvent(w, flusher, "connected", map[string]string{"message": "SSE连接成功"})

	ctx := r.Context()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 立即执行一次查询
	queryMetricsAndSend(w, flusher, r, project, category, namespace, service)

	// 每5秒执行一次查询
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queryMetricsAndSend(w, flusher, r, project, category, namespace, service)
		}
	}
}

// queryMetricsAndSend 执行批量查询并发送结果
func queryMetricsAndSend(w http.ResponseWriter, flusher http.Flusher, r *http.Request, project, category, namespace, service string) {
	// 查询项目对应的名称
	var projectName string
	if err := database.DB.Table(middleware.TableProjectDict).Where("project = ?", project).Pluck("project_name", &projectName).Error; err != nil {
		middleware.Logger.Errorf("查询项目名称失败: %v", err)
		sendSSEMetricsEvent(w, flusher, "error", map[string]string{"message": "查询项目名称失败"})
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
		Where("category = ? ", category).
		Find(&views).Error; err != nil {
		middleware.Logger.Errorf("查询监控视图失败: %v", err)
		sendSSEMetricsEvent(w, flusher, "error", map[string]string{"message": "查询视图失败"})
		return
	}

	if len(views) == 0 {
		sendSSEMetricsEvent(w, flusher, "data", []interface{}{})
		return
	}

	// 使用当前时间作为查询时间（瞬时查询）
	endTime := time.Now()
	endMillis := middleware.TimeToMilliseconds(endTime)

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

			// 检查查询中是否包含变量
			queryLower := strings.ToLower(queryWithLabels)
			needProjectVar := strings.Contains(queryLower, "$project")
			needNamespaceVar := strings.Contains(queryLower, "$namespace")
			needServiceVar := strings.Contains(queryLower, "$service")

			// namespace处理
			var namespaceValue string
			if needNamespaceVar {
				namespaceValue = namespace
				if namespaceValue == "" {
					namespaceValue = project + "-service"
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

			if needServiceVar && service != "" {
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "$service", service)
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "'$service'", "'"+service+"'")
				queryWithLabels = strings.ReplaceAll(queryWithLabels, "\"$service\"", "\""+service+"\"")
			}

			// 如果查询中没有花括号，则根据需要添加标签
			if !strings.Contains(queryWithLabels, "{") {
				var labels []string
				if !needProjectVar && projectName != "" {
					labels = append(labels, `project="`+projectName+`"`)
				}
				if len(labels) > 0 {
					queryWithLabels = queryWithLabels + `{` + strings.Join(labels, ",") + `}`
				}
			}

			// SSE只做瞬时查询（不支持折线图的范围查询）
			queryResult, err := database.QueryInstantWithMilliseconds(r.Context(), queryWithLabels, endMillis)

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

			mu.Lock()
			results = append(results, response)
			mu.Unlock()
		}(view)
	}

	// 等待所有查询完成
	wg.Wait()

	// 发送查询结果
	sendSSEMetricsEvent(w, flusher, "data", results)
}

// sendSSEMetricsEvent 发送SSE事件
func sendSSEMetricsEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
