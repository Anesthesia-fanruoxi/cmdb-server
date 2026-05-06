package project

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ProjectUpdateRequest 更新项目请求（所有字段可选，传什么更新什么）
type ProjectUpdateRequest struct {
	Project          string  `json:"project" validate:"required"` // 必填：项目标识（用于定位）
	ProjectName      *string `json:"project_name"`                // 可选
	AgentURL         *string `json:"agent_url"`                   // 可选
	BackenDomain     *string `json:"backen_domain"`               // 可选
	ApiDomain        *string `json:"api_domain"`                  // 可选
	GitVue           *string `json:"git_vue"`                     // 可选
	GitBackend       *string `json:"git_backend"`                 // 可选
	AlterFeishu      *string `json:"alter_feishu"`                // 可选
	UpdateFeishu     *string `json:"update_feishu"`               // 可选
	NotifyFeishu     *string `json:"notify_feishu"`               // 可选
	Description      *string `json:"description"`                 // 可选
	BackendTool      *string `json:"backend_tool"`                // 可选
	FrontendTool     *string `json:"frontend_tool"`               // 可选
	EnableSkywalking *bool   `json:"enable_skywalking"`           // 可选
	Color            *string `json:"color"`                       // 可选
	Logo             *string `json:"logo"`                        // 可选：项目Logo地址
}

// ProjectUpdate 更新项目（部分更新）
// PUT /api/system/project/update
func ProjectUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持PUT请求")
		return
	}

	// 解析请求
	var req ProjectUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数解析失败: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 检查项目是否存在
	var project system.ProjectDict
	if err := database.DB.Where("project = ?", req.Project).First(&project).Error; err != nil {
		middleware.Logger.Errorf("查询项目失败: %v", err)
		middleware.ResponseError(w, http.StatusNotFound, "项目不存在")
		return
	}

	// 构建更新映射（只更新传递的字段）
	updates := make(map[string]interface{})

	if req.ProjectName != nil {
		updates["project_name"] = *req.ProjectName
	}
	if req.AgentURL != nil {
		updates["agent_url"] = *req.AgentURL
	}
	if req.BackenDomain != nil {
		updates["backen_domain"] = *req.BackenDomain
	}
	if req.ApiDomain != nil {
		updates["api_domain"] = *req.ApiDomain
	}
	if req.GitVue != nil {
		updates["git_vue"] = *req.GitVue
	}
	if req.GitBackend != nil {
		updates["git_backend"] = *req.GitBackend
	}
	if req.AlterFeishu != nil {
		updates["alter_feishu"] = *req.AlterFeishu
	}
	if req.UpdateFeishu != nil {
		updates["update_feishu"] = *req.UpdateFeishu
	}
	if req.NotifyFeishu != nil {
		updates["notify_feishu"] = *req.NotifyFeishu
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.BackendTool != nil {
		updates["backend_tool"] = *req.BackendTool
	}
	if req.FrontendTool != nil {
		updates["frontend_tool"] = *req.FrontendTool
	}
	if req.EnableSkywalking != nil {
		updates["enable_skywalking"] = *req.EnableSkywalking
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Logo != nil {
		updates["logo"] = *req.Logo
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 执行更新
	if err := database.DB.Model(&project).Updates(updates).Error; err != nil {
		middleware.Logger.Errorf("更新项目失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "更新项目失败")
		return
	}

	// 重新查询项目完整信息用于通知CICD
	var updatedProject system.ProjectDict
	if err := database.DB.Where("project = ?", req.Project).First(&updatedProject).Error; err == nil {
		// 异步通知CICD服务（失败不影响主流程）
		go notifyCICDConfigUpdate(updatedProject)
	}

	middleware.ResponseSuccess(w, "更新项目成功", map[string]interface{}{
		"project": req.Project,
		"updated": len(updates),
	})
}

// notifyCICDConfigUpdate 异步通知CICD服务项目配置更新
func notifyCICDConfigUpdate(project system.ProjectDict) {
	if err := sendProjectConfigToCICD(project); err != nil {
		middleware.Logger.Errorf("通知CICD服务失败 - Project: %s, Error: %v", project.Project, err)
	} else {
		middleware.Logger.Infof("成功通知CICD服务 - Project: %s", project.Project)
	}
}

// sendProjectConfigToCICD 发送项目配置到CICD服务
func sendProjectConfigToCICD(project system.ProjectDict) error {
	// 获取配置
	cfg := config.GetConfig()

	// 检查CICD配置
	if cfg.CICD.URL == "" {
		return fmt.Errorf("CICD服务地址未配置")
	}
	if cfg.CICD.Secret == "" {
		return fmt.Errorf("CICD服务密钥未配置")
	}

	// 构建配置数据
	configData := map[string]interface{}{
		"project":           project.Project,
		"project_name":      project.ProjectName,
		"agent_url":         project.AgentURL,
		"backen_domain":     project.BackenDomain,
		"api_domain":        project.ApiDomain,
		"git_vue":           project.GitVue,
		"git_backend":       project.GitBackend,
		"alter_feishu":      project.AlterFeishu,
		"update_feishu":     project.UpdateFeishu,
		"notify_feishu":     project.NotifyFeishu,
		"description":       project.Description,
		"backend_tool":      project.BackendTool,
		"frontend_tool":     project.FrontendTool,
		"enable_skywalking": project.EnableSkywalking,
		"updated_at":        project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 序列化为JSON
	plainData, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("序列化项目配置失败: %v", err)
	}
	middleware.Logger.Infof("项目配置数据 - Project: %s, Data: %s", project.Project, string(plainData))

	middleware.Logger.Infof("准备通知CICD - Project: %s", project.Project)

	// 加密数据
	encryptedData, err := middleware.CompressAndEncrypt(plainData, cfg.CICD.Secret)
	if err != nil {
		return fmt.Errorf("加密项目配置失败: %v", err)
	}

	// 构建请求体
	requestBody := map[string]string{
		"data": encryptedData,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 构建完整URL
	updateURL := fmt.Sprintf("%s/api/config/update", cfg.CICD.URL)
	middleware.Logger.Infof("发送项目配置到CICD - URL: %s, Project: %s", updateURL, project.Project)

	// 创建HTTP客户端（设置超时避免阻塞）
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送POST请求
	resp, err := client.Post(updateURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求到CICD服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CICD服务返回错误状态: %d", resp.StatusCode)
	}

	// 解析响应
	var cicdResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cicdResp); err != nil {
		return fmt.Errorf("解析CICD响应失败: %v", err)
	}

	if cicdResp.Code != 200 {
		return fmt.Errorf("CICD服务返回错误: %s", cicdResp.Message)
	}

	return nil
}
