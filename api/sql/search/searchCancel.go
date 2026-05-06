package search

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// CancelRequest 取消查询请求
type CancelRequest struct {
	QueryID string `json:"query_id" validate:"required"`
	Agent   string `json:"agent"` // Agent名称
}

// 取消标记过期时间
const CancelTTL = 5 * time.Minute

// SqlSearchCancel 取消SQL查询
func SqlSearchCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	var req CancelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	if req.QueryID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "query_id不能为空")
		return
	}

	if database.RDB == nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "Redis客户端未初始化")
		return
	}

	ctx := context.Background()
	cancelKey := middleware.GetSqlSearchCancelKey(req.QueryID)

	// 设置取消标记
	if err := database.RDB.Set(ctx, cancelKey, "1", CancelTTL).Err(); err != nil {
		middleware.Logger.Errorf("[查询取消] 设置取消标记失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "设置取消标记失败")
		return
	}

	middleware.Logger.Infof("[查询取消] 查询 %s 已标记为取消", req.QueryID)

	// 如果提供了 Agent 信息，通知插件取消查询
	if req.Agent != "" {
		go notifyPluginCancel(req.Agent, req.QueryID)
	}

	middleware.ResponseSuccess(w, "查询已取消", nil)
}

// notifyPluginCancel 通知插件取消查询
func notifyPluginCancel(agent, queryID string) {
	// 构建取消请求
	cancelData := map[string]string{
		"query_id": queryID,
	}

	pluginPath := "/proxy/sql-plugs/api/sql/cancel"
	_, err := middleware.CallAgentWithBody(agent, pluginPath, cancelData)
	if err != nil {
		middleware.Logger.Errorf("[查询取消] 通知插件取消失败: agent=%s, query_id=%s, error=%v", agent, queryID, err)
	} else {
		middleware.Logger.Infof("[查询取消] 已通知插件取消: agent=%s, query_id=%s", agent, queryID)
	}
}

// IsQueryCancelled 检查查询是否已被取消
func IsQueryCancelled(ctx context.Context, queryID string) bool {
	if database.RDB == nil {
		return false
	}

	cancelKey := middleware.GetSqlSearchCancelKey(queryID)
	exists, err := database.RDB.Exists(ctx, cancelKey).Result()
	if err != nil {
		return false
	}

	return exists > 0
}
