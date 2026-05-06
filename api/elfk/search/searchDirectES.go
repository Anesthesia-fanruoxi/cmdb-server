package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
)

// GetEsFieldsDirectly 直接从ES获取字段信息，不通过HTTP接口
// 这个方法专为系统初始化阶段设计，避免依赖HTTP上下文和用户认证
func GetEsFieldsDirectly(agentURL, indexPattern string) (interface{}, interface{}, error) {
	// 注意：这里的 agentURL 参数实际上应该是 project，因为我们使用 CallAgent
	// 但为了保持向后兼容，我们需要从 agentURL 反推 project
	// 实际上这个函数的调用者应该传递 project 而不是 agentURL

	// 临时方案：直接使用 agentURL 作为 project（需要调用者修改）
	project := agentURL

	// 步骤1: 获取索引列表
	indicesPattern := fmt.Sprintf("%s", indexPattern) // 例如: logs-*
	pluginPath := fmt.Sprintf("/proxy/es-plugs/api/elfk/indices?index=%s", indicesPattern)

	// 使用统一的 AgentClient 调用（GET请求）
	decompressedIndicesData, err := middleware.CallAgent(project, pluginPath)
	if err != nil {
		return nil, nil, fmt.Errorf("调用Agent获取索引列表失败: %v", err)
	}

	// Agent返回的是完整结构：{"code":200,"data":{"indices":[...],"fields":{...}},"message":"..."}
	var indicesResponse struct {
		Code int `json:"code"`
		Data struct {
			Indices []string               `json:"indices"`
			Fields  map[string]interface{} `json:"fields"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(decompressedIndicesData, &indicesResponse); err != nil {
		return nil, nil, fmt.Errorf("解析索引数据失败: %v", err)
	}

	if indicesResponse.Code != 200 {
		return nil, nil, fmt.Errorf("Agent返回错误: %s", indicesResponse.Message)
	}

	// 如果索引列表为空
	if len(indicesResponse.Data.Indices) == 0 {
		return nil, nil, fmt.Errorf("未找到匹配的索引")
	}

	// Agent的indices接口已经同时返回了indices和fields，直接使用
	return indicesResponse.Data.Fields, indicesResponse.Data.Indices, nil
}
