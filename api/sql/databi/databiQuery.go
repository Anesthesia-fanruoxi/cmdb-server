package databi

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
)

// QueryRequest 查询请求结构
type QueryRequest struct {
	Project string `json:"project"` // 项目名称
	Context string `json:"context"` // SQL 语句
	Type    string `json:"type"`    // 固定为 "sql"
}

// QueryResponse 查询响应结构
type QueryResponse struct {
	Head  []string                 `json:"head"`  // 列名
	Table []map[string]interface{} `json:"table"` // 数据行
	File  interface{}              `json:"file"`  // 文件信息(通常为null)
}

// DataBIQuery 执行 Spark SQL 查询
// POST /api/sql/databi/query
func DataBIQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析请求
	var req QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求失败")
		return
	}

	// 验证参数
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if req.Context == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "context参数不能为空")
		return
	}

	// 默认 type 为 sql
	if req.Type == "" {
		req.Type = "sql"
	}

	// 验证项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !isSuperAdmin && !utils.Contains(projects, req.Project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 构建请求体(发送给 Agent)
	agentRequest := map[string]string{
		"context": req.Context,
		"type":    req.Type,
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/databi-plugs/spark/common/query"
	decryptedData, err := middleware.CallAgentWithBody(req.Project, pluginPath, agentRequest)
	if err != nil {
		middleware.Logger.Errorf("调用Agent失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "调用Agent失败: "+err.Error())
		return
	}

	// 解析响应
	var agentResp struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data QueryResponse `json:"data"`
	}
	if err := json.Unmarshal(decryptedData, &agentResp); err != nil {
		middleware.Logger.Errorf("解析Agent响应失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	// 检查响应状态
	if agentResp.Code != 200 {
		middleware.ResponseError(w, http.StatusInternalServerError, "Agent返回错误: "+agentResp.Msg)
		return
	}

	// 返回数据
	middleware.ResponseSuccess(w, agentResp.Msg, agentResp.Data)
}
