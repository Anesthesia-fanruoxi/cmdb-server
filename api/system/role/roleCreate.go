package role

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=32"`
	Code        string `json:"code" validate:"required,min=2,max=32"`
	Description string `json:"description" validate:"max=128"`
	Level       int    `json:"level" validate:"required,gte=0"`
}

// RoleCreate 创建角色
func RoleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查角色名称是否已存在
	var count int64
	database.DB.Model(&system.Role{}).Where("name = ? OR code = ?", req.Name, req.Code).Count(&count)
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "角色名称或编码已存在")
		return
	}

	// 创建角色
	role := system.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Level:       int64(req.Level),
	}

	if err := database.DB.Create(&role).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建角色失败")
		return
	}

	if req.Level == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "不能创建系统角色")
	}

	middleware.ResponseSuccess(w, "创建成功", map[string]interface{}{
		"role_id": role.ID,
	})
}
