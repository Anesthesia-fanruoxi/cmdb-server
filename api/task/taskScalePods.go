package task

import (
	"bytes"
	"cmdb/config"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PodScaleRequest 单个Pod缩放请求结构
type PodScaleRequest struct {
	Namespace string `json:"namespace"`
	Action    bool   `json:"action"`
}

// PodBatchScaleRequest 批量Pod缩放请求结构
type PodBatchScaleRequest struct {
	Namespace []string `json:"namespace"`
	Action    bool     `json:"action"`
}

// PodScaleExecutor 单个Pod缩放任务执行器
type PodScaleExecutor struct{}

// Execute 实现TaskExecutor接口，执行单个Pod缩放任务
func (e *PodScaleExecutor) Execute(_ context.Context, task *Task) error {
	middleware.Logger.Infof("开始执行Pod缩放任务: %s", task.ID)

	// 更新任务进度
	task.Progress = 10
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 解析任务参数
	var params PodScaleRequest
	if err := json.Unmarshal(task.Params, &params); err != nil {
		return fmt.Errorf("解析任务参数失败: %v", err)
	}

	// 第1步：开始处理(处理请求)
	actionDesc := "缩容"
	if params.Action {
		actionDesc = "扩容"
	}

	task.Setup.Setup1.Status = "running"
	task.Setup.Setup1.Description = "开始处理"
	task.Setup.Setup1.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 验证参数
	if params.Namespace == "" {
		task.Setup.Setup1.Status = "failed"
		task.Setup.Setup1.EndTime = time.Now()
		task.Setup.Setup1.Duration = task.Setup.Setup1.EndTime.Sub(task.Setup.Setup1.StartTime).Milliseconds()
		task.Setup.Setup1.Description = "命名空间不能为空"
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("命名空间不能为空")
	}

	// 完成步骤1
	task.Progress = 25
	task.Setup.Setup1.Status = "success"
	task.Setup.Setup1.EndTime = time.Now()
	task.Setup.Setup1.Duration = task.Setup.Setup1.EndTime.Sub(task.Setup.Setup1.StartTime).Milliseconds()
	task.Setup.Setup1.Description = "开始处理"
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第2步：开始扩容/缩容(准备调用代理)
	task.Setup.Setup2.Status = "running"
	task.Setup.Setup2.Description = fmt.Sprintf("开始%s", actionDesc)
	task.Setup.Setup2.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 完成步骤2
	task.Progress = 50
	task.Setup.Setup2.Status = "success"
	task.Setup.Setup2.EndTime = time.Now()
	task.Setup.Setup2.Duration = task.Setup.Setup2.EndTime.Sub(task.Setup.Setup2.StartTime).Milliseconds()
	task.Setup.Setup2.Description = fmt.Sprintf("开始%s", actionDesc)
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第3步：正在扩容/缩容(已经调用代理，等待返回)
	task.Setup.Setup3.Status = "running"
	task.Setup.Setup3.Description = fmt.Sprintf("正在%s", actionDesc)
	task.Setup.Setup3.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 执行Pod缩放操作
	result, err := executePodScale(params)
	if err != nil {
		task.Setup.Setup3.Status = "failed"
		task.Setup.Setup3.EndTime = time.Now()
		task.Setup.Setup3.Duration = task.Setup.Setup3.EndTime.Sub(task.Setup.Setup3.StartTime).Milliseconds()
		task.Setup.Setup3.Description = fmt.Sprintf("正在%s失败: %v", actionDesc, err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("执行Pod%s操作失败: %v", actionDesc, err)
	}

	// 完成步骤3
	task.Progress = 75
	task.Setup.Setup3.Status = "success"
	task.Setup.Setup3.EndTime = time.Now()
	task.Setup.Setup3.Duration = task.Setup.Setup3.EndTime.Sub(task.Setup.Setup3.StartTime).Milliseconds()
	task.Setup.Setup3.Description = fmt.Sprintf("正在%s", actionDesc)
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第4步：任务完成(结果返回)
	task.Setup.Setup4.Status = "running"
	task.Setup.Setup4.Description = "任务完成"
	task.Setup.Setup4.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 保存结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		task.Setup.Setup4.Status = "failed"
		task.Setup.Setup4.EndTime = time.Now()
		task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
		task.Setup.Setup4.Description = fmt.Sprintf("序列化结果失败: %v", err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	// 更新任务结果
	if err := GlobalTaskManager.UpdateTaskResult(task.ID, string(resultJSON)); err != nil {
		task.Setup.Setup4.Status = "failed"
		task.Setup.Setup4.EndTime = time.Now()
		task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
		task.Setup.Setup4.Description = fmt.Sprintf("更新任务结果失败: %v", err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("更新任务结果失败: %v", err)
	}

	// 完成步骤4
	task.Progress = 100
	task.Setup.Setup4.Status = "success"
	task.Setup.Setup4.EndTime = time.Now()
	task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
	task.Setup.Setup4.Description = "任务完成"
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	middleware.Logger.Infof("Pod缩放任务执行完成: %s", task.ID)
	return nil
}

// PodBatchScaleExecutor 批量Pod缩放任务执行器
type PodBatchScaleExecutor struct{}

// Execute 实现TaskExecutor接口，执行批量Pod缩放任务
func (e *PodBatchScaleExecutor) Execute(_ context.Context, task *Task) error {
	middleware.Logger.Infof("开始执行批量Pod缩放任务: %s", task.ID)

	// 更新任务进度
	task.Progress = 10
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 解析任务参数
	var params PodBatchScaleRequest
	if err := json.Unmarshal(task.Params, &params); err != nil {
		return fmt.Errorf("解析任务参数失败: %v", err)
	}

	// 第1步：开始处理(处理请求)
	actionDesc := "缩容"
	if params.Action {
		actionDesc = "扩容"
	}

	task.Setup.Setup1.Status = "running"
	task.Setup.Setup1.Description = "开始处理"
	task.Setup.Setup1.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 验证参数
	if len(params.Namespace) == 0 {
		task.Setup.Setup1.Status = "failed"
		task.Setup.Setup1.EndTime = time.Now()
		task.Setup.Setup1.Duration = task.Setup.Setup1.EndTime.Sub(task.Setup.Setup1.StartTime).Milliseconds()
		task.Setup.Setup1.Description = "命名空间列表不能为空"
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("命名空间列表不能为空")
	}

	// 完成步骤1
	task.Progress = 25
	task.Setup.Setup1.Status = "success"
	task.Setup.Setup1.EndTime = time.Now()
	task.Setup.Setup1.Duration = task.Setup.Setup1.EndTime.Sub(task.Setup.Setup1.StartTime).Milliseconds()
	task.Setup.Setup1.Description = "开始处理"
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第2步：开始扩容/缩容(准备调用代理)
	task.Setup.Setup2.Status = "running"
	task.Setup.Setup2.Description = fmt.Sprintf("开始%s", actionDesc)
	task.Setup.Setup2.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 完成步骤2
	task.Progress = 50
	task.Setup.Setup2.Status = "success"
	task.Setup.Setup2.EndTime = time.Now()
	task.Setup.Setup2.Duration = task.Setup.Setup2.EndTime.Sub(task.Setup.Setup2.StartTime).Milliseconds()
	task.Setup.Setup2.Description = fmt.Sprintf("开始%s", actionDesc)
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第3步：正在扩容/缩容(已经调用代理，等待返回)
	task.Setup.Setup3.Status = "running"
	task.Setup.Setup3.Description = fmt.Sprintf("正在%s", actionDesc)
	task.Setup.Setup3.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 执行批量Pod缩放操作
	result, err := executePodBatchScale(params)
	if err != nil {
		task.Setup.Setup3.Status = "failed"
		task.Setup.Setup3.EndTime = time.Now()
		task.Setup.Setup3.Duration = task.Setup.Setup3.EndTime.Sub(task.Setup.Setup3.StartTime).Milliseconds()
		task.Setup.Setup3.Description = fmt.Sprintf("正在%s失败: %v", actionDesc, err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("执行Pod%s操作失败: %v", actionDesc, err)
	}

	// 完成步骤3
	task.Progress = 75
	task.Setup.Setup3.Status = "success"
	task.Setup.Setup3.EndTime = time.Now()
	task.Setup.Setup3.Duration = task.Setup.Setup3.EndTime.Sub(task.Setup.Setup3.StartTime).Milliseconds()
	task.Setup.Setup3.Description = fmt.Sprintf("正在%s", actionDesc)
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 第4步：任务完成(结果返回)
	task.Setup.Setup4.Status = "running"
	task.Setup.Setup4.Description = "任务完成"
	task.Setup.Setup4.StartTime = time.Now()
	if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
		middleware.Logger.Errorf("更新任务状态失败: %v", err)
	}

	// 保存结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		task.Setup.Setup4.Status = "failed"
		task.Setup.Setup4.EndTime = time.Now()
		task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
		task.Setup.Setup4.Description = fmt.Sprintf("序列化结果失败: %v", err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	// 更新任务结果
	if err := GlobalTaskManager.UpdateTaskResult(task.ID, string(resultJSON)); err != nil {
		task.Setup.Setup4.Status = "failed"
		task.Setup.Setup4.EndTime = time.Now()
		task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
		task.Setup.Setup4.Description = fmt.Sprintf("更新任务结果失败: %v", err)
		if err := GlobalTaskManager.saveTaskStatusToRedis(task); err != nil {
			middleware.Logger.Errorf("更新任务状态失败: %v", err)
		}
		return fmt.Errorf("更新任务结果失败: %v", err)
	}

	// 完成步骤4
	task.Progress = 100
	task.Setup.Setup4.Status = "success"
	task.Setup.Setup4.EndTime = time.Now()
	task.Setup.Setup4.Duration = task.Setup.Setup4.EndTime.Sub(task.Setup.Setup4.StartTime).Milliseconds()
	task.Setup.Setup4.Description = "任务完成"
	if err := GlobalTaskManager.UpdateTaskProgress(task.ID, task.Progress); err != nil {
		middleware.Logger.Errorf("更新任务进度失败: %v", err)
	}

	middleware.Logger.Infof("批量Pod缩放任务执行完成: %s", task.ID)
	return nil
}

// executePodScale 执行单个Pod缩放操作
// 这个函数从api/assets/kubePod/operationStatus.go中的K8sPodScale转换而来
func executePodScale(request PodScaleRequest) (interface{}, error) {
	middleware.Logger.Infof("执行Pod缩放操作：命名空间=%s, 操作=%v", request.Namespace, request.Action)

	// 从配置中获取k8s专用密钥
	k8sSecret := config.GetConfig().K8s.Secret
	if k8sSecret == "" {
		return nil, fmt.Errorf("未配置K8s密钥")
	}

	// 加密请求数据
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}
	middleware.Logger.Infof("加密前的数据：%s", requestData)

	encryptedData, err := middleware.CompressAndEncrypt(requestData, k8sSecret)
	if err != nil {
		return nil, fmt.Errorf("加密请求数据失败: %v", err)
	}

	// 构建发送给代理的请求体
	proxyRequestBody := map[string]string{
		"data": encryptedData,
	}
	proxyRequestData, err := json.Marshal(proxyRequestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化代理请求数据失败: %v", err)
	}

	// 从配置文件获取k8s端点
	k8sEndpoint := config.GetConfig().K8s.Endpoint
	if k8sEndpoint == "" {
		return nil, fmt.Errorf("未配置K8s端点")
	}

	// 构建API请求URL
	apiURL := fmt.Sprintf("%s/api/pod/scale", k8sEndpoint)
	middleware.Logger.Infof("发送请求到: %s", apiURL)

	// 发送请求到代理服务器
	client := &http.Client{}
	proxyReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(proxyRequestData))
	if err != nil {
		return nil, fmt.Errorf("创建代理请求失败: %v", err)
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("调用K8s API失败: %v", err)
	}
	defer proxyResp.Body.Close()

	// 读取代理响应
	proxyRespBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取代理响应失败: %v", err)
	}
	middleware.Logger.Infof("接收到的初始数据：%s", proxyRespBody)

	// 解析代理响应中的加密数据
	var proxyRespData struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(proxyRespBody, &proxyRespData); err != nil {
		return nil, fmt.Errorf("解析代理响应失败: %v", err)
	}
	middleware.Logger.Infof("解析后的数据：%s", proxyRespData)

	// 解密代理响应数据
	decryptedData, err := middleware.DecryptAndDecompress(proxyRespData.Data, k8sSecret)
	if err != nil {
		return nil, fmt.Errorf("解密代理响应失败: %v", err)
	}

	// 解析解密后的数据为JSON对象
	var jsonData interface{}
	if err := json.Unmarshal(decryptedData, &jsonData); err != nil {
		return nil, fmt.Errorf("解析解密后的JSON数据失败: %v", err)
	}

	middleware.Logger.Infof("解密后的数据：%v", jsonData)
	return jsonData, nil
}

// executePodBatchScale 执行批量Pod缩放操作
// 这个函数从api/assets/kubePod/operationStatus.go中的K8sPodBatchScale转换而来
func executePodBatchScale(request PodBatchScaleRequest) (interface{}, error) {
	middleware.Logger.Infof("执行批量Pod缩放操作：命名空间=%v, 操作=%v", request.Namespace, request.Action)

	// 从配置中获取k8s专用密钥
	k8sSecret := config.GetConfig().K8s.Secret
	if k8sSecret == "" {
		return nil, fmt.Errorf("未配置K8s密钥")
	}

	// 加密请求数据
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}
	middleware.Logger.Infof("加密前的数据：%s", requestData)

	encryptedData, err := middleware.CompressAndEncrypt(requestData, k8sSecret)
	if err != nil {
		return nil, fmt.Errorf("加密请求数据失败: %v", err)
	}

	// 构建发送给代理的请求体
	proxyRequestBody := map[string]string{
		"data": encryptedData,
	}
	proxyRequestData, err := json.Marshal(proxyRequestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化代理请求数据失败: %v", err)
	}

	// 从配置文件获取k8s端点
	k8sEndpoint := config.GetConfig().K8s.Endpoint
	if k8sEndpoint == "" {
		return nil, fmt.Errorf("未配置K8s端点")
	}

	// 构建API请求URL
	apiURL := fmt.Sprintf("%s/api/pod/batch-scale", k8sEndpoint)
	middleware.Logger.Infof("发送请求到: %s", apiURL)

	// 发送请求到代理服务器
	client := &http.Client{}
	proxyReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(proxyRequestData))
	if err != nil {
		return nil, fmt.Errorf("创建代理请求失败: %v", err)
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("调用K8s API失败: %v", err)
	}
	defer proxyResp.Body.Close()

	// 读取代理响应
	proxyRespBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取代理响应失败: %v", err)
	}
	middleware.Logger.Infof("接收到的初始数据：%s", proxyRespBody)

	// 解析代理响应中的加密数据
	var proxyRespData struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(proxyRespBody, &proxyRespData); err != nil {
		return nil, fmt.Errorf("解析代理响应失败: %v", err)
	}
	middleware.Logger.Infof("解析后的数据：%s", proxyRespData)

	// 解密代理响应数据
	decryptedData, err := middleware.DecryptAndDecompress(proxyRespData.Data, k8sSecret)
	if err != nil {
		return nil, fmt.Errorf("解密代理响应失败: %v", err)
	}

	// 解析解密后的数据为JSON对象
	var jsonData interface{}
	if err := json.Unmarshal(decryptedData, &jsonData); err != nil {
		return nil, fmt.Errorf("解析解密后的JSON数据失败: %v", err)
	}

	middleware.Logger.Infof("解密后的数据：%v", jsonData)
	return jsonData, nil
}

// 注册Pod任务执行器
func init() {
	// 等待GlobalTaskManager初始化
	// 这在程序启动时会被InitTaskSystem函数调用
	go func() {
		// 等待任务管理器初始化完成
		for GlobalTaskManager == nil {
			time.Sleep(100 * time.Millisecond)
		}

		// 注册Pod缩放任务执行器
		GlobalTaskManager.RegisterExecutor("pod_scale", &PodScaleExecutor{})
		GlobalTaskManager.RegisterExecutor("pod_batch_scale", &PodBatchScaleExecutor{})

		middleware.Logger.Info("Pod缩放任务执行器注册成功")
	}()
}

// 注册Pod任务执行器
func RegisterPodExecutors() {
	if GlobalTaskManager == nil {
		middleware.Logger.Error("TaskManager未初始化，无法注册Pod执行器")
		return
	}

	GlobalTaskManager.RegisterExecutor("pod_scale", &PodScaleExecutor{})
	GlobalTaskManager.RegisterExecutor("pod_batch_scale", &PodBatchScaleExecutor{})
	middleware.Logger.Info("Pod缩放任务执行器注册成功，类型: pod_scale, pod_batch_scale")
}
