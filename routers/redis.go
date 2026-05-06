package routers

import (
	"cmdb/api/redis/search"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitRedisRoutes 初始化Redis相关路由
func InitRedisRoutes(mux *http.ServeMux) {
	// Redis管理
	mux.HandleFunc(path.RedisProjectsPath, utils.JWTAuth(search.RedisProjects, "redis:search:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.RedisInfoPath, utils.JWTAuth(search.RedisInfo, "redis:search:r", "", "获取Redis信息"))
	mux.HandleFunc(path.RedisTreePath, utils.JWTAuth(search.RedisTree, "redis:search:r", "", "获取Redis keys树形列表"))
	mux.HandleFunc(path.RedisKeyPath, utils.JWTAuth(search.RedisKey, "redis:search:r", "", "获取Redis key值"))
	mux.HandleFunc(path.RedisDeletePath, utils.JWTAuth(search.RedisDelete, "redis:search:w", "", "删除Redis key"))
}
