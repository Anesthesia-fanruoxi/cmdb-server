package document

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"

	"cmdb/models/system"
)

// DocDetail 获取文档详情
func DocDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	middleware.ResponseSuccess(w, "获取成功", doc)
}
