package kubePod

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type StatusRequest struct {
	Project string `json:"project,omitempty"`
}

// K8sGetStatus 处理 api/namespace/status 请求，获取k8s集群状态信息
func K8sGetStatus(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 1. 解析客户端原始请求
	var requestBody StatusRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求数据格式")
		return
	}

	// 默认使用"*"表示查询所有项目
	if requestBody.Project == "" {
		requestBody.Project = "*"
	}

	middleware.Logger.Infof("接受到的请求：%v", requestBody)

	// 从配置中获取k8s专用密钥
	k8sSecret := config.GetConfig().K8s.Secret
	if k8sSecret == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "未配置K8s密钥")
		return
	}

	// 2. 加密请求数据
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化请求数据失败")
		return
	}

	encryptedData, err := middleware.CompressAndEncrypt(requestData, k8sSecret)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "加密请求数据失败")
		return
	}

	// 3. 构建发送给代理的请求体
	proxyRequestBody := map[string]string{
		"data": encryptedData,
	}
	proxyRequestData, err := json.Marshal(proxyRequestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化代理请求数据失败")
		return
	}

	// 从配置文件获取k8s端点
	k8sEndpoint := config.GetConfig().K8s.Endpoint
	if k8sEndpoint == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "未配置K8s端点")
		return
	}

	// 构建API请求URL
	apiURL := fmt.Sprintf("%s/api/namespace/status", k8sEndpoint)

	// 4. 发送请求到代理服务器
	client := &http.Client{}
	proxyReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(proxyRequestData))
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建代理请求失败")
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("调用K8s API失败: %v", err))
		return
	}
	defer proxyResp.Body.Close()

	// 5. 读取代理响应
	proxyRespBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取代理响应失败")
		return
	}

	// 6. 解析代理响应中的加密数据
	var proxyRespData struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(proxyRespBody, &proxyRespData); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析代理响应失败")
		return
	}

	// 7. 解密代理响应数据
	decryptedData, err := middleware.DecryptAndDecompress(proxyRespData.Data, k8sSecret)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解密代理响应失败")
		return
	}

	// 8. 解析解密后的数据为JSON对象
	var jsonData interface{}
	if err := json.Unmarshal(decryptedData, &jsonData); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析解密后的JSON数据失败")
		return
	}

	// 一次性加载所有项目字典数据
	var projectDicts []system.ProjectDict
	err = database.DB.Table(middleware.TableProjectDict).Find(&projectDicts).Error
	if err != nil {
		middleware.Logger.Warnf("加载项目字典表失败: %v", err)
	}

	// 构建项目字典映射，便于快速查找
	projectDictMap := make(map[string]string)
	for _, dict := range projectDicts {
		projectDictMap[dict.Project] = dict.ProjectName
	}

	// 处理返回的数据，为每个项目添加project_name和domain字段
	count := 0         // 总数量
	activeCount := 0   // 运行中数量
	inactiveCount := 0 // 停止中数量

	switch data := jsonData.(type) {
	case []interface{}:
		// 数据是数组类型
		count = len(data) // 设置总数量

		for i, item := range data {
			if projectItem, ok := item.(map[string]interface{}); ok {
				// 获取项目代码
				projectCode, _ := projectItem["project"].(string)
				if projectCode != "" {
					// 从内存中查找项目名称
					projectName := ""

					// 直接匹配
					if name, exists := projectDictMap[projectCode]; exists {
						projectName = name
					} else {
						// 尝试去掉"-service"后缀
						trimmedCode := strings.TrimSuffix(projectCode, "-service")
						if name, exists := projectDictMap[trimmedCode]; exists {
							projectName = name
						}
					}

					// 生成域名，包含环境名称
					domain := ""
					// 获取命名空间，格式通常为：project-env-service
					namespace, _ := projectItem["namespace"].(string)
					if namespace != "" {
						// 提取环境名称
						envName := ""
						parts := strings.Split(namespace, "-")
						if len(parts) >= 2 {
							// 假设倒数第二个部分是环境名
							if parts[len(parts)-1] == "service" && len(parts) >= 3 {
								envName = parts[len(parts)-2]
							} else {
								envName = parts[len(parts)-1]
							}
						}

						// 去除项目代码中的"-service"后缀
						domainPrefix := strings.TrimSuffix(projectCode, "-service")
						// 去除所有"-"
						domainPrefix = strings.ReplaceAll(domainPrefix, "-", "")

						// 拼接域名: project + env + domain
						if domainPrefix != "" && envName != "" {
							domain = fmt.Sprintf("%s%s.hzbxhd.com", domainPrefix, envName)
						}
					} else {
						// 如果没有命名空间信息，使用原来的逻辑
						domainPrefix := strings.TrimSuffix(projectCode, "-service")
						domainPrefix = strings.ReplaceAll(domainPrefix, "-", "")
					}

					// 添加字段到项目信息中
					projectItem["project_name"] = projectName
					projectItem["domain"] = domain

					// 高效统计活跃状态
					isActive := false // 默认为非活跃

					// 首先尝试获取is_active字段
					if val, exists := projectItem["is_active"]; exists && val != nil {
						// 根据类型直接判断
						switch v := val.(type) {
						case bool:
							isActive = v
						case float64:
							isActive = (v == 1)
						case int:
							isActive = (v == 1)
						case string:
							isActive = (v == "true" || v == "1")
						}
					} else if status, ok := projectItem["status"].(string); ok {
						// 如果没有is_active字段，但有status字段
						lowerStatus := strings.ToLower(status)
						isActive = (lowerStatus == "running" || lowerStatus == "active")
					}

					// 根据活跃状态更新计数
					if isActive {
						activeCount++
					} else {
						inactiveCount++
					}

					data[i] = projectItem
				}
			}
		}
		jsonData = data
	default:
		middleware.Logger.Warnf("返回数据格式不符合预期，期望数组类型，实际类型: %T", jsonData)
	}

	// 构建包含额外统计信息的响应数据
	responseData := map[string]interface{}{
		"result":         jsonData,
		"count":          count,
		"active_count":   activeCount,
		"inactive_count": inactiveCount,
	}

	// 9. 返回响应给客户端
	middleware.ResponseSuccess(w, "success", responseData)
}
