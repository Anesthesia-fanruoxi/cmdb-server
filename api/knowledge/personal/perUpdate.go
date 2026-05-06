package personal

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// UpdateRequest 更新个人文档请求
type UpdateRequest struct {
	Id       uint   `json:"id" validate:"required"`
	Title    string `json:"title" validate:"required,min=2,max=200"`
	Category string `json:"category" validate:"required,min=1,max=200"`
	Content  string `json:"content" validate:"required"`
}

// PersonalUpdate 更新个人文档
func PersonalUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求体
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查文档是否存在并验证所有者
	var doc knowledge.Personal
	if err := database.DB.Where("id = ? AND user_name = ?", req.Id, userInfo.UserName).First(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "文档不存在或无权访问")
		return
	}

	// 更新文档
	doc.Title = req.Title
	doc.Content = req.Content
	doc.Category = req.Category
	doc.UserName = userInfo.UserName

	if err := database.DB.Save(&doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新文档失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", doc)
}
