package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// CreateDeptRequest 创建部门请求
type CreateDeptRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=64"`
	Code        string `json:"code" validate:"required,min=2,max=64"`
	ParentID    int64  `json:"parent_id"`
	Description string `json:"description" validate:"max=255"`
	Sort        int    `json:"sort" validate:"min=0"`
}

// DeptCreate 创建部门
func DeptCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req CreateDeptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查部门名称和编码是否已存在
	var count int64
	database.DB.Model(&system.Department{}).Where("name = ? OR code = ?", req.Name, req.Code).Count(&count)
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "部门名称或编码已存在")
		return
	}
	if req.ParentID == 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "系统部门不能创建子部门")
	}
	// 如果有父部门，检查父部门是否存在
	if req.ParentID != 0 {
		var parentCount int64
		database.DB.Model(&system.Department{}).Where("id = ?", req.ParentID).Count(&parentCount)
		if parentCount == 0 {
			middleware.ResponseError(w, http.StatusBadRequest, "父部门不存在")
			return
		}
	}

	// 创建部门
	dept := system.Department{
		Name:        req.Name,
		Code:        req.Code,
		ParentID:    req.ParentID,
		Description: req.Description,
		Sort:        req.Sort,
	}

	if err := database.DB.Create(&dept).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建部门失败")
		return
	}

	middleware.ResponseSuccess(w, "创建成功", map[string]interface{}{
		"dept_id": dept.ID,
	})
}
