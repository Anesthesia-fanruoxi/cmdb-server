package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// RecycleDocDetailResponse 文档回收站详情响应
type RecycleDocDetailResponse struct {
	ID          int64  `json:"id"`
	DocID       int64  `json:"doc_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Category    string `json:"category"`
	Project     string `json:"project"`
	DeleterName string `json:"deleter_name"`
	DeletedAt   string `json:"deleted_at"`
}

// RecyclePersonalDetailResponse 个人文档回收站详情响应
type RecyclePersonalDetailResponse struct {
	ID        int64  `json:"id"`
	DocID     int64  `json:"doc_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserName  string `json:"user_name"`
	DeletedAt string `json:"deleted_at"`
}

// RecyclePublicDetailResponse 对外文档回收站详情响应
type RecyclePublicDetailResponse struct {
	ID        int64  `json:"id"`
	DocID     int64  `json:"doc_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	Creator   string `json:"creator"`
	UserName  string `json:"user_name"`
	DeletedAt string `json:"deleted_at"`
}

// RecycleDetail 获取回收站文档详情
func RecycleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	docType := r.URL.Query().Get("type")
	docID := r.URL.Query().Get("id")

	// 处理公共文档回收站详情
	if docType == "doc" {
		var doc knowledge.DocumentRecycle
		if err := database.DB.First(&doc, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取文档详情失败")
			return
		}

		detail := RecycleDocDetailResponse{
			ID:          doc.ID,
			DocID:       doc.DocID,
			Title:       doc.Title,
			Content:     doc.Content,
			Category:    doc.Category,
			Project:     doc.Project,
			DeleterName: doc.DeleterName,
			DeletedAt:   doc.DeletedAt.Format("2006-01-02 15:04:05"),
		}

		middleware.ResponseSuccess(w, "获取成功", detail)
		return
	}

	// 处理个人文档回收站详情
	if docType == "personal" {
		var doc knowledge.PersonalRecycle
		if err := database.DB.First(&doc, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取文档详情失败")
			return
		}

		detail := RecyclePersonalDetailResponse{
			ID:        doc.ID,
			DocID:     doc.DocID,
			Title:     doc.Title,
			Content:   doc.Content,
			UserName:  doc.UserName,
			DeletedAt: doc.DeletedAt.Format("2006-01-02 15:04:05"),
		}

		middleware.ResponseSuccess(w, "获取成功", detail)
		return
	}

	// 处理对外文档回收站详情
	if docType == "public" {
		var doc knowledge.PublicRecycle
		if err := database.DB.First(&doc, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取文档详情失败")
			return
		}

		detail := RecyclePublicDetailResponse{
			ID:        doc.ID,
			DocID:     doc.DocID,
			Title:     doc.Title,
			Content:   doc.Content,
			Category:  doc.Category,
			Creator:   doc.Creator,
			UserName:  doc.UserName,
			DeletedAt: doc.DeletedAt.Format("2006-01-02 15:04:05"),
		}

		middleware.ResponseSuccess(w, "获取成功", detail)
		return
	}

	middleware.ResponseError(w, http.StatusBadRequest, "无效的文档类型")
}
