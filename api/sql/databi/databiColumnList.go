package databi

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// ColumnInfo 表字段信息
type ColumnInfo struct {
	ColName  string  `json:"col_name"`
	DataType string  `json:"data_type"`
	Comment  *string `json:"comment"`
}

// DataBIColumnList 获取表字段信息
// GET /api/sql/databi/column/list?project=xxx&table=db.table_name
func DataBIColumnList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	query := r.URL.Query()
	project := query.Get("project")
	table := query.Get("table")

	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if table == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "table参数不能为空")
		return
	}

	// 验证项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isSuperAdmin && !utils.Contains(projects, project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/databi-plugs/manage/table/column/list?table=" + table
	data, err := middleware.CallAgent(project, pluginPath)
	if err != nil {
		middleware.Logger.Errorf("[DataBI] 获取表字段失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "获取表字段失败: "+err.Error())
		return
	}

	// 解析响应
	var agentResp struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data []ColumnInfo `json:"data"`
	}
	if err := json.Unmarshal(data, &agentResp); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}
	if agentResp.Code != 200 {
		middleware.ResponseError(w, http.StatusInternalServerError, "Agent返回错误: "+agentResp.Msg)
		return
	}

	middleware.ResponseSuccess(w, agentResp.Msg, agentResp.Data)
}
