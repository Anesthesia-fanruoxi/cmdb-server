package path

// 审计管理路径
const (
	// K8s Pod管理
	KubePodProjectsPath  = "/api/assets/kubePod/projects" // 获取可访问的项目列表
	KubeGetPodStatusPath = "/api/assets/kubePod/status"   // 获取namespace状态

	// K8s服务管理
	KubeServiceProjectsPath = "/api/assets/kubeService/projects"  // 获取可访问的项目列表
	KubeGetNSPath           = "/api/assets/kubeService/namespace" // 获取namespace状态
	KubeGetServicePath      = "/api/assets/kubeService/list"      // 获取Pod服务列表

	// 项目发版管理
	ProUpdateProjectsPath = "/api/assets/proUpdate/projects" // 获取可访问的项目列表
	ProUpdateCreatePath   = "/api/assets/proUpdate/create"   // 创建项目发版任务
	ProUpdateListPath     = "/api/assets/proUpdate/list"     // 获取项目发版列表
	ProUpdateRecordsPath  = "/api/assets/proUpdate/records"  // 接收项目发版记录
	ProUpdateLogsWSPath   = "/ws/assets/proUpdate/logs"      // WebSocket日志代理
	ProUpdateCancelPath   = "/api/assets/proUpdate/cancel"   // 取消项目发版任务

	// sse接口
	ProUpdateListDetailPath    = "/sse/assets/proUpdate/list-detail"    // 获取项目发版列表详情
	ProUpdateRecordsDetailPath = "/sse/assets/proUpdate/records-detail" // 查看项目发版记录详情
	ProUpdateMobileSSEPath     = "/sse/assets/proUpdate/mobile"         // 手机端SSE接口

	// 域名管理
	DomainListPath         = "/api/assets/domain/list"   // 获取域名解析列表
	DomainAddPath          = "/api/assets/domain/add"    // 添加域名解析记录
	DomainUpdatePath       = "/api/assets/domain/update" // 更新域名解析记录
	DomainUpdateStatusPath = "/api/assets/domain/status" // 更新域名解析记录状态
	DomainDeletePath       = "/api/assets/domain/delete" // 删除域名解析记录

	// 密钥管理
	KeyProjectsPath     = "/api/assets/key/projects"     // 获取可访问的项目列表
	KeyEncryptPath      = "/api/assets/key/encrypt"      // 加密接口
	KeyDecryptPath      = "/api/assets/key/decrypt"      // 解密接口
	KeyBatchDecryptPath = "/api/assets/key/batchDecrypt" // 批量解密接口

	// 文件管理
	FileProjectsPath = "/api/assets/file/projects" // 获取可访问的项目列表
	FileListPath     = "/api/assets/file/list"     // 获取目录文件列表
	FileUploadPath   = "/api/assets/file/upload"   // 上传文件

	// 阿里云余额
	AlBalanceProjectsPath   = "/api/assets/alBalance/projects"      // 获取可访问项目列表
	AlBalancePath           = "/api/assets/alBalance/balance"       // 查询阿里云账户余额
	AlBalanceConfigListPath = "/api/assets/alBalance/config/list"   // 获取监控项目配置
	AlBalanceConfigPath     = "/api/assets/alBalance/config/update" // 更新监控项目配置
	AlEcsListPath           = "/api/assets/alBalance/ecs/list"      // 查询ECS到期信息
)
