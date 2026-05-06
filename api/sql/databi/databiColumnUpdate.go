package databi

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"io"
	"net/http"
)

// ColumnUpdateRequest 修改列注释请求
type ColumnUpdateRequest struct {
	Project string   `json:"project"` // 项目名称
	Table   string   `json:"table"`   // 表名，格式：db.table
	ColName []string `json:"colName"` // 列名列表
	Comment []string `json:"comment"` // 注释列表（与 colName 一一对应）
}

// DataBIColumnUpdate 修改列注释
// POST /api/sql/databi/column/update
func DataBIColumnUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	var req ColumnUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求失败")
		return
	}

	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if req.Table == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "table参数不能为空")
		return
	}
	if len(req.ColName) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "colName不能为空")
		return
	}
	if len(req.ColName) != len(req.Comment) {
		middleware.ResponseError(w, http.StatusBadRequest, "colName与comment数量不一致")
		return
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

	// 构建发给 Agent 的请求体（不含 project 字段）
	agentRequest := map[string]interface{}{
		"table":   req.Table,
		"colName": req.ColName,
		"comment": req.Comment,
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/databi-plugs/manage/table/column/update"
	data, err := middleware.CallAgentWithBody(req.Project, pluginPath, agentRequest)
	if err != nil {
		middleware.Logger.Errorf("[DataBI] 修改列注释失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "修改列注释失败: "+err.Error())
		return
	}

	// 解析响应
	var agentResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(data, &agentResp); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}
	if agentResp.Code != 200 {
		middleware.ResponseError(w, http.StatusInternalServerError, "Agent返回错误: "+agentResp.Msg)
		return
	}

	middleware.ResponseSuccess(w, agentResp.Msg, nil)
}
