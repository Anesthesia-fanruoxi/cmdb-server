package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// BatchRefreshRequest 批量恢复文档请求
type RecycleBatchRefreshRequest struct {
	IDs  []int64 `json:"ids" validate:"required,min=1"`
	Type string  `json:"type" validate:"required,oneof=doc personal public"`
}

// BatchRefresh 批量恢复文档
func RecycleBatchRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req RecycleBatchRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 处理公共文档
	if req.Type == "doc" {
		// 查询回收站文档
		var recycleList []knowledge.DocumentRecycle
		if err := tx.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		// 恢复文档
		for _, recycle := range recycleList {
			doc := &knowledge.Document{
				Title:    recycle.Title,
				Content:  recycle.Content,
				Category: recycle.Category,
				Project:  recycle.Project,
				Creator:  recycle.DeleterName,
				UserName: recycle.DeleterName,
				Version:  1,
			}

			if err := tx.Create(doc).Error; err != nil {
				tx.Rollback()
				middleware.ResponseError(w, http.StatusInternalServerError, "恢复文档失败")
				return
			}
		}

		// 删除回收站记录
		if err := tx.Delete(&knowledge.DocumentRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	// 处理个人文档
	if req.Type == "personal" {
		// 查询回收站文档
		var recycleList []knowledge.PersonalRecycle
		if err := tx.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		// 恢复文档
		for _, recycle := range recycleList {
			doc := &knowledge.Personal{
				Title:    recycle.Title,
				Content:  recycle.Content,
				UserName: recycle.UserName,
			}

			if err := tx.Create(doc).Error; err != nil {
				tx.Rollback()
				middleware.ResponseError(w, http.StatusInternalServerError, "恢复文档失败")
				return
			}
		}

		// 删除回收站记录
		if err := tx.Delete(&knowledge.PersonalRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	// 处理个人文档
	if req.Type == "public" {
		// 查询回收站文档
		var recycleList []knowledge.PublicRecycle
		if err := tx.Where("id IN ?", req.IDs).Find(&recycleList).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "查询回收站文档失败")
			return
		}

		// 恢复文档
		for _, recycle := range recycleList {
			doc := &knowledge.Public{
				Title:    recycle.Title,
				Category: recycle.Category,
				Creator:  recycle.UserName,
				Content:  recycle.Content,
				UserName: recycle.UserName,
			}

			if err := tx.Create(doc).Error; err != nil {
				tx.Rollback()
				middleware.ResponseError(w, http.StatusInternalServerError, "恢复文档失败")
				return
			}
		}

		// 删除回收站记录
		if err := tx.Delete(&knowledge.PersonalRecycle{}, "id IN ?", req.IDs).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	tx.Commit()
	middleware.ResponseSuccess(w, "恢复成功", nil)
}
