package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetELFKIndexInfo 使用ELFK查询索引信息（Agent直接返回标准JSON）
func GetELFKIndexInfo(w http.ResponseWriter, _ *http.Request, project, index string) {
	// 使用统一的 AgentClient 调用（GET请求）
	pluginPath := fmt.Sprintf("/proxy/es-plugs/api/elfk/indices?index=%s", index)
	decompressedData, err := middleware.CallAgent(project, pluginPath)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "调用Agent失败: "+err.Error())
		return
	}

	middleware.Logger.Infof("解密后的数据: %s", string(decompressedData))

	// Agent返回的是完整结构：{"code":200,"data":{"indices":[...],"fields":{...}},"message":"..."}
	var agentResponse struct {
		Code int `json:"code"`
		Data struct {
			Indices []string               `json:"indices"`
			Fields  map[string]interface{} `json:"fields"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(decompressedData, &agentResponse); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("解析解密后的数据失败: %v", err))
		return
	}

	if agentResponse.Code != 200 {
		middleware.ResponseError(w, http.StatusInternalServerError, agentResponse.Message)
		return
	}

	if len(agentResponse.Data.Indices) == 0 {
		middleware.ResponseError(w, http.StatusNotFound, "未找到匹配的索引")
		return
	}

	// 返回结果
	finalResult := map[string]interface{}{
		"indices": agentResponse.Data.Indices,
		"fields":  agentResponse.Data.Fields,
	}

	middleware.ResponseSuccess(w, "获取索引映射成功", finalResult)
}
