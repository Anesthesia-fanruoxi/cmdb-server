package project

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"time"
)

// ProjectCreateRequest 创建项目请求
type ProjectCreateRequest struct {
	Project          string `json:"project" validate:"required"`      // 必填：项目简称
	ProjectName      string `json:"project_name" validate:"required"` // 必填：项目名称
	AgentURL         string `json:"agent_url"`                        // 可选
	BackenDomain     string `json:"backen_domain"`                    // 可选
	ApiDomain        string `json:"api_domain"`                       // 可选
	GitVue           string `json:"git_vue"`                          // 可选
	GitBackend       string `json:"git_backend"`                      // 可选
	AlterFeishu      string `json:"alter_feishu"`                     // 可选
	UpdateFeishu     string `json:"update_feishu"`                    // 可选
	NotifyFeishu     string `json:"notify_feishu"`                    // 可选
	Description      string `json:"description"`                      // 可选
	BackendTool      string `json:"backend_tool"`                     // 可选
	FrontendTool     string `json:"frontend_tool"`                    // 可选
	EnableSkywalking bool   `json:"enable_skywalking"`                // 可选
	Color            string `json:"color"`                            // 可选
	Logo             string `json:"logo"`                             // 可选：项目Logo地址
}

// ProjectCreate 创建项目
// POST /api/system/project/create
func ProjectCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 解析请求
	var req ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数解析失败: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目简称不能为空")
		return
	}
	if req.ProjectName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目名称不能为空")
		return
	}

	// 检查项目是否已存在
	var count int64
	if err := database.DB.Model(&system.ProjectDict{}).Where("project = ?", req.Project).Count(&count).Error; err != nil {
		middleware.Logger.Errorf("检查项目是否存在失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "检查项目失败")
		return
	}
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "项目简称已存在")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 创建项目记录
	project := system.ProjectDict{
		Project:          req.Project,
		ProjectName:      req.ProjectName,
		AgentURL:         req.AgentURL,
		BackenDomain:     req.BackenDomain,
		ApiDomain:        req.ApiDomain,
		GitVue:           req.GitVue,
		GitBackend:       req.GitBackend,
		AlterFeishu:      req.AlterFeishu,
		UpdateFeishu:     req.UpdateFeishu,
		NotifyFeishu:     req.NotifyFeishu,
		Description:      req.Description,
		BackendTool:      req.BackendTool,
		FrontendTool:     req.FrontendTool,
		EnableSkywalking: req.EnableSkywalking,
		Color:            req.Color,
		Logo:             req.Logo,
		CreatedBy:        int64(userInfo.UserID),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := database.DB.Create(&project).Error; err != nil {
		middleware.Logger.Errorf("创建项目失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "创建项目失败")
		return
	}

	middleware.ResponseSuccess(w, "创建项目成功", map[string]interface{}{
		"id":      project.ID,
		"project": project.Project,
	})
}
