package routers

import (
	"cmdb/api/asset/test"
	"cmdb/utils"
	"net/http"
)

// InitTestAssetRoutes 初始化测试环境资产管理相关路由
func InitTestAssetRoutes(mux *http.ServeMux) {

	// 集群相关路由
	mux.HandleFunc(ClusterTestStatusPath, utils.JWTAuth(test.GetClusterStatus))
	mux.HandleFunc(ClusterTestScalePath, utils.JWTAuth(test.ScalePod))
	mux.HandleFunc(ClusterTestServicePath, utils.JWTAuth(test.GetClusterServices))
	//mux.HandleFunc(ClusterTestNamespacePath, utils.JWTAuth(test.GetNamespacesByProject))
	mux.HandleFunc(ClusterTestBatchScalePath, utils.JWTAuth(test.BatchScalePods))

	// 迭代相关路由
	mux.HandleFunc(IterationStartPath, utils.JWTAuth(test.StartIteration))
}
