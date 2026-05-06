package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// RecycleDocListResponse 文档回收站列表响应
type RecycleDocListResponse struct {
	ID          int64  `json:"id"`
	DocID       int64  `json:"doc_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Project     string `json:"project"`
	DeleterName string `json:"deleter_name"`
	DeletedAt   string `json:"deleted_at"`
}

// RecyclePersonalListResponse 个人文档回收站列表响应
type RecyclePersonalListResponse struct {
	ID        int64  `json:"id"`
	DocID     int64  `json:"doc_id"`
	Title     string `json:"title"`
	UserName  string `json:"user_name"`
	DeletedAt string `json:"deleted_at"`
}

// RecyclePublicListResponse 对外文档回收站列表响应
type RecyclePublicListResponse struct {
	ID        int64  `json:"id"`
	DocID     int64  `json:"doc_id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Creator   string `json:"creator"`
	UserName  string `json:"user_name"`
	DeletedAt string `json:"deleted_at"`
}

// RecycleList 获取回收站列表
func RecycleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	docType := r.URL.Query().Get("type")

	// 处理公共文档回收站
	if docType == "doc" {
		// 构建查询
		query := database.DB.Model(&knowledge.DocumentRecycle{})

		// 获取列表数据
		var docs []knowledge.DocumentRecycle
		if err := query.Order("deleted_at DESC").Find(&docs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取回收站列表失败")
			return
		}

		// 构建响应数据
		list := make([]RecycleDocListResponse, 0, len(docs))
		for _, doc := range docs {
			list = append(list, RecycleDocListResponse{
				ID:          doc.ID,
				DocID:       doc.DocID,
				Title:       doc.Title,
				Category:    doc.Category,
				Project:     doc.Project,
				DeleterName: doc.DeleterName,
				DeletedAt:   doc.DeletedAt.Format("2006-01-02 15:04:05"),
			})
		}

		middleware.ResponseSuccess(w, "获取成功", list)
		return
	}

	// 处理个人文档回收站
	if docType == "personal" {
		// 构建查询
		query := database.DB.Model(&knowledge.PersonalRecycle{})

		// 获取列表数据
		var docs []knowledge.PersonalRecycle
		if err := query.Order("deleted_at DESC").Find(&docs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取回收站列表失败")
			return
		}

		// 构建响应数据
		list := make([]RecyclePersonalListResponse, 0, len(docs))
		for _, doc := range docs {
			list = append(list, RecyclePersonalListResponse{
				ID:        doc.ID,
				DocID:     doc.DocID,
				Title:     doc.Title,
				UserName:  doc.UserName,
				DeletedAt: doc.DeletedAt.Format("2006-01-02 15:04:05"),
			})
		}

		middleware.ResponseSuccess(w, "获取成功", list)
		return
	}

	// 处理公共文档回收站
	if docType == "public" {
		// 构建查询
		query := database.DB.Model(&knowledge.PublicRecycle{})

		// 获取列表数据
		var docs []knowledge.PublicRecycle
		if err := query.Order("deleted_at DESC").Find(&docs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取回收站列表失败")
			return
		}

		// 构建响应数据
		list := make([]RecyclePublicListResponse, 0, len(docs))
		for _, doc := range docs {
			list = append(list, RecyclePublicListResponse{
				ID:        doc.ID,
				DocID:     doc.DocID,
				Title:     doc.Title,
				Category:  doc.Category,
				Creator:   doc.Creator,
				UserName:  doc.UserName,
				DeletedAt: doc.DeletedAt.Format("2006-01-02 15:04:05"),
			})
		}

		middleware.ResponseSuccess(w, "获取成功", list)
		return
	}
}
