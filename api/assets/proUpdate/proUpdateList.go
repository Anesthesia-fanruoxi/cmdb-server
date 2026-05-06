package proUpdate

import (
	"cmdb/models/assets"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// ProUpdateList 获取项目发版列表
// GET /api/assets/proUpdate/list
func ProUpdateList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "未登录")
		return
	}

	// 检查接口权限（防止绕过前端）
	hasPermission := false
	for _, perm := range userInfo.Permissions {
		if perm == "assets:proUpdate:r" || perm == "assets:proUpdate:w" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		middleware.ResponseError(w, http.StatusForbidden, "无接口访问权限")
		return
	}

	// 判断是否有前后端操作权限（角色ID为1、2、3时有完整权限）
	allowedRoleIDs := []int64{1, 2, 3}
	hasFullAccess := utils.ContainsInt64(allowedRoleIDs, userInfo.RoleID)

	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 查询项目字典
	var projectDicts []system.ProjectDict
	query := database.DB

	// 如果不是超级管理员，只查询有权限的项目
	if !isSuperAdmin {
		if len(projects) == 0 {
			// 没有可访问的项目，返回空列表
			response := assets.ProjectUpdateListResponse{
				Total: 0,
				Items: []assets.ProjectUpdateInfo{},
			}
			middleware.ResponseSuccess(w, "获取项目发版列表成功", response)
			return
		}
		query = query.Where("project IN ?", projects)
	}

	if err := query.Find(&projectDicts).Error; err != nil {
		middleware.Logger.Errorf("查询项目字典失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询项目字典失败")
		return
	}

	// 查询所有发版记录
	var allRecords []assets.ReleaseRecord
	if err := database.DB.Find(&allRecords).Error; err != nil {
		middleware.Logger.Errorf("查询发版记录失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "查询发版记录失败")
		return
	}

	// 获取类型参数（web=前端，空=后端）
	requestType := r.URL.Query().Get("type")

	// 如果用户没有完整权限（角色ID不是1、2、3），强制只返回前端数据
	if !hasFullAccess {
		requestType = "web"
	}

	// 构建发版记录映射，按项目名称索引（根据 type 筛选）
	recordMap := make(map[string][]assets.ReleaseRecord)
	for _, record := range allRecords {
		if requestType == "web" {
			// 前端：只统计 type=web 的记录
			if record.Type == "web" {
				recordMap[record.Project] = append(recordMap[record.Project], record)
			}
		} else {
			// 后端：统计 type != web 的记录
			if record.Type != "web" {
				recordMap[record.Project] = append(recordMap[record.Project], record)
			}
		}
	}

	// 准备结果列表
	items := make([]assets.ProjectUpdateInfo, 0)

	// 遍历所有项目字典，为每个项目统计发版信息
	for _, project := range projectDicts {
		// 屏蔽ops项目
		if project.Project == "ops" {
			continue
		}

		// 根据请求类型确定Git地址
		var gitURL string
		if requestType == "web" {
			gitURL = project.GitVue // web类型返回前端Git地址
		} else {
			gitURL = project.GitBackend // 后端类型返回后端Git地址
		}

		// 创建项目信息
		item := assets.ProjectUpdateInfo{
			Project:          project.Project,
			ProjectName:      project.ProjectName,
			GitURL:           gitURL,
			FeishuURL:        project.UpdateFeishu,
			BackendTool:      project.BackendTool,
			FrontendTool:     project.FrontendTool,
			EnableSkywalking: project.EnableSkywalking,
			LastUpdate:       "-",
			Status:           "-",
			TotalUpdates:     0,
		}

		// 统计该项目的发版记录（直接用项目名查询，不再拼接-web）
		if records, exists := recordMap[project.Project]; exists {
			item.TotalUpdates = len(records)

			// 找到最新的发版记录（按开始时间）
			var latestRecord *assets.ReleaseRecord
			for i := range records {
				if !records[i].StartedAt.IsZero() {
					if latestRecord == nil || records[i].StartedAt.After(latestRecord.StartedAt) {
						latestRecord = &records[i]
					}
				}
			}

			// 设置最新记录信息
			if latestRecord != nil {
				item.LastUpdate = middleware.FormatTimeToChina(latestRecord.StartedAt)
				item.Status = latestRecord.Status
			}
		}

		items = append(items, item)
	}

	// 构造响应
	response := assets.ProjectUpdateListResponse{
		Total: len(items),
		Items: items,
	}

	// 返回成功响应
	middleware.ResponseSuccess(w, "获取项目发版列表成功", response)
}
