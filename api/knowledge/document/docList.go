package document

import (
	"cmdb/models/knowledge"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// DocListResponse 文档列表响应
type DocListResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Project   string `json:"project"`
	Creator   string `json:"creator"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DocList 获取文档列表
func DocList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	// 构建查询
	query := database.DB.Model(&knowledge.Document{})

	// 如果不是超级管理员，需要根据部门查询可访问的项目
	if role.Level > 0 {
		// 使用 ProjectAuth 获取可访问的项目列表
		auth := utils.ProjectAuth(
			uint(userInfo.UserID),
			uint(userInfo.DeptID),
			uint(userInfo.RoleID),
			role.Level,
			userInfo.OnlineAssets,
			userInfo.TestAssets,
			r.URL.Path,
		)

		projects, err := auth.GetAccessibleProjects("prod")
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询部门项目失败")
			return
		}

		if len(projects) == 0 {
			middleware.ResponseSuccess(w, "获取成功", []DocListResponse{})
			return
		}

		query = query.Where("project IN ?", projects)
	}

	// 获取列表数据
	var docs []knowledge.Document
	if err := query.Select("id, title, category, project, creator, user_name, created_at, updated_at").
		Order("user_name").
		Find(&docs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取文档列表失败")
		return
	}

	// 构建响应数据
	list := make([]DocListResponse, 0, len(docs))
	for _, doc := range docs {
		list = append(list, DocListResponse{
			ID:        doc.ID,
			Title:     doc.Title,
			Category:  doc.Category,
			Project:   doc.Project,
			Creator:   doc.Creator,
			UserName:  doc.UserName,
			CreatedAt: doc.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.ResponseSuccess(w, "获取成功", list)
}
