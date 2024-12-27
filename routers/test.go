package routers

import (
	"cmdb/api/asset/test"
	"cmdb/utils"
	"net/http"
)

// InitTestAssetRoutes 初始化测试环境资产管理相关路由
func InitTestAssetRoutes(mux *http.ServeMux) {

	// 集群相关路由
	mux.HandleFunc(ClusterTestStatusPath, utils.JWTAuth(test.GetClusterStatus, "assets:test:overview"))        //环境预览接口
	mux.HandleFunc(ClusterTestScalePath, utils.JWTAuth(test.ScalePod, "assets:test"))                          //扩缩容接口
	mux.HandleFunc(ClusterTestServicePath, utils.JWTAuth(test.GetClusterServices, "assets:test:port-mapping")) //端口映射
	mux.HandleFunc(ClusterTestBatchScalePath, utils.JWTAuth(test.BatchScalePods, "assets:test"))               //批量扩缩容

	// 迭代相关路由
	mux.HandleFunc(IterationStartPath, utils.JWTAuth(test.StartIteration, "assets:test"))
}
