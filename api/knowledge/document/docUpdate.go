package document

import (
	"cmdb/models/knowledge"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// UpdateRequest 更新文档请求
type UpdateRequest struct {
	Id       uint   `json:"id" validate:"required"`
	Title    string `json:"title" validate:"required,min=2,max=200"`
	Content  string `json:"content" validate:"required"`
	Project  string `json:"project" validate:"required,min=2,max=200"`
	Category string `json:"category" validate:"required"`
}

// DocUpdate 更新文档
func DocUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 解析请求体
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查文档是否存在
	var doc knowledge.Document
	if err := database.DB.Where("id = ?", req.Id).First(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "文档不存在")
		return
	}

	// 验证项目权限
	auth := utils.ProjectAuth(
		uint(userInfo.UserID),
		uint(userInfo.DeptID),
		uint(userInfo.RoleID),
		role.Level,
		userInfo.OnlineAssets,
		userInfo.TestAssets,
		r.URL.Path,
	)

	// 验证当前项目权限
	if hasPermission, errMsg := auth.CheckProjectPermission(doc.Project); !hasPermission {
		middleware.ResponseError(w, http.StatusForbidden, errMsg)
		return
	}

	// 如果要更改项目，还需要验证目标项目的权限
	if req.Project != doc.Project {
		if hasPermission, _ := auth.CheckProjectPermission(req.Project); !hasPermission {
			middleware.ResponseError(w, http.StatusForbidden, "无权将文档移动到目标项目")
			return
		}
	}

	// 保存历史版本
	history := knowledge.DocumentHistory{
		DocID:    doc.ID,
		Title:    doc.Title,
		Content:  doc.Content,
		Category: doc.Category,
		Project:  doc.Project,
		Version:  doc.Version,
		Creator:  doc.Creator,
		UserName: userInfo.UserName,
	}
	if err := database.DB.Create(&history).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "保存历史版本失败")
		return
	}

	// 更新文档
	doc.Title = req.Title
	doc.Content = req.Content
	doc.Category = req.Category
	doc.Project = req.Project
	doc.UserName = userInfo.UserName // 更新人
	doc.Version++                    // 版本号加1

	if err := database.DB.Save(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新文档失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", doc)
}
