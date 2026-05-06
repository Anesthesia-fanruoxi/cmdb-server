package proUpdate

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/assets"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ProUpdateCancelRequest 取消任务请求结构
type ProUpdateCancelRequest struct {
	ID string `json:"id" validate:"required"` // 任务ID
}

// ProUpdateCancel 取消项目发版任务
// POST /api/assets/proUpdate/cancel
func ProUpdateCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求参数
	var req ProUpdateCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.Logger.Errorf("解析取消任务请求失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	// 验证参数
	if req.ID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 从 Redis 中查询任务信息（获取 remote 和 project）
	ctx := context.Background()
	taskInfo, err := getTaskInfoFromRedis(ctx, req.ID)
	if err != nil {
		middleware.Logger.Errorf("从Redis获取任务信息失败 - TaskID: %s, Error: %v", req.ID, err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取任务信息失败: %v", err))
		return
	}

	middleware.Logger.Infof("从Redis获取任务信息 - TaskID: %s, Remote: %s, Project: %s", req.ID, taskInfo.Remote, taskInfo.Project)

	middleware.Logger.Infof("收到取消任务请求 - TaskID: %s, Remote: %s", req.ID, taskInfo.Remote)

	// 调用代理服务取消任务
	if err := cancelTaskOnProxy(req.ID, taskInfo); err != nil {
		middleware.Logger.Errorf("调用代理服务取消任务失败 - TaskID: %s, Error: %v", req.ID, err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("取消任务失败: %v", err))
		return
	}

	middleware.Logger.Infof("任务取消成功 - TaskID: %s", req.ID)
	middleware.ResponseSuccess(w, "任务取消请求已发送", map[string]interface{}{
		"task_id": req.ID,
	})
}

// cancelTaskOnProxy 调用代理服务取消任务
func cancelTaskOnProxy(taskID string, taskInfo *TaskInfoFromRedis) error {
	// 从配置文件读取服务地址
	cfg := config.GetConfig()

	// 根据 remote 选择目标地址和加密密钥
	var baseURL string
	var cancelURL string
	var routeBy string
	var salt string

	if taskInfo.Remote == "server" {
		// 直接走 cicd-service
		baseURL = cfg.CICD.URL
		cancelURL = fmt.Sprintf("%s/api/task/cancel", baseURL)
		salt = cfg.CICD.Secret
		routeBy = "server"
		middleware.Logger.Infof("取消任务路由到CICD服务 - URL: %s", cancelURL)
	} else {
		// agent: 使用已从 Redis 获取的项目信息
		if taskInfo.Project == "" {
			return fmt.Errorf("缺少项目信息")
		}

		// 提取项目标识（处理 -web 后缀）
		projectIdentifier := extractProjectIdentifier(taskInfo.Project)
		middleware.Logger.Infof("取消任务 - 项目名称: %s, 提取标识: %s", taskInfo.Project, projectIdentifier)

		// 从数据库获取项目的 agent_url
		var err error
		baseURL, err = system.GetProjectAgentURL(projectIdentifier)
		if err != nil {
			return fmt.Errorf("获取项目 %s 的Agent地址失败: %v", projectIdentifier, err)
		}
		cancelURL = fmt.Sprintf("%s/proxy/cicd-agent/api/task/cancel", baseURL)
		salt = cfg.Agent.Secret
		routeBy = fmt.Sprintf("agent(project=%s)", projectIdentifier)
		middleware.Logger.Infof("项目 %s 路由到cmdb-agent - URL: %s", projectIdentifier, cancelURL)
	}

	// 构建明文请求体
	requestBody := map[string]string{
		"id": taskID,
	}

	// 序列化明文数据
	plainData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 检查加密密钥
	if salt == "" {
		return fmt.Errorf("系统配置错误：缺少加密密钥 (remote=%s)", taskInfo.Remote)
	}

	// 加密数据
	encryptedData, err := middleware.CompressAndEncrypt(plainData, salt)
	if err != nil {
		return fmt.Errorf("加密请求数据失败: %v", err)
	}

	// 构建发送给CICD代理的加密请求体
	finalRequestBody := map[string]string{
		"data": encryptedData,
	}

	jsonData, err := json.Marshal(finalRequestBody)
	if err != nil {
		return fmt.Errorf("序列化加密请求数据失败: %v", err)
	}

	middleware.Logger.Infof("调用代理服务取消任务 - RouteBy: %s, URL: %s, TaskID: %s, 使用密钥类型: %s",
		routeBy, cancelURL, taskID, taskInfo.Remote)

	// 发送HTTP POST请求
	resp, err := http.Post(cancelURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送取消请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}
	middleware.Logger.Infof("相应为: %s", string(respBody))
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("代理服务返回错误状态: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 统一处理：所有响应都是加密的
	var encryptedResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"` // 加密字符串
	}

	if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
		return fmt.Errorf("解析响应失败: %v, 响应: %s", err, string(respBody))
	}

	// 解密响应数据（使用与请求相同的密钥）
	decryptedData, err := middleware.DecryptAndDecompress(encryptedResp.Data, salt)
	if err != nil {
		return fmt.Errorf("解密响应数据失败: %v (HTTP状态: %d, remote: %s)", err, resp.StatusCode, taskInfo.Remote)
	}

	middleware.Logger.Infof("取消任务解密成功 - TaskID: %s, Remote: %s, 解密后响应: %s",
		taskID, taskInfo.Remote, string(decryptedData))

	// 检查解密后的响应是否成功
	var finalResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(decryptedData, &finalResp); err == nil {
		if finalResp.Code != 200 {
			return fmt.Errorf("取消任务失败: %s", finalResp.Message)
		}
	}

	return nil
}

// TaskInfoFromRedis 从Redis获取的任务信息
type TaskInfoFromRedis struct {
	Remote  string // 任务的remote值
	Project string // 任务的项目名称
}

// getTaskInfoFromRedis 从Redis中获取任务信息（remote和project）
func getTaskInfoFromRedis(ctx context.Context, taskID string) (*TaskInfoFromRedis, error) {
	// 获取任务记录
	recordJSON, err := GetTaskRecord(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务记录失败: %v", err)
	}

	// 解析任务记录
	var taskRecords []assets.RedisTaskRecord
	if err := json.Unmarshal([]byte(recordJSON), &taskRecords); err != nil {
		return nil, fmt.Errorf("解析任务记录失败: %v", err)
	}

	if len(taskRecords) == 0 {
		return nil, fmt.Errorf("未找到任务记录")
	}

	taskRecord := taskRecords[0]

	// 提取项目名称
	project := taskRecord.ProjectName
	if project == "" {
		return nil, fmt.Errorf("任务记录中没有项目信息")
	}

	// 查找最新步骤的remote值（取最后一个有remote值的步骤）
	remote := "agent" // 默认值
	for i := len(taskRecord.StepDetail) - 1; i >= 0; i-- {
		step := taskRecord.StepDetail[i]
		if step.Remote != "" {
			remote = step.Remote
			break
		}
	}

	return &TaskInfoFromRedis{
		Remote:  remote,
		Project: project,
	}, nil
}
