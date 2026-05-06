package project

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// ProjectDetailInfo 项目详细信息
type ProjectDetailInfo struct {
	ID               int64  `json:"id"`
	Project          string `json:"project"`
	ProjectName      string `json:"project_name"`
	AgentURL         string `json:"agent_url"`
	BackenDomain     string `json:"backen_domain"`
	ApiDomain        string `json:"api_domain"`
	GitVue           string `json:"git_vue"`
	GitBackend       string `json:"git_backend"`
	AlterFeishu      string `json:"alter_feishu"`
	UpdateFeishu     string `json:"update_feishu"`
	NotifyFeishu     string `json:"notify_feishu"`
	Description      string `json:"description"`
	BackendTool      string `json:"backend_tool"`
	FrontendTool     string `json:"frontend_tool"`
	EnableSkywalking bool   `json:"enable_skywalking"`
	Color            string `json:"color"`
	Logo             string `json:"logo"`
	CreatedBy        int64  `json:"created_by"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ProjectDetailList 获取项目详情
// GET /api/system/project/detail?project=xxx
func ProjectDetailList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取项目标识
	projectID := r.URL.Query().Get("project")
	if projectID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 查询项目详情
	var project system.ProjectDict
	if err := database.DB.Where("project = ?", projectID).First(&project).Error; err != nil {
		middleware.Logger.Errorf("查询项目详情失败: %v", err)
		middleware.ResponseError(w, http.StatusNotFound, "项目不存在")
		return
	}

	// 构建详细信息
	detail := ProjectDetailInfo{
		ID:               project.ID,
		Project:          project.Project,
		ProjectName:      project.ProjectName,
		AgentURL:         project.AgentURL,
		BackenDomain:     project.BackenDomain,
		ApiDomain:        project.ApiDomain,
		GitVue:           project.GitVue,
		GitBackend:       project.GitBackend,
		AlterFeishu:      project.AlterFeishu,
		UpdateFeishu:     project.UpdateFeishu,
		NotifyFeishu:     project.NotifyFeishu,
		Description:      project.Description,
		BackendTool:      project.BackendTool,
		FrontendTool:     project.FrontendTool,
		EnableSkywalking: project.EnableSkywalking,
		Color:            project.Color,
		Logo:             project.Logo,
		CreatedBy:        project.CreatedBy,
		CreatedAt:        middleware.FormatTimeToChina(project.CreatedAt),
		UpdatedAt:        middleware.FormatTimeToChina(project.UpdatedAt),
	}

	middleware.ResponseSuccess(w, "获取项目详情成功", detail)
}
