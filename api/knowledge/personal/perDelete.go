package personal

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// PersonalDelete 删除个人文档（移入回收站）
func PersonalDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取文档ID
	docID, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的文档ID")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 查询文档并验证所有者
	var doc knowledge.Personal
	if err := database.DB.Where("id = ? AND user_name = ?", docID, userInfo.UserName).First(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "文档不存在或无权访问")
		return
	}

	// 保存到回收站
	recycle := knowledge.PersonalRecycle{
		DocID:    doc.ID,
		Title:    doc.Title,
		Category: doc.Category,
		Content:  doc.Content,
		UserName: userInfo.UserName,
	}

	if err := database.DB.Create(&recycle).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "移入回收站失败")
		return
	}

	// 删除原文档（软删除）
	if err := database.DB.Delete(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除文档失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
