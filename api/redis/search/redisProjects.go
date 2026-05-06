package search

import (
	"cmdb/utils"
	"net/http"
)

// RedisProjects 获取Redis管理页面可访问的项目列表
// GET /api/assets/redis/projects
func RedisProjects(w http.ResponseWriter, r *http.Request) {
	utils.GetProjectListByMenuPath(w, r, "/assets/redis")
}
