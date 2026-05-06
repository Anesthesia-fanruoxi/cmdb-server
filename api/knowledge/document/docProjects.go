package document

import (
	"cmdb/utils"
	"net/http"
	"net/url"
	"strings"
)

// DocProjects 获取知识库文档页面可访问的项目列表
// GET /api/knowledge/doc/projects
// 从Referer header获取当前页面路径，自动匹配菜单权限
func DocProjects(w http.ResponseWriter, r *http.Request) {
	// 从Referer获取当前页面路径
	menuPath := "/knowledge/document" // 默认值
	referer := r.Header.Get("Referer")
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			path := strings.TrimSuffix(u.Path, "/")
			if strings.HasPrefix(path, "/knowledge/") {
				menuPath = path
			}
		}
	}
	// 文档模块需要包含ops项目
	utils.GetProjectListByMenuPath(w, r, menuPath, true)
}
