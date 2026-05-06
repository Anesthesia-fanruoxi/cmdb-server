package document

import (
	"cmdb/models/knowledge"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// DocDelete 删除文档（移入回收站）
func DocDelete(w http.ResponseWriter, r *http.Request) {
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

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 查询文档
	var doc knowledge.Document
	if err := database.DB.First(&doc, docID).Error; err != nil {
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

	if hasPermission, errMsg := auth.CheckProjectPermission(doc.Project); !hasPermission {
		middleware.ResponseError(w, http.StatusForbidden, errMsg)
		return
	}

	// 保存到回收站
	recycle := knowledge.DocumentRecycle{
		DocID:       doc.ID,
		Title:       doc.Title,
		Content:     doc.Content,
		Project:     doc.Project,
		Category:    doc.Category,
		Creator:     doc.Creator,
		DeleterName: userInfo.UserName,
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
