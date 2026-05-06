package document

import (
	"cmdb/models/knowledge"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// CreateRequest 创建文档请求
type CreateRequest struct {
	Title    string `json:"title" validate:"required,min=2,max=200"`
	Content  string `json:"content" validate:"required"`
	Project  string `json:"project" validate:"required,min=1,max=100"`
	Category string `json:"category" validate:"required"`
}

// DocCreate 创建文档
func DocCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 创建文档
	doc := &knowledge.Document{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Project:  req.Project,
		Creator:  userInfo.UserName,
		UserName: userInfo.UserName,
	}

	if err := database.DB.Create(doc).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建文档失败")
		return
	}

	middleware.ResponseSuccess(w, "创建成功", map[string]interface{}{
		"doc_id": doc.ID,
	})
}
