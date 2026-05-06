package kubeService

import (
	"cmdb/utils"
	"net/http"
)

// KubeServiceProjects 获取K8s服务页面可访问的项目列表
// GET /api/assets/kubeService/projects
func KubeServiceProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/kubeService")
}
