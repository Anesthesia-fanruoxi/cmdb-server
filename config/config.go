package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 用于加载配置文件
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	K8s        K8sConfig        `mapstructure:"k8s"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Job        JobConfig        `mapstructure:"job"`
	Logs       LogsConfig       `mapstructure:"logs"`       // 日志配置
	Prometheus PrometheusConfig `mapstructure:"prometheus"` // Prometheus配置
	SLS        SLSConfig        `mapstructure:"sls"`
	Email      EmailConfig      `mapstructure:"email"`  // 邮件配置
	Domain     DomainConfig     `mapstructure:"domain"` // 域名配置
	FeiShu     FeiShuConfig     `mapstructure:"feishu"` // 飞书配置
	CICD       CICDConfig       `mapstructure:"cicd"`   // CICD配置
	Agent      AgentConfig      `mapstructure:"agent"`  // Agent插件配置
	Phone      PhoneConfig      `mapstructure:"phone"`  // 手机号查询接口配置
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	UserName        string `mapstructure:"user_name"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 连接最大生命周期(秒)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"` // JWT密钥
	ExpireHours int    `mapstructure:"expire"` // 过期时间(小时)
}
type K8sConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Secret   string `mapstructure:"secret"`
}

type AuthConfig struct {
	PasswordSalt string `mapstructure:"password_salt"` // 密码加密盐
	BcryptCost   int    `mapstructure:"bcrypt_cost"`   // bcrypt加密成本
}

type JobConfig struct {
	// TaskSalt 已移至 AgentConfig.Secret
}

// LogsConfig 日志配置
type LogsConfig struct {
	Enable bool   `mapstructure:"enable"` // 是否启用日志
	Level  string `mapstructure:"level"`  // 日志级别: info, error
}

type PrometheusConfig struct {
	BaseURL   string `mapstructure:"base_url"`   // 基础URL (用于指标查询，可以是VictoriaMetrics)
	AlertsURL string `mapstructure:"alerts_url"` // 告警URL (用于告警查询，必须是Prometheus)
	Timeout   int    `mapstructure:"timeout"`    // 超时时间(秒)
}

// SLSConfig SLS配置
type SLSConfig struct {
	Current   string            `mapstructure:"current"`   // 当前默认账号
	Profiles  []SLSProfile      `mapstructure:"profiles"`  // 账号配置列表
	Endpoints map[string]string `mapstructure:"endpoints"` // 账号名称到endpoint的映射
}

// SLSProfile SLS账号配置
type SLSProfile struct {
	Name            string `mapstructure:"name"`              // 账号名称
	Mode            string `mapstructure:"mode"`              // 认证模式: AK, RamRoleArn等
	AccessKeyID     string `mapstructure:"access_key_id"`     // AccessKey ID
	AccessKeySecret string `mapstructure:"access_key_secret"` // AccessKey Secret
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost string `mapstructure:"smtp_host"` // SMTP服务器地址
	SMTPPort int    `mapstructure:"smtp_port"` // SMTP服务器端口
	SMTPUser string `mapstructure:"smtp_user"` // SMTP用户名
	SMTPPass string `mapstructure:"smtp_pass"` // SMTP密码
	From     string `mapstructure:"from"`      // 发件人邮箱
}

// DomainConfig 域名配置
type DomainConfig struct {
	AliyunAccessKeyID     string `mapstructure:"aliyun_access_key_id"`     // 阿里云AccessKeyID
	AliyunAccessKeySecret string `mapstructure:"aliyun_access_key_secret"` // 阿里云AccessKeySecret
	AliyunDomain          string `mapstructure:"aliyun_domain"`            // 阿里云域名
}

// FeiShuConfig 飞书配置
type FeiShuConfig struct {
	AppID             string `mapstructure:"app_id"`             // 飞书应用ID
	AppSecret         string `mapstructure:"app_secret"`         // 飞书应用密钥
	BaseURL           string `mapstructure:"base_url"`           // 飞书API基础地址
	VerificationToken string `mapstructure:"verification_token"` // 飞书事件验证Token
	EncryptKey        string `mapstructure:"encrypt_key"`        // 飞书事件加密Key
}

// CICDConfig CICD配置
type CICDConfig struct {
	URL    string `mapstructure:"url"`    // CICD服务地址
	Secret string `mapstructure:"secret"` // CICD服务密钥
}

// AgentConfig Agent插件配置
type AgentConfig struct {
	ImageUrl    string `mapstructure:"image_url"`    // 容器镜像仓库地址（如：hub.hzbxhd.com/test）
	DownloadUrl string `mapstructure:"download_url"` // 二进制文件下载地址（如：https://cdn.example.com/plugins）
	Secret      string `mapstructure:"secret"`       // Agent插件加密密钥（用于任务调度、IP项目关联等）
}

// PhoneConfig 手机号查询接口配置
type PhoneConfig struct {
	Secret string `mapstructure:"secret"` // 手机号查询接口加密密钥
}

var globalConfig Config

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("mysql.max_idle_conns", 10)
	viper.SetDefault("mysql.max_open_conns", 100)
	viper.SetDefault("mysql.conn_max_lifetime", 3600)
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", 24)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("auth.bcrypt_cost", 10)
	viper.SetDefault("job.task_salt", "")
	viper.SetDefault("logs.enable", true)                            // 设置日志默认为启用
	viper.SetDefault("logs.level", "info")                           // 设置日志默认级别为info
	viper.SetDefault("prometheus.base_url", "http://localhost:9090") // 指标查询默认地址
	viper.SetDefault("prometheus.alerts_url", "")                    // 告警查询地址(空则使用base_url)
	viper.SetDefault("prometheus.timeout", 30)                       // 默认超时时间30秒
	viper.SetDefault("sls.current", "bjjf")                          // 设置SLS默认账号
	viper.SetDefault("email.smtp_host", "smtp.example.com")          // 默认SMTP服务器地址
	viper.SetDefault("email.smtp_port", 465)                         // 默认SMTP服务器端口
	viper.SetDefault("email.smtp_user", "")                          // 默认SMTP用户名
	viper.SetDefault("email.smtp_pass", "")                          // 默认SMTP密码
	viper.SetDefault("email.from", "noreply@example.com")            // 默认发件人邮箱
	viper.SetDefault("domain.aliyun_access_key_id", "")              // 默认阿里云AccessKeyID
	viper.SetDefault("domain.aliyun_access_key_secret", "")          // 默认阿里云AccessKeySecret
	viper.SetDefault("domain.aliyun_domain", "")                     // 默认阿里云域名
	viper.SetDefault("feishu.app_id", "")                            // 默认飞书应用ID
	viper.SetDefault("feishu.app_secret", "")                        // 默认飞书应用密钥
	viper.SetDefault("feishu.base_url", "https://open.feishu.cn")    // 默认飞书API地址
	viper.SetDefault("cicd.url", "http://192.168.6.2:8083")          // 默认CICD代理服务地址
	viper.SetDefault("cicd.secret", "")                              // 默认CICD代理服务密钥
	viper.SetDefault("phone.secret", "")                             // 默认手机号查询接口加密密钥

	if err := viper.ReadInConfig(); err != nil {
		// 如果读取配置文件失败，使用默认配置
		globalConfig = Config{
			Server: ServerConfig{
				Port: 8080,
			},
			MySQL: MySQLConfig{
				Host:            "localhost",
				Port:            3306,
				UserName:        "root",
				Password:        "your_password",
				Database:        "cmdb",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 3600,
			},
			Redis: RedisConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			JWT: JWTConfig{
				Secret:      "your-secret-key",
				ExpireHours: 24,
			},
			Auth: AuthConfig{
				PasswordSalt: "default-salt",
				BcryptCost:   10,
			},
			Job: JobConfig{},
			Logs: LogsConfig{
				Enable: true, // 默认启用日志
				Level:  "info",
			},
			Prometheus: PrometheusConfig{
				BaseURL:   "http://localhost:9090",
				AlertsURL: "",
				Timeout:   30,
			},
			SLS: SLSConfig{},
			Email: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 465,
				SMTPUser: "",
				SMTPPass: "",
				From:     "noreply@example.com",
			},
			Domain: DomainConfig{
				AliyunAccessKeyID:     "",
				AliyunAccessKeySecret: "",
				AliyunDomain:          "",
			},
			FeiShu: FeiShuConfig{
				AppID:     "",
				AppSecret: "",
				BaseURL:   "https://open.feishu.cn",
			},
			CICD: CICDConfig{
				URL:    "http://192.168.6.2:8083",
				Secret: "",
			},
		}
		return nil
	}

	if err := viper.Unmarshal(&globalConfig); err != nil {
		return err
	}

	// 手动设置SLS配置
	if viper.IsSet("sls") {
		// 读取当前默认账号
		globalConfig.SLS.Current = viper.GetString("sls.current")

		// 读取配置文件的profiles部分
		if viper.IsSet("sls.profiles") {
			var profiles []SLSProfile

			// 遍历profiles数组
			profilesArray := viper.Get("sls.profiles")
			if profilesSlice, ok := profilesArray.([]interface{}); ok {
				for _, p := range profilesSlice {
					if profileMap, ok := p.(map[string]interface{}); ok {
						profile := SLSProfile{
							Name:            fmt.Sprintf("%v", profileMap["name"]),
							Mode:            fmt.Sprintf("%v", profileMap["mode"]),
							AccessKeyID:     fmt.Sprintf("%v", profileMap["access_key_id"]),
							AccessKeySecret: fmt.Sprintf("%v", profileMap["access_key_secret"]),
						}
						profiles = append(profiles, profile)
					}
				}
			}

			globalConfig.SLS.Profiles = profiles
		}

		// 读取endpoints部分
		if viper.IsSet("sls.endpoints") {
			endpoints := make(map[string]string)
			endpointsMap := viper.Get("sls.endpoints")
			if endpointsData, ok := endpointsMap.(map[string]interface{}); ok {
				for k, v := range endpointsData {
					endpoints[k] = fmt.Sprintf("%v", v)
				}
			}
			globalConfig.SLS.Endpoints = endpoints
		}
	}

	return nil
}

func GetConfig() *Config {
	return &globalConfig
}

// GetAgentConfig 获取Agent插件配置
func GetAgentConfig() AgentConfig {
	return globalConfig.Agent
}
