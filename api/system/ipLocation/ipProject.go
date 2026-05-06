package ipLocation

import (
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/*
Redis 键值设计：
1. 项目 -> IP：cmdb:ip_project:project:{project} -> String(单个IP)
2. IP -> 项目集合：cmdb:ip_project:ip:{ip} -> Set{project1, project2, ...}
   （因为多个项目可能共享同一个出口IP）
3. 过期时间：永久保存（Agent只在IP变动时才上报）
注：Redis key统一定义在 pkg/middleware/redisKey.go
*/

// RecordIPProject Agent上报IP与项目关联（使用Agent Secret加密）
func RecordIPProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析加密请求
	var encryptedReq system.IPProjectRequest
	if err := json.Unmarshal(body, &encryptedReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求失败")
		return
	}

	if encryptedReq.Data == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少加密数据")
		return
	}

	// 获取Agent加密盐
	cfg := config.GetConfig()
	agentSecret := cfg.Agent.Secret
	if agentSecret == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "Agent加密盐未配置")
		return
	}

	// 解密数据
	decryptedData, err := middleware.DecryptAndDecompress(encryptedReq.Data, agentSecret)
	if err != nil {
		middleware.Logger.Errorf("解密IP项目数据失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "解密数据失败")
		return
	}

	// 解析解密后的数据
	var ipProjectData system.IPProjectData
	if err := json.Unmarshal(decryptedData, &ipProjectData); err != nil {
		middleware.Logger.Errorf("解析IP项目数据失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "数据格式错误")
		return
	}

	// 验证必填字段
	if ipProjectData.IP == "" || ipProjectData.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "IP和Project不能为空")
		return
	}

	// 存储到Redis
	if err := saveIPProjectToRedis(ipProjectData.IP, ipProjectData.Project); err != nil {
		middleware.Logger.Errorf("保存IP项目关联失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "保存失败")
		return
	}

	middleware.Logger.Debugf("记录IP项目关联: IP=%s, Project=%s", ipProjectData.IP, ipProjectData.Project)
	middleware.ResponseSuccess(w, "记录成功", nil)
}

// QueryIPProject 查询IP与项目关联（使用Agent Secret加密）
func QueryIPProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 直接解析查询请求
	var queryReq system.IPProjectQueryRequest
	if err := json.Unmarshal(body, &queryReq); err != nil {
		middleware.Logger.Errorf("解析查询请求失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "数据格式错误")
		return
	}

	// 至少需要一个查询条件
	if queryReq.IP == "" && queryReq.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "至少需要提供IP或Project之一")
		return
	}

	var response system.IPProjectQueryResponse

	// 根据IP查询项目
	if queryReq.IP != "" {
		projects, err := getProjectsByIP(queryReq.IP)
		if err != nil {
			middleware.Logger.Errorf("查询IP关联的项目失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
			return
		}
		response.IP = queryReq.IP
		response.Projects = projects
	}

	// 根据项目查询IP
	if queryReq.Project != "" {
		ip, err := getIPByProject(queryReq.Project)
		if err != nil {
			middleware.Logger.Errorf("查询项目关联的IP失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
			return
		}
		response.Project = queryReq.Project
		response.IP = ip
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}

// saveIPProjectToRedis 保存IP与项目关联到Redis
func saveIPProjectToRedis(ip, project string) error {
	ctx := context.Background()

	// 1. 项目 -> IP（String类型，一对一，永久保存）
	projectKey := middleware.GetProjectIPKey(project)
	if err := database.RDB.Set(ctx, projectKey, ip, 0).Err(); err != nil {
		return fmt.Errorf("设置项目IP关联失败: %v", err)
	}

	// 2. IP -> 项目集合（Set类型，一对多，永久保存）
	ipKey := middleware.GetIPProjectKey(ip)
	if err := database.RDB.SAdd(ctx, ipKey, project).Err(); err != nil {
		return fmt.Errorf("添加IP项目关联失败: %v", err)
	}

	return nil
}

// getProjectsByIP 根据IP查询关联的项目列表
func getProjectsByIP(ip string) ([]string, error) {
	ctx := context.Background()
	key := middleware.GetIPProjectKey(ip)

	projects, err := database.RDB.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return []string{}, nil
	}

	return projects, nil
}

// getIPByProject 根据项目查询关联的IP（单个）
func getIPByProject(project string) (string, error) {
	ctx := context.Background()
	key := middleware.GetProjectIPKey(project)

	ip, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		// 如果key不存在，返回空字符串
		if err.Error() == "redis: nil" {
			return "", nil
		}
		return "", err
	}

	return ip, nil
}
