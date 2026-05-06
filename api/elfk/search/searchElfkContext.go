package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SearchELFKContext 处理ELFK上下文查询
func SearchELFKContext(w http.ResponseWriter, bodyData []byte, startTime time.Time) {
	// 解析JSON请求
	var contextReq map[string]interface{}
	if err := json.Unmarshal(bodyData, &contextReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}

	// 获取project参数
	project, ok := contextReq["project"].(string)
	if !ok || project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 验证doc_id参数
	if _, ok := contextReq["doc_id"].(string); !ok {
		middleware.ResponseError(w, http.StatusBadRequest, "doc_id参数不能为空")
		return
	}

	// 从原始请求中移除project字段，其他参数保持不变
	delete(contextReq, "project")

	// 确保_source字段存在
	contextReq["_source"] = true

	// 序列化请求体用于日志
	reqJSON, _ := json.Marshal(contextReq)
	middleware.Logger.Infof("上下文查询请求(明文): %s", string(reqJSON))

	// 使用统一的 AgentClient 调用
	pluginPath := "/proxy/es-plugs/api/elfk/context"
	decompressedData, err := middleware.CallAgentWithBody(project, pluginPath, contextReq)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "调用Agent失败: "+err.Error())
		return
	}

	// 解析解密后的JSON数据（Agent新格式）
	var agentResponse struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(decompressedData, &agentResponse); err != nil {
		middleware.Logger.Errorf("解析解密后的数据失败: %v, 数据: %s", err, string(decompressedData))
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("解析解密后的数据失败: %v", err))
		return
	}

	// 检查Agent返回的状态码
	if agentResponse.Code != 200 {
		middleware.Logger.Errorf("Agent上下文查询失败: %s", agentResponse.Message)
		middleware.ResponseError(w, http.StatusInternalServerError, agentResponse.Message)
		return
	}

	// 添加处理时间到结果中
	agentResponse.Data["took_ms"] = time.Since(startTime).Milliseconds()

	// 处理返回结构，保持与SLS返回格式一致
	if _, ok := agentResponse.Data["took"]; ok {
		delete(agentResponse.Data, "took")
	}

	// 返回结果（只返回data部分）
	middleware.ResponseSuccess(w, "查询成功", agentResponse.Data)
}
