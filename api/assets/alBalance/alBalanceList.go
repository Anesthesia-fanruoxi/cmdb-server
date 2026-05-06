package alBalance

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"sync"
)

// BalanceInfo 余额信息
type BalanceInfo struct {
	Currency            string `json:"currency"`
	AvailableAmount     string `json:"available_amount"`
	AvailableCashAmount string `json:"available_cash_amount"`
	CreditAmount        string `json:"credit_amount"`
}

// ProjectBalanceResult 单个项目的余额查询结果
type ProjectBalanceResult struct {
	Project string       `json:"project"`
	Data    *BalanceInfo `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GetBalance 从 Redis 读取配置的项目列表，并发查询余额后一次性返回
// GET /api/assets/alBalance/balance
func GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 从 Redis 读取已配置的项目列表
	val, err := database.RDB.Get(r.Context(), alBalanceConfigKey).Result()
	if err == database.RedisNil {
		middleware.ResponseSuccess(w, "success", []ProjectBalanceResult{})
		return
	}
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取配置失败: "+err.Error())
		return
	}

	var configuredProjects []string
	if err := json.Unmarshal([]byte(val), &configuredProjects); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析配置失败")
		return
	}
	if len(configuredProjects) == 0 {
		middleware.ResponseSuccess(w, "success", []ProjectBalanceResult{})
		return
	}

	// 获取当前用户可访问的项目列表，取交集
	accessibleProjects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var projects []string
	for _, proj := range configuredProjects {
		if isSuperAdmin || utils.Contains(accessibleProjects, proj) {
			projects = append(projects, proj)
		}
	}

	if len(projects) == 0 {
		middleware.ResponseSuccess(w, "success", []ProjectBalanceResult{})
		return
	}

	// 并发查询每个项目的余额
	results := make([]ProjectBalanceResult, len(projects))
	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)
		go func(idx int, proj string) {
			defer wg.Done()

			data, err := middleware.CallAgent(proj, "/proxy/al-plugs/api/account/balance")
			if err != nil {
				middleware.Logger.Errorf("[alBalance] 项目 %s 查询余额失败: %v", proj, err)
				results[idx] = ProjectBalanceResult{Project: proj, Error: err.Error()}
				return
			}

			var agentResp struct {
				Code    int         `json:"code"`
				Message string      `json:"message"`
				Data    BalanceInfo `json:"data"`
			}
			if err := json.Unmarshal(data, &agentResp); err != nil {
				results[idx] = ProjectBalanceResult{Project: proj, Error: "解析响应失败"}
				return
			}
			if agentResp.Code != 200 {
				results[idx] = ProjectBalanceResult{Project: proj, Error: agentResp.Message}
				return
			}

			results[idx] = ProjectBalanceResult{Project: proj, Data: &agentResp.Data}
		}(i, project)
	}

	wg.Wait()
	middleware.ResponseSuccess(w, "success", results)
}
