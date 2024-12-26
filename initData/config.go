package initData

import "cmdb/config"

var cfg *config.Config

// InitConfig 初始化配置
func InitConfig(configFile string) error {
	var err error
	cfg, err = config.LoadConfig(configFile)
	if err != nil {
		return err
	}
	return nil
}

// GetConfig 获取配置
func GetConfig() *config.Config {
	return cfg
}
