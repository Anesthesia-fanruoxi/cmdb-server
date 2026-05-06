package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// ProjectMenuItem 项目菜单关联项
type ProjectMenuItem struct {
	Project string  `json:"project" validate:"required"`
	MenuIDs []int64 `json:"menu_ids" validate:"required"`
}

// UpdateUserProject 更新用户项目关联请求
type UpdateUserProject struct {
	UserID       int64             `json:"user_id" validate:"required"`
	ProjectMenus []ProjectMenuItem `json:"project_menus" validate:"required"`
}

// UserProjectUpdate 更新用户项目关联
// 用于为用户直接授权项目访问权限，支持跨部门协作
func UserProjectUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req UpdateUserProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查用户是否存在
	var user system.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "用户不存在")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 先删除该用户的所有项目关联（前端传的是完整数据，不在列表中的即为解绑）
	if err := tx.Where("user_id = ?", req.UserID).Delete(&system.UserProject{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除原有项目关联失败")
		return
	}

	// 重新插入所有关联
	middleware.Logger.Infof("[用户项目绑定] 开始保存: userID=%d, projectMenus=%+v", req.UserID, req.ProjectMenus)
	for _, item := range req.ProjectMenus {
		middleware.Logger.Infof("[用户项目绑定] 项目: %s, 菜单IDs: %v", item.Project, item.MenuIDs)
		for _, menuID := range item.MenuIDs {
			userProject := system.UserProject{
				UserID:  req.UserID,
				Project: item.Project,
				MenuID:  menuID,
			}
			middleware.Logger.Infof("[用户项目绑定] 插入记录: userID=%d, project=%s, menuID=%d", req.UserID, item.Project, menuID)
			if err := tx.Create(&userProject).Error; err != nil {
				tx.Rollback()
				middleware.ResponseError(w, http.StatusInternalServerError, "创建项目关联失败")
				return
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新项目关联失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
