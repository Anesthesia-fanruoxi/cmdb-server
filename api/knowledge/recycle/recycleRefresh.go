package recycle

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// RefreshRequest 恢复文档请求
type RecycleRefreshRequest struct {
	ID   int64  `json:"id" validate:"required"`
	Type string `json:"type" validate:"required,oneof=doc personal public"` // doc:公共文档 personal:个人文档
}

// Refresh 从回收站恢复文档
func RecycleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req RecycleRefreshRequest
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
		// 恢复公共文档
		var recycle knowledge.DocumentRecycle
		if err := tx.First(&recycle, req.ID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

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

		// 删除回收站记录
		if err := tx.Delete(&recycle).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	// 处理个人文档
	if req.Type == "personal" {
		var recycle knowledge.PersonalRecycle
		if err := tx.First(&recycle, req.ID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

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

		// 删除回收站记录
		if err := tx.Delete(&recycle).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	// 处理公共文档
	if req.Type == "public" {
		var recycle knowledge.PublicRecycle
		if err := tx.First(&recycle, req.ID).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusNotFound, "回收站文档不存在")
			return
		}

		doc := &knowledge.Public{
			Title:    recycle.Title,
			Content:  recycle.Content,
			Category: recycle.Category,
			Creator:  recycle.Creator,
			UserName: recycle.UserName,
		}

		if err := tx.Create(doc).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "恢复文档失败")
			return
		}

		// 删除回收站记录
		if err := tx.Delete(&recycle).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "删除回收站记录失败")
			return
		}
	}

	tx.Commit()
	middleware.ResponseSuccess(w, "恢复成功", nil)
}
