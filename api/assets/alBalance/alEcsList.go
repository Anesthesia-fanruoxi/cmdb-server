package alBalance

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"sync"
)

// EcsInstance ECS实例信息
type EcsInstance struct {
	InstanceId   string `json:"instance_id"`
	InstanceName string `json:"instance_name"`
	Status       string `json:"status"`
	ExpiredTime  string `json:"expired_time"`
	RegionId     string `json:"region_id"`
	InstanceType string `json:"instance_type"`
	ChargeType   string `json:"charge_type"`
}

// EcsData ECS查询返回数据
type EcsData struct {
	RegionId  string        `json:"region_id"`
	Total     int           `json:"total"`
	Instances []EcsInstance `json:"instances"`
}

// ProjectEcsResult 单个项目的ECS查询结果
type ProjectEcsResult struct {
	Project string   `json:"project"`
	Data    *EcsData `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// GetEcsList 从 Redis 读取配置的项目列表，并发查询ECS到期信息后一次性返回
// GET /api/assets/alBalance/ecs/list
func GetEcsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 从 Redis 读取已配置的项目列表
	val, err := database.RDB.Get(r.Context(), alBalanceConfigKey).Result()
	if err == database.RedisNil {
		middleware.ResponseSuccess(w, "success", []ProjectEcsResult{})
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
		middleware.ResponseSuccess(w, "success", []ProjectEcsResult{})
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
		middleware.ResponseSuccess(w, "success", []ProjectEcsResult{})
		return
	}

	// 并发查询每个项目的ECS到期信息
	results := make([]ProjectEcsResult, len(projects))
	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)
		go func(idx int, proj string) {
			defer wg.Done()

			data, err := middleware.CallAgent(proj, "/proxy/al-plugs/api/resource/ecs/expiry")
			if err != nil {
				middleware.Logger.Errorf("[alEcs] 项目 %s 查询ECS失败: %v", proj, err)
				results[idx] = ProjectEcsResult{Project: proj, Error: err.Error()}
				return
			}

			var agentResp struct {
				Code    int     `json:"code"`
				Message string  `json:"message"`
				Data    EcsData `json:"data"`
			}
			if err := json.Unmarshal(data, &agentResp); err != nil {
				results[idx] = ProjectEcsResult{Project: proj, Error: "解析响应失败"}
				return
			}
			if agentResp.Code != 200 {
				results[idx] = ProjectEcsResult{Project: proj, Error: agentResp.Message}
				return
			}

			results[idx] = ProjectEcsResult{Project: proj, Data: &agentResp.Data}
		}(i, project)
	}

	wg.Wait()
	middleware.ResponseSuccess(w, "success", results)
}
