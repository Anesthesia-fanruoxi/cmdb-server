package metrics

import (
	"cmdb/utils"
	"net/http"
	"net/url"
	"strings"
)

// MetricsProjects 获取监控指标页面可访问的项目列表
// GET /api/monitor/metrics/projects
// 从Referer header获取当前页面路径，自动匹配菜单权限
func MetricsProjects(w http.ResponseWriter, r *http.Request) {
	// 从Referer获取当前页面路径
	menuPath := "/monitor/metrics" // 默认值
	referer := r.Header.Get("Referer")
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			// 提取路径，如 /monitor/trafficswitching
			path := strings.TrimSuffix(u.Path, "/")
			if strings.HasPrefix(path, "/monitor/") {
				menuPath = path
			}
		}
	}
	utils.GetProjectListByMenuPath(w, r, menuPath)
}
