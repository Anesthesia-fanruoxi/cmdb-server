package domain

import (
	"cmdb/config"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

// CreateClient 创建阿里云DNS客户端
// 从配置中读取凭据并初始化
func CreateClient() (*alidns20150109.Client, error) {
	// 从配置中读取阿里云访问凭证
	cfg := config.GetConfig()
	accessKeyId := cfg.Domain.AliyunAccessKeyID
	accessKeySecret := cfg.Domain.AliyunAccessKeySecret

	// 配置认证信息
	authConfig := &openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
	}

	// Endpoint 设置
	authConfig.Endpoint = tea.String("alidns.cn-beijing.aliyuncs.com")
	client, err := alidns20150109.NewClient(authConfig)

	return client, err
}

// GetDomainName 获取配置中的域名
func GetDomainName() string {
	cfg := config.GetConfig()
	return cfg.Domain.AliyunDomain
}
