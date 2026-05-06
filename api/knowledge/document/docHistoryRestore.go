package document

import (
	"cmdb/models/knowledge"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// DocHistoryRestore 恢复历史版本
func DocHistoryRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var req struct {
		HistoryID int64 `json:"history_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 查询历史版本
	var history knowledge.DocumentHistory
	if err := tx.First(&history, req.HistoryID).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusNotFound, "历史版本不存在")
		return
	}

	// 查询当前文档
	var doc knowledge.Document
	if err := tx.First(&doc, history.DocID).Error; err != nil {
		tx.Rollback()
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
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, errMsg)
		return
	}

	// 保存当前版本到历史记录
	currentHistory := knowledge.DocumentHistory{
		DocID:    doc.ID,
		Title:    doc.Title,
		Content:  doc.Content,
		Category: doc.Category,
		Project:  doc.Project,
		Version:  doc.Version,
		UserName: userInfo.UserName,
	}
	if err := tx.Create(&currentHistory).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "保存当前版本失败")
		return
	}

	// 恢复历史版本到当前文档
	doc.Title = history.Title
	doc.Content = history.Content
	doc.Category = history.Category
	doc.Project = history.Project
	doc.UserName = userInfo.UserName
	doc.Version++ // 版本号加1

	if err := tx.Save(&doc).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "恢复历史版本失败")
		return
	}

	// 删除已恢复的历史版本记录
	if err := tx.Delete(&history).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除历史版本失败")
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(w, "恢复成功", doc)
}
