package personal

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// PersonalListResponse 个人文档列表响应
type PersonalListResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PersonalList 获取个人文档列表
func PersonalList(w http.ResponseWriter, r *http.Request) {
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

	// 构建查询
	query := database.DB.Model(&knowledge.Personal{})

	// 获取列表数据
	var docs []knowledge.Personal
	if err := query.Select("id, title, user_name, created_at, updated_at").
		Where("user_name = ?", userInfo.UserName).
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取个人文档列表失败")
		return
	}

	// 构建响应数据
	list := make([]PersonalListResponse, 0, len(docs))
	for _, doc := range docs {
		list = append(list, PersonalListResponse{
			ID:        doc.ID,
			Title:     doc.Title,
			UserName:  doc.UserName,
			CreatedAt: doc.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.ResponseSuccess(w, "获取成功", list)
}
