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

// DocHistoryDetail 获取历史版本详情
func DocHistoryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取历史版本ID
	historyID, err := strconv.ParseInt(r.URL.Query().Get("history_id"), 10, 64)
	if err != nil || historyID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的历史版本ID")
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

	// 查询历史版本
	var history knowledge.DocumentHistory
	if err := database.DB.First(&history, historyID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "历史版本不存在")
		return
	}

	// 查询原文档以验证权限
	var doc knowledge.Document
	if err := database.DB.First(&doc, history.DocID).Error; err != nil {
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

	// 构建响应数据
	type HistoryDetailResponse struct {
		ID        int64  `json:"id"`
		DocID     int64  `json:"doc_id"`
		Title     string `json:"title"`
		Category  string `json:"category"`
		Project   string `json:"project"`
		Content   string `json:"content"`
		Version   int64  `json:"version"`
		UserName  string `json:"user_name"`
		CreatedAt string `json:"created_at"`
	}

	resp := HistoryDetailResponse{
		ID:        history.ID,
		DocID:     history.DocID,
		Title:     history.Title,
		Category:  history.Category,
		Project:   history.Project,
		Content:   history.Content,
		Version:   history.Version,
		UserName:  history.UserName,
		CreatedAt: history.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	middleware.ResponseSuccess(w, "获取成功", resp)
}
