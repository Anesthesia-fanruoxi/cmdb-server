package search

import (
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// SharedKeywordCreateReq 创建请求
type SharedKeywordCreateReq struct {
	Project  string `json:"project" validate:"required"`
	Category string `json:"category" validate:"required"`
	ViewID   int64  `json:"view_id" validate:"required"`
	Keyword  string `json:"keyword" validate:"required"`
	Remark   string `json:"remark"`
}

// SharedKeywordUpdateReq 更新请求
type SharedKeywordUpdateReq struct {
	ID       int64  `json:"id" validate:"required"`
	Project  string `json:"project" validate:"required"`
	Category string `json:"category" validate:"required"`
	ViewID   int64  `json:"view_id" validate:"required"`
	Keyword  string `json:"keyword" validate:"required"`
	Remark   string `json:"remark"`
}

// SharedKeywordCreate 创建共享关键词
func SharedKeywordCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req SharedKeywordCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	keyword := elfk.SharedKeyword{
		Project:  req.Project,
		Category: req.Category,
		ViewID:   req.ViewID,
		Keyword:  req.Keyword,
		Remark:   req.Remark,
		Creator:  userInfo.UserName,
	}

	// 检查关键词是否已存在（project + category + view_id + keyword 组合唯一）
	var count int64
	database.DB.Model(&elfk.SharedKeyword{}).Where(
		"project = ? AND category = ? AND view_id = ? AND keyword = ?",
		req.Project, req.Category, req.ViewID, req.Keyword,
	).Count(&count)
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "该项目分类视图下关键词已存在")
		return
	}

	if err := database.DB.Create(&keyword).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建失败")
		return
	}

	middleware.ResponseSuccess(w, "创建成功", keyword)
}

// SharedKeywordListResp 列表响应
type SharedKeywordListResp struct {
	List  []elfk.SharedKeyword `json:"list"`
	Total int64                `json:"total"`
}

// SharedKeywordList 查询共享关键词列表（最多返回100条）
func SharedKeywordList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 搜索参数
	project := r.URL.Query().Get("project")
	category := r.URL.Query().Get("category")
	viewID := r.URL.Query().Get("view_id")
	search := r.URL.Query().Get("search") // 混合搜索：关键词、备注、用户

	query := database.DB.Model(&elfk.SharedKeyword{})

	if project != "" {
		query = query.Where("project LIKE ?", "%"+project+"%")
	}
	if category != "" {
		query = query.Where("category LIKE ?", "%"+category+"%")
	}
	if viewID != "" {
		query = query.Where("view_id = ?", viewID)
	}
	if search != "" {
		query = query.Where("keyword LIKE ? OR remark LIKE ? OR creator LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	// 查询数据（最多100条）
	var list []elfk.SharedKeyword
	if err := query.Order("created_at DESC").Limit(100).Find(&list).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	resp := SharedKeywordListResp{
		List:  list,
		Total: total,
	}

	middleware.ResponseSuccess(w, "查询成功", resp)
}

// SharedKeywordUpdate 更新共享关键词
func SharedKeywordUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req SharedKeywordUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	var keyword elfk.SharedKeyword
	if err := database.DB.First(&keyword, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "记录不存在")
		return
	}

	// 检查关键词是否已存在（排除自己，project + category + view_id + keyword 组合唯一）
	var count int64
	database.DB.Model(&elfk.SharedKeyword{}).Where(
		"project = ? AND category = ? AND view_id = ? AND keyword = ? AND id != ?",
		req.Project, req.Category, req.ViewID, req.Keyword, req.ID,
	).Count(&count)
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "该项目分类视图下关键词已存在")
		return
	}

	keyword.Project = req.Project
	keyword.Category = req.Category
	keyword.ViewID = req.ViewID
	keyword.Keyword = req.Keyword
	keyword.Remark = req.Remark

	if err := database.DB.Save(&keyword).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", keyword)
}

// SharedKeywordDelete 删除共享关键词
func SharedKeywordDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := database.DB.Delete(&elfk.SharedKeyword{}, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
