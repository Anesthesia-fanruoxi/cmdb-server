package kubeService

import (
	"bytes"
	"cmdb/config"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ServiceRequest 表示请求参数
type NSRequest struct {
	Project string `json:"project"`
	Type    string `json:"type"`
}

// GetPodService 获取Pod相关的服务列表
func GetNS(w http.ResponseWriter, r *http.Request) {
	// 1. 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 2. 解析客户端请求
	var requestBody NSRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求数据格式")
		return
	}

	middleware.Logger.Infof("接受到的请求：%v", requestBody)

	// 从配置中获取k8s专用密钥
	k8sSecret := config.GetConfig().K8s.Secret
	if k8sSecret == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "未配置K8s密钥")
		return
	}

	// 3. 加密压缩请求数据
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

	// 4. 构建发送给代理的请求体
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
	apiURL := fmt.Sprintf("%s/api/namespaces", k8sEndpoint)

	// 5. 发送请求到代理服务器
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

	// 6. 读取代理响应
	proxyRespBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取代理响应失败")
		return
	}
	middleware.Logger.Infof("接受到的数据为：%s", proxyRespBody)
	// 7. 解析代理响应中的加密数据
	var proxyRespData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(proxyRespBody, &proxyRespData); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析代理响应失败")
		return
	}

	// 检查代理响应状态
	if proxyRespData.Code != http.StatusOK {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("代理服务返回错误: %s", proxyRespData.Msg))
		return
	}
	middleware.Logger.Infof("解析后的加密内容为：%s", proxyRespData.Data)
	// 8. 解密代理响应数据
	decryptedData, err := middleware.DecryptAndDecompress(proxyRespData.Data, k8sSecret)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解密代理响应失败")
		return
	}
	middleware.Logger.Infof("解密后的数据为：%s", decryptedData)

	// 8. 解析解密后的数据为JSON对象
	var jsonData interface{}
	if err := json.Unmarshal(decryptedData, &jsonData); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析解密后的JSON数据失败")
		return
	}
	// 9. 返回响应给客户端
	middleware.ResponseSuccess(w, "success", jsonData)
}
