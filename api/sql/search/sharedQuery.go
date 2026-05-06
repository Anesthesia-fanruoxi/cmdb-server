package search

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// SharedQueryCreateReq 创建请求
type SharedQueryCreateReq struct {
	Project string `json:"project" validate:"required"`
	DBName  string `json:"db_name" validate:"required"`
	Query   string `json:"query" validate:"required"`
	Remark  string `json:"remark"`
}

// SharedQueryUpdateReq 更新请求
type SharedQueryUpdateReq struct {
	ID      int64  `json:"id" validate:"required"`
	Project string `json:"project" validate:"required"`
	DBName  string `json:"db_name" validate:"required"`
	Query   string `json:"query" validate:"required"`
	Remark  string `json:"remark"`
}

// SharedQueryListResp 列表响应
type SharedQueryListResp struct {
	List  []sql.SharedQuery `json:"list"`
	Total int64             `json:"total"`
}

// SharedQueryCreate 创建共享查询
func SharedQueryCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	var req SharedQueryCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	query := sql.SharedQuery{
		Project: req.Project,
		DBName:  req.DBName,
		Query:   req.Query,
		Remark:  req.Remark,
		Creator: userInfo.UserName,
	}

	if err := database.DB.Create(&query).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建失败")
		return
	}

	middleware.ResponseSuccess(w, "创建成功", query)
}

// SharedQueryList 查询共享查询列表（最多返回100条）
func SharedQueryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 搜索参数
	project := r.URL.Query().Get("project")
	dbName := r.URL.Query().Get("db_name")
	search := r.URL.Query().Get("search") // 混合搜索：备注、用户

	query := database.DB.Model(&sql.SharedQuery{})

	if project != "" {
		query = query.Where("project LIKE ?", "%"+project+"%")
	}
	if dbName != "" {
		query = query.Where("db_name LIKE ?", "%"+dbName+"%")
	}
	if search != "" {
		query = query.Where("remark LIKE ? OR creator LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	// 查询数据（最多100条）
	var list []sql.SharedQuery
	if err := query.Order("created_at DESC").Limit(100).Find(&list).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	resp := SharedQueryListResp{
		List:  list,
		Total: total,
	}

	middleware.ResponseSuccess(w, "查询成功", resp)
}

// SharedQueryUpdate 更新共享查询
func SharedQueryUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req SharedQueryUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	var query sql.SharedQuery
	if err := database.DB.First(&query, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "记录不存在")
		return
	}

	query.Project = req.Project
	query.DBName = req.DBName
	query.Query = req.Query
	query.Remark = req.Remark

	if err := database.DB.Save(&query).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", query)
}

// SharedQueryDelete 删除共享查询
func SharedQueryDelete(w http.ResponseWriter, r *http.Request) {
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

	if err := database.DB.Delete(&sql.SharedQuery{}, id).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
