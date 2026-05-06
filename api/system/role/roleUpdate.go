package role

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID          int64  `json:"id" validate:"required"`                   // 角色ID，必填
	Name        string `json:"name" validate:"omitempty,min=2,max=32"`   // 角色名称，选填
	Code        string `json:"code" validate:"omitempty,min=2,max=32"`   // 角色编码，选填
	Description string `json:"description" validate:"omitempty,max=128"` // 角色描述，选填
	Level       int    `json:"level" validate:"omitempty,gte=0"`         // 角色级别，选填
}

// RoleUpdate 更新角色
func RoleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 直接解析为请求结构体
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 查询角色是否存在
	var role system.Role
	if err := database.DB.First(&role, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "角色不存在")
		return
	}
	// 不允许删除超级管理员角色
	if role.Level == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "不能修改超级管理员角色")
		return
	}
	// 构建更新数据
	updates := make(map[string]interface{})

	// 解码原始JSON以检查字段是否存在
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	var rawData map[string]json.RawMessage
	if err := decoder.Decode(&rawData); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 检查字段是否在原始请求中存在
	if _, exists := rawData["name"]; exists {
		// 检查名称是否重复
		var count int64
		if err := database.DB.Model(&system.Role{}).Where("name = ? AND id != ?", req.Name, req.ID).Count(&count).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "检查角色名称失败")
			return
		}
		if count > 0 {
			middleware.ResponseError(w, http.StatusConflict, "角色名称已存在")
			return
		}
		updates["name"] = req.Name
	}

	if _, exists := rawData["code"]; exists {
		// 检查编码是否重复
		var count int64
		if err := database.DB.Model(&system.Role{}).Where("code = ? AND id != ?", req.Code, req.ID).Count(&count).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "检查角色编码失败")
			return
		}
		if count > 0 {
			middleware.ResponseError(w, http.StatusConflict, "角色编码已存在")
			return
		}
		updates["code"] = req.Code
	}

	if _, exists := rawData["description"]; exists {
		updates["description"] = req.Description
	}

	if _, exists := rawData["level"]; exists {
		updates["level"] = int64(req.Level)
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 执行更新
	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新角色失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
