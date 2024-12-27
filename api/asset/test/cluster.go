package test

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterServicesRequest 获取集群服务请求结构
type ClusterServicesRequest struct {
	Projects []string `json:"projects"` // 项目列表
}

// GetClusterServices 获取服务端口信息
func GetClusterServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 从查询参数获取项目列表
	projectsParam := r.URL.Query().Get("projects")
	var projects []string
	if projectsParam != "" {
		projects = strings.Split(projectsParam, ",")
	}

	// 获取k8s客户端
	k8sClient := initData.GetK8s()
	if k8sClient == nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "K8s客户端未初始化")
		return
	}

	// 获取配置
	cfg := initData.GetConfig()
	if cfg == nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取配置失败")
		return
	}

	// 获取所有Service（跨namespace）
	services, err := k8sClient.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取Service列表失败")
		return
	}

	// 获取项目字典
	db := initData.GetDB()
	var projectDicts []struct {
		Project     string `gorm:"column:project"`
		ProjectName string `gorm:"column:project_name"`
	}
	if err := db.Table("project_dict").Select("project, project_name").Find(&projectDicts).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取项目字典失败")
		return
	}

	// 构建项目字典映射
	projectDict := make(map[string]string)
	for _, dict := range projectDicts {
		projectDict[dict.Project] = dict.ProjectName
	}

	// 如果指定了项目列表且列表为空，直接返回空结果
	if projects != nil && len(projects) == 0 {
		utils.Success(w, []models.ServicePortResponse{})
		return
	}

	// 构建项目过滤集合
	projectFilter := make(map[string]bool)
	if len(projects) > 0 {
		for _, p := range projects {
			projectFilter[p] = true
		}
	}

	// 构建返回数据
	var result []models.ServicePortResponse
	for _, svc := range services.Items {
		// 跳过包含public的namespace
		if strings.Contains(svc.Namespace, "public") {
			continue
		}

		// 检查namespace是否包含service或middleware
		nsLower := strings.ToLower(svc.Namespace)
		if !strings.Contains(nsLower, "service") && !strings.Contains(nsLower, "middleware") {
			continue
		}

		// 获取项目代码（namespace的第一个部分）
		parts := strings.Split(svc.Namespace, "-")
		if len(parts) < 2 {
			continue
		}
		projectKey := parts[0]

		// 如果指定了项目列表，检查当前项目是否在列表中
		if len(projectFilter) > 0 && !projectFilter[projectKey] {
			continue
		}

		// 获取项目名称
		projectName, ok := projectDict[projectKey]
		if !ok {
			continue
		}

		// 构建端口列表（只包含NodePort）
		var portList []models.ServicePort
		for _, port := range svc.Spec.Ports {
			// 跳过没有NodePort的端口
			if port.NodePort == 0 {
				continue
			}

			portInfo := models.ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				NodePort:   port.NodePort,
				Protocol:   string(port.Protocol),
				TargetPort: port.TargetPort.IntVal,
			}
			portList = append(portList, portInfo)
		}

		// 如果没��NodePort端口，跳过这个服务
		if len(portList) == 0 {
			continue
		}

		servicePort := models.ServicePortResponse{
			Project:     projectKey,
			ProjectName: projectName,
			Namespace:   svc.Namespace,
			ServiceName: svc.Name,
			Ports:       portList,
		}
		result = append(result, servicePort)
	}

	utils.Success(w, result)
}

// ClusterStatusRequest 获取集群状态请求结构
type ClusterStatusRequest struct {
	Projects []string `json:"projects"` // 项目列表
}

// GetClusterStatus 获取所有命名空间的Pod状态
func GetClusterStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 从查询参数获取项目列表
	projectsParam := r.URL.Query().Get("projects")
	var projects []string
	if projectsParam != "" {
		projects = strings.Split(projectsParam, ",")
	}

	// 获取k8s客户端
	k8sClient := initData.GetK8s()
	if k8sClient == nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "K8s客户端未初始化")
		return
	}

	// 获取配置
	cfg := initData.GetConfig()
	if cfg == nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取配置失败")
		return
	}

	// 如果指定了项目列表且列表为空，直接返回空结果
	if projects != nil && len(projects) == 0 {
		utils.Success(w, []models.NamespaceStatusResponse{})
		return
	}

	// 构建项目过滤集合
	projectFilter := make(map[string]bool)
	if len(projects) > 0 {
		for _, p := range projects {
			projectFilter[p] = true
		}
	}

	// 获取所有Pod
	pods, err := k8sClient.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取Pod列表失败")
		return
	}

	// 获取所有Namespace
	namespaces, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取Namespace列表失败")
		return
	}

	// 获取项目字典
	db := initData.GetDB()
	var projectDicts []struct {
		Project     string `gorm:"column:project"`
		ProjectName string `gorm:"column:project_name"`
	}
	if err := db.Table("project_dict").Select("project, project_name").Find(&projectDicts).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取项目字典失败")
		return
	}

	// 构建项目字典映射
	projectDict := make(map[string]string)
	for _, dict := range projectDicts {
		projectDict[dict.Project] = dict.ProjectName
	}

	// 记录每个namespace是否有pod
	nsHasPods := make(map[string]bool)
	for _, pod := range pods.Items {
		nsHasPods[pod.Namespace] = true
	}

	// 获取所有namespace的操作记录
	var operations []struct {
		Namespace          string `gorm:"column:namespace"`
		ActionUserName     string `gorm:"column:action_user_name"`
		ActionTime         string `gorm:"column:action_time"`
		ActionTimestamp    string `gorm:"column:action_timestamp"`
		OperatUserName     string `gorm:"column:operat_user_name"`
		OperationTime      string `gorm:"column:operation_time"`
		OperationTimestamp string `gorm:"column:operation_timestamp"`
	}
	if err := db.Table("operations").
		Select("namespace, action_user_name, action_time, action_timestamp, operat_user_name, operation_time, operation_timestamp").
		Find(&operations).Error; err != nil {
		fmt.Printf("获取操作记录失败: %v\n", err)
	}

	// 构建namespace到操作记录的映射
	nsOperations := make(map[string]struct {
		ActionUserName     string
		ActionTime         string
		ActionTimestamp    string
		OperatUserName     string
		OperationTime      string
		OperationTimestamp string
	})
	for _, op := range operations {
		nsOperations[op.Namespace] = struct {
			ActionUserName     string
			ActionTime         string
			ActionTimestamp    string
			OperatUserName     string
			OperationTime      string
			OperationTimestamp string
		}{
			ActionUserName:     op.ActionUserName,
			ActionTime:         op.ActionTime,
			ActionTimestamp:    op.ActionTimestamp,
			OperatUserName:     op.OperatUserName,
			OperationTime:      op.OperationTime,
			OperationTimestamp: op.OperationTimestamp,
		}
	}

	// 构建返回数据
	var result []models.NamespaceStatusResponse
	for _, ns := range namespaces.Items {
		// 跳过包含public的namespace
		if strings.Contains(ns.Name, "public") {
			continue
		}

		// 检查namespace是否包含service
		nsLower := strings.ToLower(ns.Name)
		if !strings.Contains(nsLower, "service") {
			continue
		}

		// 获取项目代码（namespace的第一个部分）
		parts := strings.Split(ns.Name, "-")
		if len(parts) < 2 {
			continue
		}
		projectKey := parts[0]

		// 如果指定了项目列表，检查当前项目是否在列表中
		if len(projectFilter) > 0 && !projectFilter[projectKey] {
			continue
		}

		// 获取项目名称
		projectName, ok := projectDict[projectKey]
		if !ok {
			continue
		}

		// 构建二级域名（取namespace中的前两部分）
		subDomain := fmt.Sprintf("%s%s.%s", parts[0], parts[1], cfg.Server.Domain)

		// 获取操作记录
		var status models.NamespaceStatusResponse
		status.Project = projectKey
		status.ProjectName = projectName
		status.Namespace = ns.Name
		status.SubDomain = subDomain
		status.HasPods = nsHasPods[ns.Name]

		// 添加操作记录
		if op, exists := nsOperations[ns.Name]; exists {
			actionUserName := op.ActionUserName
			actionTime := op.ActionTime
			actionTimestamp := op.ActionTimestamp
			operatUserName := op.OperatUserName
			operationTime := op.OperationTime
			operationTimestamp := op.OperationTimestamp

			status.ActionUserName = &actionUserName
			status.ActionTime = &actionTime
			status.ActionTimestamp = &actionTimestamp
			status.OperatUserName = &operatUserName
			status.OperationTime = &operationTime
			status.OperationTimestamp = &operationTimestamp
		}

		result = append(result, status)
	}

	utils.Success(w, result)
}

// ScalePod 扩缩容接口
func ScalePod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析请求体
	var req models.ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填参数
	if req.Namespace == "" || req.Action == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "命名空间和操作类型不能为空")
		return
	}

	// 验证操作类型
	if req.Action != "scale_up" && req.Action != "scale_down" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "操作类型只能是 scale_up 或 scale_down")
		return
	}

	// 获取k8s客户端
	k8sClient := initData.GetK8s()
	if k8sClient == nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "K8s客户端未初始化")
		return
	}

	// 获取该命名空间下的所有 Deployment
	deployments, err := k8sClient.AppsV1().Deployments(req.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("获取Deployment列表失败: %v", err)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取应用列表失败")
		return
	}

	// 设置副本数
	var replicas int32
	if req.Action == "scale_up" {
		replicas = 1
	} else {
		replicas = 0
	}

	// 遍历更新所有 Deployment
	for _, deployment := range deployments.Items {
		deployment.Spec.Replicas = &replicas
		_, err = k8sClient.AppsV1().Deployments(req.Namespace).Update(context.Background(), &deployment, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("更新Deployment %s 失败: %v", deployment.Name, err)
			continue
		}
	}

	// 记录操作日志
	db := initData.GetDB()
	now := time.Now()
	operation := struct {
		Namespace       string `gorm:"column:namespace"`
		Action          string `gorm:"column:action"`
		ActionUserName  string `gorm:"column:action_user_name"`
		ActionTime      string `gorm:"column:action_time"`
		ActionTimestamp string `gorm:"column:action_timestamp"`
	}{
		Namespace:       req.Namespace,
		Action:          req.Action,
		ActionUserName:  r.Header.Get("Username"),
		ActionTime:      now.Format("2006-01-02 15:04:05"),
		ActionTimestamp: fmt.Sprintf("%d", now.UnixMilli()),
	}
	if err := db.Table("operations").Create(&operation).Error; err != nil {
		log.Printf("记录操作日志失败: %v", err)
	}

	utils.Success(w, map[string]interface{}{
		"namespace": req.Namespace,
		"action":    req.Action,
		"replicas":  replicas,
	})
}

// BatchScalePods 批量扩缩容接口
func BatchScalePods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析请求体
	var req models.BatchScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填参数
	if len(req.Namespaces) == 0 || req.Action == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "命名空间列表和操作类型不能为空")
		return
	}

	// 验证操作类型
	if req.Action != "scale_up" && req.Action != "scale_down" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "操作类型只能是 scale_up 或 scale_down")
		return
	}

	// 立即返回成功响应
	utils.Success(w, map[string]interface{}{
		"message": "批量扩缩容任务已提交",
		"total":   len(req.Namespaces),
	})

	// 启动异步处理
	go func() {
		// 获取k8s客户端
		k8sClient := initData.GetK8s()
		if k8sClient == nil {
			log.Printf("K8s客户端未初始化")
			return
		}

		// 设置批量操作的配置
		config := models.BatchScaleConfig{
			MaxConcurrent:  5,                      // 最大并发数
			IntervalDelay:  time.Millisecond * 100, // 操作间隔
			RequestTimeout: time.Second * 30,       // 单个请求超时时间
		}

		// 创建结果通道
		resultChan := make(chan models.ScaleResult, len(req.Namespaces))

		// 创建信号量来控制并发数
		sem := make(chan struct{}, config.MaxConcurrent)

		// 并发处理每个命名空间
		for _, namespace := range req.Namespaces {
			sem <- struct{}{} // 获取信号量

			go func(ns string) {
				defer func() {
					<-sem // 释放信号量
				}()

				// 设置副本数
				var replicas int32
				if req.Action == "scale_up" {
					replicas = 1
				} else {
					replicas = 0
				}

				// 获取该命名空间下的所有 Deployment
				deployments, err := k8sClient.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
				if err != nil {
					log.Printf("获取命名空间 %s 的Deployment列表失败: %v", ns, err)
					resultChan <- models.ScaleResult{Namespace: ns, Error: err}
					return
				}

				// 更新所有 Deployment
				for _, deployment := range deployments.Items {
					deployment.Spec.Replicas = &replicas
					_, err = k8sClient.AppsV1().Deployments(ns).Update(context.Background(), &deployment, metav1.UpdateOptions{})
					if err != nil {
						log.Printf("更新命名空间 %s 的Deployment %s 失败: %v", ns, deployment.Name, err)
						resultChan <- models.ScaleResult{Namespace: ns, Error: err}
						return
					}
				}

				// 记录操作日志
				db := initData.GetDB()
				now := time.Now()
				operation := struct {
					Namespace       string `gorm:"column:namespace"`
					Action          string `gorm:"column:action"`
					ActionUserName  string `gorm:"column:action_user_name"`
					ActionTime      string `gorm:"column:action_time"`
					ActionTimestamp string `gorm:"column:action_timestamp"`
				}{
					Namespace:       ns,
					Action:          req.Action,
					ActionUserName:  r.Header.Get("Username"),
					ActionTime:      now.Format("2006-01-02 15:04:05"),
					ActionTimestamp: fmt.Sprintf("%d", now.UnixMilli()),
				}
				if err := db.Table("operations").Create(&operation).Error; err != nil {
					log.Printf("记录操作日志失败: %v", err)
				}

				resultChan <- models.ScaleResult{Namespace: ns, Error: nil}
			}(namespace)

			// 添加操作间隔
			time.Sleep(config.IntervalDelay)
		}

		// 收集结果并记录日志
		successCount := 0
		failedCount := 0
		for i := 0; i < len(req.Namespaces); i++ {
			result := <-resultChan
			if result.Error != nil {
				failedCount++
				log.Printf("命名空间 %s 处理失败: %v", result.Namespace, result.Error)
			} else {
				successCount++
				log.Printf("命名空间 %s 处理成功", result.Namespace)
			}
		}

		log.Printf("批量扩缩容完成，成功: %d，失败: %d", successCount, failedCount)
	}()
}
