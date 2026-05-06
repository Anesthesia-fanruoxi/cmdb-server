package project

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// AgentProjectListResponse 项目列表响应
type AgentProjectListResponse struct {
	Project     string `json:"project"`
	ProjectName string `json:"project_name"`
	AgentURL    string `json:"agent_url"`
	Description string `json:"description"`
	HasAgent    bool   `json:"has_agent"`
}

// AgentProjectList 获取已配置Agent的项目列表
func AgentProjectList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取项目权限
	accessibleProjects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 查询项目（只返回配置了agent_url的项目）
	var projects []system.ProjectDict
	query := database.DB.Where("agent_url != ? AND agent_url IS NOT NULL", "").Order("project ASC")

	// 如果不是超级管理员，只查询有权限的项目
	if !isSuperAdmin {
		if len(accessibleProjects) == 0 {
			middleware.ResponseSuccess(w, "查询成功", []AgentProjectListResponse{})
			return
		}
		query = query.Where("project IN ?", accessibleProjects)
	}

	if err := query.Find(&projects).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目列表失败")
		return
	}

	// 转换为响应格式
	var response []AgentProjectListResponse
	for _, p := range projects {
		response = append(response, AgentProjectListResponse{
			Project:     p.Project,
			ProjectName: p.ProjectName,
			AgentURL:    p.AgentURL,
			Description: p.Description,
			HasAgent:    p.AgentURL != "",
		})
	}

	middleware.ResponseSuccess(w, "查询成功", response)
}
