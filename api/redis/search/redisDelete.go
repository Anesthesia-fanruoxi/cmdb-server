package search

import (
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// RedisDelete 删除Redis key
// @Summary 删除Redis key
// @Description 根据project和key删除指定Redis key
// @Tags 资产管理
// @Accept json
// @Produce json
// @Param project query string true "项目标识"
// @Param key query string true "Redis key名称"
// @Success 200 {object} map[string]interface{}
// @Router /api/redis/delete [delete]
func RedisDelete(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	key := r.URL.Query().Get("key")

	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if key == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "key参数不能为空")
		return
	}

	projectInfo, err := system.GetProjectInfo(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目信息失败: %v", err))
		return
	}

	agentURL := projectInfo.AgentURL
	if agentURL == "" {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("项目 %s 未配置agent地址", project))
		return
	}

	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret

	reqURL := fmt.Sprintf("%s/proxy/redis-plugs/api/delete?key=%s", agentURL, url.QueryEscape(key))

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建请求失败")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("请求Agent失败: %v", err))
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取Agent响应失败")
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	if encryptedRespData, ok := result["data"].(string); ok && encryptedRespData != "" {
		decompressedData, err := middleware.DecryptAndDecompress(encryptedRespData, taskSalt)
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "解密解压失败: "+err.Error())
			return
		}

		var deleteData map[string]interface{}
		if err := json.Unmarshal(decompressedData, &deleteData); err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "解析Redis delete数据失败")
			return
		}

		if innerData, ok := deleteData["data"]; ok {
			middleware.ResponseSuccess(w, "删除Redis key成功", innerData)
		} else {
			middleware.ResponseSuccess(w, "删除Redis key成功", deleteData)
		}
	} else {
		errorMsg := extractErrorMessage(result, resp.StatusCode)
		middleware.ResponseError(w, resp.StatusCode, errorMsg)
	}
}
