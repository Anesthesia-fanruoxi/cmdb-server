package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// RefreshRequest 恢复文档请求
type RecycleRemoveRequest struct {
	ID   int64  `json:"id" validate:"required"`
	Type string `json:"type" validate:"required,oneof=doc personal public"` // doc:公共文档 personal:个人文档
}

// Remove 彻底删除文档
func RecycleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取文档ID和类型
	docID, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	docType := r.URL.Query().Get("type")

	// 处理公共文档
	if docType == "doc" {
		// 查询回收站文档
		var recycle knowledge.DocumentRecycle
		if err := database.DB.First(&recycle, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&recycle).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	// 处理个人文档
	if docType == "personal" {
		// 查询回收站文档
		var recycle knowledge.PersonalRecycle
		if err := database.DB.First(&recycle, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&recycle).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	// 处理个人文档
	if docType == "public" {
		// 查询回收站文档
		var recycle knowledge.PublicRecycle
		if err := database.DB.First(&recycle, docID).Error; err != nil {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&recycle).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
