package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// RoleResponse 角色返回结构体
type RoleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// 将Role模型转换为RoleResponse
func convertToRoleResponse(role models.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListRoles 获取角色列表
func ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var roles []models.Role
	if err := initData.GetDB().Find(&roles).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取角色列表失败")
		return
	}

	// 转换为响应格式
	var response []RoleResponse
	for _, role := range roles {
		response = append(response, convertToRoleResponse(role))
	}

	utils.Success(w, map[string]interface{}{
		"total": len(response),
		"list":  response,
	})
}

// CreateRole 创建角色
func CreateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填字段
	if role.Name == "" || role.Code == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色名称和编码不能为空")
		return
	}

	// 检查角色编码是否已存在
	var count int64
	if err := initData.GetDB().Model(&models.Role{}).Where("code = ?", role.Code).Count(&count).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建角色失败")
		return
	}

	if count > 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色编码已存在")
		return
	}

	// 创建角色
	if err := initData.GetDB().Create(&role).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建角色失败")
		return
	}

	utils.Success(w, convertToRoleResponse(role))
}

// UpdateRole 更新角色
func UpdateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if role.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色ID不能为空")
		return
	}

	// 检查角色是否存在
	var existingRole models.Role
	if err := initData.GetDB().First(&existingRole, role.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "角色不存在")
		return
	}

	// 如果修改了角色编码，检查新编码是否已存在
	if role.Code != "" && role.Code != existingRole.Code {
		var count int64
		if err := initData.GetDB().Model(&models.Role{}).Where("code = ? AND id != ?", role.Code, role.ID).Count(&count).Error; err != nil {
			utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新角色失败")
			return
		}

		if count > 0 {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色编码已存在")
			return
		}
	}

	// 更新角色
	if err := initData.GetDB().Model(&existingRole).Updates(role).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新角色失败")
		return
	}

	// 重新查询完整的角色信息
	if err := initData.GetDB().First(&existingRole, role.ID).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取更新后的角色信息失败")
		return
	}

	utils.Success(w, convertToRoleResponse(existingRole))
}

// DeleteRole 删除角色
func DeleteRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if req.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "角色ID不能为空")
		return
	}

	// 检查角色是否存在
	var role models.Role
	if err := initData.GetDB().First(&role, req.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "角色不存在")
		return
	}

	// 检查是否是内置角色
	if role.Code == "admin" || role.Code == "user" {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "不能删除内置角色")
		return
	}

	// 检查是否有用户正在使用该角色
	var count int64
	if err := initData.GetDB().Model(&models.User{}).Where("role_id = ?", req.ID).Count(&count).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除角色失败")
		return
	}

	if count > 0 {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "该角色下还有用户，不能删除")
		return
	}

	// 删除角色
	if err := initData.GetDB().Delete(&role).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除角色失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"id": role.ID,
	})
}
