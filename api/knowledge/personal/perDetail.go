package personal

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strconv"
)

// PersonalDetail 获取个人文档详情
func PersonalDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取文档ID
	docID, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的文档ID")
		return
	}

	// 查询文档，并验证所有者
	var doc knowledge.Personal
	if err := database.DB.Where("id = ? AND user_name = ?", docID, userInfo.UserName).First(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "文档不存在或无权访问")
		return
	}

	middleware.ResponseSuccess(w, "获取成功", doc)
}
