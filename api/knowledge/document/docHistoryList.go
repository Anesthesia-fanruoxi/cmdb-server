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

// DocHistoryList 获取文档历史版本列表
func DocHistoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	docID, err := strconv.ParseInt(r.URL.Query().Get("doc_id"), 10, 64)
	if err != nil || docID == 0 {
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

	// 先查询文档信息以验证权限
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

	// 查询历史版本列表
	var histories []knowledge.DocumentHistory
	if err := database.DB.Where("doc_id = ?", docID).
		Order("version DESC").
		Find(&histories).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取历史版本失败")
		return
	}

	// 构建响应数据
	type HistoryResponse struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		Category  string `json:"category"`
		Project   string `json:"project"`
		Version   int64  `json:"version"`
		UserName  string `json:"user_name"`
		CreatedAt string `json:"created_at"`
	}

	list := make([]HistoryResponse, 0, len(histories))
	for _, h := range histories {
		list = append(list, HistoryResponse{
			ID:        h.ID,
			Title:     h.Title,
			Category:  h.Category,
			Project:   h.Project,
			Version:   h.Version,
			UserName:  h.UserName,
			CreatedAt: h.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	middleware.ResponseSuccess(w, "获取成功", list)
}
