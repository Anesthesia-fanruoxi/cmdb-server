package key

import (
	"cmdb/config"
	"cmdb/models"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// KeyDecrypt 处理数据解密请求
func KeyDecrypt(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取查询参数
	encryptedData := r.URL.Query().Get("encryptedData")
	if encryptedData == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "encryptedData参数不能为空")
		return
	}

	project := r.URL.Query().Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
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
		RecordAudit(r, "decrypt", project, 1, "failed")
		middleware.ResponseError(w, http.StatusForbidden, "无权访问密钥服务")
		return
	}

	// 验证当前请求的项目是否在可访问列表中
	if !isSuperAdmin && !utils.Contains(projects, project) {
		RecordAudit(r, "decrypt", project, 1, "failed")
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 使用传入的project参数作为agent项目
	agentProject := project

	// 获取agent地址
	agentURL, err := GetAgentURL(agentProject)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取代理地址失败: "+err.Error())
		return
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 构建完整的API URL - 使用查询参数而不是请求体
	apiURL := fmt.Sprintf("%s/proxy/key-plugs/decrypt?encryptedData=%s", agentURL, url.QueryEscape(encryptedData))
	middleware.Logger.Infof("构建完整的API URL: %s", apiURL)
	// 创建GET请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建HTTP请求失败")
		return
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "发送请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取响应失败")
		return
	}

	// 获取任务加密盐
	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	middleware.Logger.Infof("使用taskSalt解密代理响应，taskSalt长度: %d", len(taskSalt))

	// 记录原始响应
	middleware.Logger.Infof("代理返回原始响应，状态码: %d, 响应体: %s", resp.StatusCode, string(respBody))

	// 先尝试解析响应体，无论状态码如何
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	middleware.Logger.Infof("解析后的响应结构: %+v", result)

	// 检查是否包含data字段（无论状态码如何）
	if encryptedRespData, ok := result["data"].(string); ok && encryptedRespData != "" {
		middleware.Logger.Infof("开始解密代理响应data字段，加密数据长度: %d", len(encryptedRespData))

		// 解密并解压数据 - 这里解密的是代理的响应体
		decompressedData, err := middleware.DecryptAndDecompress(encryptedRespData, taskSalt)
		if err != nil {
			middleware.Logger.Errorf("解密解压失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "解密解压失败: "+err.Error())
			return
		}

		middleware.Logger.Infof("解密解压成功，解密后数据: %s", string(decompressedData))

		// 解密后得到的是代理处理的结果（这里是解密后的明文）
		var decryptedResult string
		if err := json.Unmarshal(decompressedData, &decryptedResult); err != nil {
			// 如果不是JSON字符串，直接使用字符串（这是正常情况）
			decryptedResult = string(decompressedData)
			middleware.Logger.Infof("使用原始解密数据: %s", decryptedResult)
		} else {
			middleware.Logger.Infof("JSON解析成功，解密结果: %s", decryptedResult)
		}

		// 构造标准格式的响应
		response := models.DecryptResponse{
			Data: decryptedResult,
		}

		// 记录成功审计
		RecordAudit(r, "decrypt", agentProject, 1, "success")

		// 返回结果
		middleware.ResponseSuccess(w, "解密成功", response)
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
		RecordAudit(r, "decrypt", agentProject, 1, "failed")

		middleware.ResponseError(w, resp.StatusCode, errorMsg)
	}
}
