package key

import (
	"bytes"
	"cmdb/config"
	"cmdb/models"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// KeyBatchDecrypt 处理批量数据解密请求
// 限制：一次最多解密20个数据项
func KeyBatchDecrypt(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭请求体失败: %v", closeErr)
		}
	}()

	// 记录原始请求体以便调试
	middleware.Logger.Infof("批量解密接口收到请求体: %s", string(body))

	// 解析JSON请求
	var batchReq models.BatchDecryptRequest
	if err := json.Unmarshal(body, &batchReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("解析JSON失败: %v, 请求体: %s", err, string(body)))
		return
	}

	middleware.Logger.Infof("解析成功，project: %s, 数据数量: %d", batchReq.Project, len(batchReq.Data))

	// 验证必填字段
	if batchReq.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	if len(batchReq.Data) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "加密数据数组不能为空")
		return
	}

	if len(batchReq.Data) > 20 {
		RecordAudit(r, "batchDecrypt", batchReq.Project, len(batchReq.Data), "failed")
		middleware.ResponseError(w, http.StatusBadRequest, "批量解密一次最多支持20个数据，当前提交了"+fmt.Sprintf("%d", len(batchReq.Data))+"个，请分批处理")
		return
	}

	// 获取项目列表并验证当前请求的项目是否在可访问列表中
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果不是超级管理员且没有可访问的项目，返回无权限
	if !isSuperAdmin && len(projects) == 0 {
		RecordAudit(r, "batchDecrypt", batchReq.Project, len(batchReq.Data), "failed")
		middleware.ResponseError(w, http.StatusForbidden, "无权访问密钥服务")
		return
	}

	// 验证当前请求的项目是否在可访问列表中
	if !isSuperAdmin && !utils.Contains(projects, batchReq.Project) {
		RecordAudit(r, "batchDecrypt", batchReq.Project, len(batchReq.Data), "failed")
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 使用传入的project参数作为agent项目
	agentProject := batchReq.Project

	// 获取agent地址
	agentURL, err := GetAgentURL(agentProject)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取代理地址失败: "+err.Error())
		return
	}

	// 构建请求数据 - 代理期望直接的数组格式，不需要包装
	// 将数组直接转换为JSON
	requestDataJSON, err := json.Marshal(batchReq.Data)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化请求数据失败")
		return
	}

	// 获取任务加密盐
	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	if taskSalt == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "任务加密盐未配置")
		return
	}

	// 压缩并加密请求数据
	encryptedRequestData, err := middleware.CompressAndEncrypt(requestDataJSON, taskSalt)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "加密压缩请求数据失败: "+err.Error())
		return
	}

	// 构建加密后的请求体
	encryptedRequest := map[string]interface{}{
		"data": encryptedRequestData,
	}

	// 将加密后的请求体转换为JSON
	encryptedRequestJSON, err := json.Marshal(encryptedRequest)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化加密请求数据失败")
		return
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 构建完整的API URL
	apiURL := fmt.Sprintf("%s/proxy/key-plugs/batchDecrypt", agentURL)

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(encryptedRequestJSON))
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建HTTP请求失败")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "发送请求失败: "+err.Error())
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭响应体失败: %v", closeErr)
		}
	}()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取响应失败")
		return
	}

	// 先尝试解析响应体，无论状态码如何
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	// 检查是否包含data字段（无论状态码如何）
	if encryptedRespData, ok := result["data"].(string); ok && encryptedRespData != "" {

		// 解密并解压数据
		decompressedData, err := middleware.DecryptAndDecompress(encryptedRespData, taskSalt)
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "解密解压失败: "+err.Error())
			return
		}

		// 解析批量解密结果
		// 先尝试解析为JSON格式
		var batchResults []map[string]interface{}
		if err := json.Unmarshal(decompressedData, &batchResults); err != nil {
			// 如果不是JSON格式，可能是直接的字符串结果
			// 记录解密后的原始数据以便调试
			middleware.Logger.Infof("批量解密返回的原始数据: %s", string(decompressedData))

			// 尝试解析为字符串格式的响应
			var stringResult string
			if err2 := json.Unmarshal(decompressedData, &stringResult); err2 != nil {
				middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("解析批量解密结果失败，原始数据: %s, JSON错误: %v, 字符串错误: %v", string(decompressedData), err, err2))
				return
			}

			// 如果是字符串格式，尝试再次解析
			if err3 := json.Unmarshal([]byte(stringResult), &batchResults); err3 != nil {
				middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("二次解析批量解密结果失败，字符串内容: %s, 错误: %v", stringResult, err3))
				return
			}
		}

		// 构造标准格式的响应
		var response models.BatchDecryptResponse
		for i, item := range batchResults {
			encrypted, okEncrypted := item["encrypted"].(string)
			decrypted, okDecrypted := item["decrypted"].(string)

			if okEncrypted && okDecrypted {
				response = append(response, models.BatchDecryptItem{
					Encrypted: encrypted,
					Decrypted: decrypted,
				})
			} else {
				middleware.Logger.Warnf("批量解密第%d项格式错误: %+v", i, item)
			}
		}

		// 记录成功审计
		RecordAudit(r, "batchDecrypt", batchReq.Project, len(batchReq.Data), "success")

		// 返回结果
		middleware.ResponseSuccess(w, "批量解密成功", response)
	} else {
		// 如果没有data字段或data为空，说明是真正的错误响应
		middleware.Logger.Errorf("代理返回错误响应，状态码: %d, 无有效data字段", resp.StatusCode)

		// 提取错误信息
		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		} else if errStr, ok := result["error"].(string); ok {
			errorMsg = errStr
		} else if message, ok := result["message"].(string); ok {
			errorMsg = message
		} else {
			errorMsg = fmt.Sprintf("未知错误，状态码: %d, 完整响应: %+v", resp.StatusCode, result)
		}

		// 记录失败审计
		RecordAudit(r, "batchDecrypt", batchReq.Project, len(batchReq.Data), "failed")

		middleware.ResponseError(w, resp.StatusCode, errorMsg)
	}
}
