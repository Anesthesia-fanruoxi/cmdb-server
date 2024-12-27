package main

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/routers"
	"cmdb/utils"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 加载配置
	if err := initData.InitConfig("config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := initData.GetConfig()

	// 设置模型配置
	models.SetConfig(cfg)

	// 初始化数据库
	if err := initData.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化Redis连接（包含角色权限初始化）
	if err := initData.InitRedis(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	// 初始化K8s客户端
	if err := initData.InitK8s(cfg.Kubernetes.ConfigPath); err != nil {
		log.Printf("初始化K8s客户端失败: %v", err)
	}

	// 设置JWT配置
	utils.SetConfig(cfg)

	// 设置路由
	mux := http.NewServeMux()
	routers.InitRoutes(mux)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务器启动在 %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
