package initData

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient *kubernetes.Clientset

// InitK8s 初始化K8s客户端
func InitK8s(kubeconfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("构建K8s配置失败: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建K8s客户端失败: %v", err)
	}

	k8sClient = clientset
	return nil
}

// GetK8s 获取K8s客户端
func GetK8s() *kubernetes.Clientset {
	return k8sClient
}

// GetK8sOrNil 获取K8s客户端，如果未初始化则返回nil
func GetK8sOrNil() *kubernetes.Clientset {
	return k8sClient
}
