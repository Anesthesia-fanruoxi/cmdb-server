package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// BatchRemoveRequest 批量删除文档请求
type RecycleBatchRemoveRequest struct {
	IDs  []int64 `json:"ids" validate:"required,min=1"`
	Type string  `json:"type" validate:"required,oneof=doc personal public"`
}

// BatchRemove 批量彻底删除文档
func RecycleBatchRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req RecycleBatchRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 处理公共文档
	if req.Type == "doc" {
		// 查询回收站文档
		var recycleList []knowledge.DocumentRecycle
		if err := database.DB.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		if len(recycleList) == 0 {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&knowledge.DocumentRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	// 处理个人文档
	if req.Type == "personal" {
		// 查询回收站文档
		var recycleList []knowledge.PersonalRecycle
		if err := database.DB.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		if len(recycleList) == 0 {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&knowledge.PersonalRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	// 处理个人文档
	if req.Type == "public" {
		// 查询回收站文档
		var recycleList []knowledge.PublicRecycle
		if err := database.DB.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		if len(recycleList) == 0 {
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		// 彻底删除文档
		if err := database.DB.Delete(&knowledge.PublicRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
			return
		}
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
