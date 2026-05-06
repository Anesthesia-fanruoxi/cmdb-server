package project

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Total int64              `json:"total"`
	Items []ProjectBriefInfo `json:"items"`
}

// ProjectBriefInfo 项目简要信息
type ProjectBriefInfo struct {
	ID               int64  `json:"id"`
	Project          string `json:"project"`
	ProjectName      string `json:"project_name"`
	Description      string `json:"description"`
	AgentURL         string `json:"agent_url"`
	AlterFeishu      string `json:"alter_feishu"`
	UpdateFeishu     string `json:"update_feishu"`
	NotifyFeishu     string `json:"notify_feishu"`
	BackendTool      string `json:"backend_tool"`
	FrontendTool     string `json:"frontend_tool"`
	EnableSkywalking bool   `json:"enable_skywalking"`
	Logo             string `json:"logo"`
	CreatedAt        string `json:"created_at"`
}

// ProjectList 获取项目列表（简要信息）
// GET /api/system/project/list
func ProjectList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取项目权限
	accessibleProjects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 查询项目
	var projects []system.ProjectDict
	query := database.DB.Order("created_at DESC")

	// 如果不是超级管理员，只查询有权限的项目
	if !isSuperAdmin {
		if len(accessibleProjects) == 0 {
			// 没有可访问的项目，返回空列表
			middleware.ResponseSuccess(w, "获取项目列表成功", ProjectListResponse{Total: 0, Items: []ProjectBriefInfo{}})
			return
		}
		query = query.Where("project IN ?", accessibleProjects)
	}

	if err := query.Find(&projects).Error; err != nil {
		middleware.Logger.Errorf("查询项目列表失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目列表失败")
		return
	}

	// 构建简要信息列表
	items := make([]ProjectBriefInfo, 0, len(projects))
	for _, p := range projects {
		items = append(items, ProjectBriefInfo{
			ID:               p.ID,
			Project:          p.Project,
			ProjectName:      p.ProjectName,
			Description:      p.Description,
			AgentURL:         p.AgentURL,
			AlterFeishu:      p.AlterFeishu,
			UpdateFeishu:     p.UpdateFeishu,
			NotifyFeishu:     p.NotifyFeishu,
			BackendTool:      p.BackendTool,
			FrontendTool:     p.FrontendTool,
			EnableSkywalking: p.EnableSkywalking,
			Logo:             p.Logo,
			CreatedAt:        middleware.FormatTimeToChina(p.CreatedAt),
		})
	}

	// 构造响应
	response := ProjectListResponse{
		Total: int64(len(items)),
		Items: items,
	}

	middleware.ResponseSuccess(w, "获取项目列表成功", response)
}
