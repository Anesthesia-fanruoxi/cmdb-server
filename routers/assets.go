package routers

import (
	"cmdb/api/assets/alBalance"
	"cmdb/api/assets/domain"
	"cmdb/api/assets/file"
	"cmdb/api/assets/key"
	"cmdb/api/assets/kubePod"
	"cmdb/api/assets/kubeService"
	"cmdb/api/assets/proUpdate"
	"cmdb/routers/path"
	"cmdb/utils"
	"net/http"
)

// InitAssetsRoutes 初始化审计相关路由
func InitAssetsRoutes(mux *http.ServeMux) {
	// K8s Pod管理
	mux.HandleFunc(path.KubePodProjectsPath, utils.JWTAuth(kubePod.KubePodProjects, "assets:kubePod:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.KubeGetPodStatusPath, utils.JWTAuth(kubePod.K8sGetStatus, "assets:kubePod:r", "", "获取pod状态"))

	// K8s服务管理
	mux.HandleFunc(path.KubeServiceProjectsPath, utils.JWTAuth(kubeService.GetProjectList, "assets:kubeService:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.KubeGetServicePath, utils.JWTAuth(kubeService.GetService, "assets:kubeService:r", "", ""))
	mux.HandleFunc(path.KubeGetNSPath, utils.JWTAuth(kubeService.GetNS, "assets:kubeService:r", "", ""))

	// 项目发版管理
	mux.HandleFunc(path.ProUpdateProjectsPath, utils.JWTAuth(proUpdate.ProUpdateProjects, "assets:proUpdate:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.ProUpdateCreatePath, utils.JWTAuth(proUpdate.ProUpdateCreate, "assets:proUpdate:w", "o", "创建项目发版任务"))
	mux.HandleFunc(path.ProUpdateListPath, utils.JWTAuth(proUpdate.ProUpdateList, "assets:proUpdate:r", "", "获取项目发版列表"))
	mux.HandleFunc(path.ProUpdateListDetailPath, proUpdate.ProUpdateListDetail)      // 获取项目发版详情
	mux.HandleFunc(path.ProUpdateRecordsDetailPath, proUpdate.ProUpdateRecordDetail) // 获取项目发版详情
	mux.HandleFunc(path.ProUpdateRecordsPath, proUpdate.ProUpdateRecords)            // 不需要JWT认证，允许外部系统调用
	mux.HandleFunc(path.ProUpdateLogsWSPath, proUpdate.ProUpdateLogsWS)              // WebSocket不能使用JWT认证
	mux.HandleFunc(path.ProUpdateCancelPath, utils.JWTAuth(proUpdate.ProUpdateCancel, "assets:proUpdate:w", "o", "取消项目发版任务"))
	mux.HandleFunc(path.ProUpdateMobileSSEPath, proUpdate.ProUpdateMobileSSE) // 手机端SSE接口

	// 域名管理
	mux.HandleFunc(path.DomainListPath, utils.JWTAuth(domain.DomainList, "assets:domain:r", "", "获取域名解析列表"))
	mux.HandleFunc(path.DomainAddPath, utils.JWTAuth(domain.DomainAdd, "assets:domain:w", "o", "添加域名解析记录"))
	mux.HandleFunc(path.DomainUpdatePath, utils.JWTAuth(domain.DomainUpdate, "assets:domain:w", "o", "更新域名解析记录"))
	mux.HandleFunc(path.DomainUpdateStatusPath, utils.JWTAuth(domain.DomainUpdateStatus, "assets:domain:w", "o", "更新域名解析记录状态"))
	mux.HandleFunc(path.DomainDeletePath, utils.JWTAuth(domain.DomainDelete, "assets:domain:w", "o", "删除域名解析记录"))

	// 密钥管理
	mux.HandleFunc(path.KeyProjectsPath, utils.JWTAuth(key.KeyProjects, "assets:key:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.KeyEncryptPath, utils.JWTAuth(key.KeyEncrypt, "assets:key:r", "", "数据加密"))
	mux.HandleFunc(path.KeyDecryptPath, utils.JWTAuth(key.KeyDecrypt, "assets:key:r", "", "数据解密"))
	mux.HandleFunc(path.KeyBatchDecryptPath, utils.JWTAuth(key.KeyBatchDecrypt, "assets:key:r", "", "批量数据解密"))

	// 文件管理
	mux.HandleFunc(path.FileProjectsPath, utils.JWTAuth(file.FileProjects, "assets:file:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.FileListPath, utils.JWTAuth(file.FileList, "assets:file:r", "", "获取目录文件列表"))
	mux.HandleFunc(path.FileUploadPath, utils.JWTAuth(file.FileUpload, "assets:file:w", "", "上传文件"))

	// 阿里云余额
	mux.HandleFunc(path.AlBalanceProjectsPath, utils.JWTAuth(alBalance.AlBalanceProjects, "assets:alBalance:r", "", "获取可访问项目列表"))
	mux.HandleFunc(path.AlBalancePath, utils.JWTAuth(alBalance.GetBalance, "assets:alBalance:r", "", "查询阿里云账户余额"))
	mux.HandleFunc(path.AlBalanceConfigListPath, utils.JWTAuth(alBalance.GetConfig, "assets:alBalance:r", "", "获取监控项目配置"))
	mux.HandleFunc(path.AlBalanceConfigPath, utils.JWTAuth(alBalance.UpdateConfig, "assets:alBalance:w", "o", "更新监控项目配置"))
	mux.HandleFunc(path.AlEcsListPath, utils.JWTAuth(alBalance.GetEcsList, "assets:alBalance:r", "", "查询ECS到期信息"))
}
