package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

// Config 配置结构体
type Config struct {
	Server     ServerConfig     `yaml:"server"`     // 服务器配置
	MySQL      MySQLConfig      `yaml:"mysql"`      // MySQL配置
	Redis      RedisConfig      `yaml:"redis"`      // Redis配置
	Prometheus PrometheusConfig `yaml:"prometheus"` // Prometheus配置
	Kubernetes KubernetesConfig `yaml:"kubernetes"` // Kubernetes配置
	Security   SecurityConfig   `yaml:"security"`   // 安全配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host   string `yaml:"host"` // 监听地址
	Port   int    `yaml:"port"` // 监听端口
	Domain string `yaml:"domain"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `yaml:"host"`     // 数据库主机地址
	Port     int    `yaml:"port"`     // 数据库端口
	Username string `yaml:"username"` // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	Database string `yaml:"database"` // 数据库名称
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`     // Redis主机地址
	Port     int    `yaml:"port"`     // Redis端口
	Password string `yaml:"password"` // Redis密码
	DB       int    `yaml:"db"`       // Redis数据库索引
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Host string `yaml:"host"` // Prometheus主机地址
	Port int    `yaml:"port"` // Prometheus端口
}

// KubernetesConfig Kubernetes配置
type KubernetesConfig struct {
	ConfigPath string `yaml:"config_path"` // K8s集群配置文件路径
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret    string `yaml:"jwt_secret"`    // JWT签名密钥
	PasswordSalt string `yaml:"password_salt"` // 密码加密盐值
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetDSN 返回MySQL连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.Username,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.Database,
	)
}
