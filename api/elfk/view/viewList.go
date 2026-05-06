package view

import (
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

func ListView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 获取项目参数
	project := r.URL.Query().Get("project")

	// 创建基础查询
	query := database.DB.Model(&elfk.View{})

	// 当没有project参数时
	if project == "" {
		if len(projects) > 0 {
			// 直接使用字符串数组作为IN条件
			query = query.Where("project IN (?)", projects)
		} else {
			middleware.ResponseSuccess(w, "获取视图列表成功", []elfk.ViewResponse{})
			return
		}
	} else {
		// 验证指定项目权限
		if !isSuperAdmin && !utils.Contains(projects, project) {
			middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
			return
		}
		query = query.Where("project = ?", project)
	}

	// 执行查询
	var views []elfk.View
	if err := query.Find(&views).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取视图列表失败")
		return
	}

	// 转换为响应结构
	var response []elfk.ViewResponse
	for _, v := range views {
		response = append(response, elfk.ViewResponse{
			ID:           v.ID,
			Name:         v.Name,
			Project:      v.Project,
			Category:     v.Category,
			IndexPattern: v.IndexPattern,
			TimeField:    v.TimeField,
			TimeFormat:   v.TimeFormat,
			Description:  v.Description,
			UpdateTime:   v.UpdateTime,
		})
	}

	middleware.ResponseSuccess(w, "获取视图列表成功", response)
}
