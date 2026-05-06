package alBalance

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

const alBalanceConfigKey = "cmdb:alBalance:projects"

// ConfigRequest 配置请求体
type ConfigRequest struct {
	Projects []string `json:"projects"`
}

// UpdateConfig 保存监控的项目列表到 Redis（永久存储）
// POST /api/assets/alBalance/config/update
func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	data, err := json.Marshal(req.Projects)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化失败")
		return
	}

	if err := database.RDB.Set(r.Context(), alBalanceConfigKey, string(data), 0).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "保存配置失败: "+err.Error())
		return
	}

	middleware.ResponseSuccess(w, "配置保存成功", req.Projects)
}

// GetConfig 获取已配置的项目列表
// GET /api/assets/alBalance/config/list
func GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	val, err := database.RDB.Get(r.Context(), alBalanceConfigKey).Result()
	if err == database.RedisNil {
		middleware.ResponseSuccess(w, "success", []string{})
		return
	}
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取配置失败: "+err.Error())
		return
	}

	var projects []string
	if err := json.Unmarshal([]byte(val), &projects); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析配置失败")
		return
	}

	middleware.ResponseSuccess(w, "success", projects)
}
