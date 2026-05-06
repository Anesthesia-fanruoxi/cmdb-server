package security

import (
	"cmdb/pkg/middleware"
	"cmdb/pkg/security"
	"encoding/json"
	"io"
	"net/http"
)

// GetConfig 获取安全配置
func SecurityGetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := security.GetConfig(r.Context())
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取配置失败: "+err.Error())
		return
	}
	middleware.ResponseSuccess(w, "获取成功", config)
}

// UpdateConfig 更新安全配置
func SecurityUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败: "+err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// 解析JSON
	var config security.SecurityConfig
	if err := json.Unmarshal(body, &config); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if config.BlacklistCount < 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "黑名单错误次数必须大于0")
		return
	}
	if config.BlacklistLock < 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "黑名单锁定时间必须大于0")
		return
	}
	if config.RateLimitCD < 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "限流冷却时间必须大于0")
		return
	}
	if config.RateLimitTTL < 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "Redis存储时间必须大于0")
		return
	}

	if err := security.UpdateConfig(r.Context(), &config); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新配置失败: "+err.Error())
		return
	}
	middleware.ResponseSuccess(w, "更新成功", nil)
}
