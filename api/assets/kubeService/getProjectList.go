package kubeService

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// ProjectItem 项目列表项
type ProjectItem struct {
	Value string `json:"value"` // 项目标识
	Label string `json:"label"` // 项目显示名称
}

// GetProjectList 获取K8s项目列表接口
// 支持权限验证和测试环境权限忽略
func GetProjectList(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取项目列表并验证权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "test")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 添加调试日志
	userInfo, _ := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	middleware.Logger.Infof("K8s项目列表权限调试 - 用户: %s, TestAssets: %v, OnlineAssets: %v, 超级管理员: %v, 可访问项目数: %d",
		userInfo.UserName, userInfo.TestAssets, userInfo.OnlineAssets, isSuperAdmin, len(projects))

	// 如果不是超级管理员且没有可访问的项目，返回空列表
	if !isSuperAdmin && len(projects) == 0 {
		middleware.ResponseSuccess(w, "查询成功", []ProjectItem{})
		return
	}

	// 构建项目列表数据
	var projectItems []ProjectItem

	if isSuperAdmin || len(projects) > 0 {
		// 查询项目字典数据，获取显示名称
		var dictItems []system.ProjectDict
		query := database.DB.Table(middleware.TableProjectDict)

		// 如果不是超级管理员，只查询有权限的项目
		if !isSuperAdmin {
			query = query.Where("project IN ?", projects)
		}

		if err := query.Find(&dictItems).Error; err != nil {
			middleware.Logger.Errorf("查询项目字典失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "查询项目列表失败")
			return
		}

		// 构建响应数据
		for _, item := range dictItems {
			// 屏蔽ops项目（运维文档）
			if item.Project == "ops" {
				continue
			}

			projectItems = append(projectItems, ProjectItem{
				Value: item.Project,
				Label: item.ProjectName,
			})
		}
	}

	// 返回结果
	middleware.Logger.Infof("返回项目列表，共 %d 个项目", len(projectItems))
	middleware.ResponseSuccess(w, "查询成功", projectItems)
}
