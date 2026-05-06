package search

import (
	"cmdb/config"
	"cmdb/pkg/middleware"
	"fmt"
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// SLSClientManager SLS客户端管理器
type SLSClientManager struct {
	defaultAccount string
	clients        map[string]*sls20201230.Client // 账号名到客户端的映射
	mu             sync.RWMutex
}

// GlobalSLSManager 全局SLS客户端管理器
var GlobalSLSManager *SLSClientManager

// InitSLSManager 初始化SLS客户端管理器
func InitSLSManager(cfg *config.SLSConfig) error {
	middleware.Logger.Infof("[SLS] 初始化SLS客户端管理器, 默认账号: %s", cfg.Current)

	manager := &SLSClientManager{
		defaultAccount: cfg.Current,
		clients:        make(map[string]*sls20201230.Client),
	}

	// 为每个配置创建客户端
	for _, profile := range cfg.Profiles {
		// 获取endpoint
		endpoint, ok := cfg.Endpoints[profile.Name]
		if !ok {
			// 如果没有指定endpoint，使用默认值
			endpoint = "cn-hangzhou.log.aliyuncs.com"
		}

		// 创建客户端
		client, err := createSLSClient(profile, endpoint)
		if err != nil {
			middleware.Logger.Errorf("[SLS] 创建账号 %s 的SLS客户端失败: %v", profile.Name, err)
			return fmt.Errorf("创建SLS客户端失败(账号:%s): %v", profile.Name, err)
		}

		// 存储客户端
		manager.clients[profile.Name] = client
	}

	GlobalSLSManager = manager
	middleware.Logger.Infof("[SLS] SLS客户端管理器初始化完成, 默认账号: %s, 客户端数量: %d", manager.defaultAccount, len(manager.clients))

	return nil
}

// GetClient 获取SLS客户端
func (m *SLSClientManager) GetClient(project string) (*sls20201230.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 尝试直接使用project作为账号名
	client, ok := m.clients[project]
	if ok {
		return client, nil
	}

	// 如果找不到对应的账号，使用默认账号
	defaultClient, ok := m.clients[m.defaultAccount]
	if ok {
		return defaultClient, nil
	}

	return nil, fmt.Errorf("找不到账号 [%s] 对应的SLS客户端，且默认账号 [%s] 也不可用",
		project, m.defaultAccount)
}

// GetDefaultAccount 获取当前默认SLS账号
func (m *SLSClientManager) GetDefaultAccount() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultAccount
}

// createSLSClient 创建SLS客户端
func createSLSClient(profile config.SLSProfile, endpoint string) (*sls20201230.Client, error) {
	switch profile.Mode {
	case "AK":
		return createAKClient(profile.AccessKeyID, profile.AccessKeySecret, endpoint)
	default:
		return nil, fmt.Errorf("不支持的认证模式: %s", profile.Mode)
	}
}

// createAKClient 创建AK模式的SLS客户端
func createAKClient(accessKeyID, accessKeySecret, endpoint string) (*sls20201230.Client, error) {
	// 创建凭证配置
	credentialConfig := &credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	// 创建凭证实例
	cred, err := credential.NewCredential(credentialConfig)
	if err != nil {
		return nil, fmt.Errorf("创建凭证失败: %v", err)
	}

	// 创建OpenAPI配置
	slsConfig := &openapi.Config{
		Credential: cred,
		Endpoint:   tea.String(endpoint),
	}

	// 创建客户端
	return sls20201230.NewClient(slsConfig)
}
