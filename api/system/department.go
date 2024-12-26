package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// DepartmentResponse 部门返回结构体
type DepartmentResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	ParentID    *uint  `json:"parent_id"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// 将Department模型转换为DepartmentResponse
func convertToDepartmentResponse(dept models.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:          dept.ID,
		Name:        dept.Name,
		Code:        dept.Code,
		ParentID:    dept.ParentID,
		Description: dept.Description,
		CreatedAt:   dept.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   dept.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListDepartments 获取部门列表
func ListDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var departments []models.Department
	if err := initData.GetDB().Find(&departments).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取部门列表失败")
		return
	}

	// 转换为响应格式
	var response []DepartmentResponse
	for _, dept := range departments {
		response = append(response, convertToDepartmentResponse(dept))
	}

	utils.Success(w, map[string]interface{}{
		"total": len(response),
		"list":  response,
	})
}

// CreateDepartment 创建部门
func CreateDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var dept models.Department
	if err := json.NewDecoder(r.Body).Decode(&dept); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填字段
	if dept.Name == "" || dept.Code == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门名称和编码不能为空")
		return
	}

	// 检查部门编码是否已存在
	var count int64
	if err := initData.GetDB().Model(&models.Department{}).Where("code = ?", dept.Code).Count(&count).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建部门失败")
		return
	}

	if count > 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门编码已存在")
		return
	}

	// 如果指定了父部门，检查父部门是否存在
	if dept.ParentID != nil && *dept.ParentID != 0 {
		var parent models.Department
		if err := initData.GetDB().First(&parent, *dept.ParentID).Error; err != nil {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "父部门不存在")
			return
		}
	}

	// 创建部门
	if err := initData.GetDB().Create(&dept).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建部门失败")
		return
	}

	utils.Success(w, convertToDepartmentResponse(dept))
}

// UpdateDepartment 更新部门
func UpdateDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var dept models.Department
	if err := json.NewDecoder(r.Body).Decode(&dept); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if dept.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门ID不能为空")
		return
	}

	// 检查部门是否存在
	var existingDept models.Department
	if err := initData.GetDB().First(&existingDept, dept.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "部门不存在")
		return
	}

	// 如果修改了部门编码，检查新编码是否已存在
	if dept.Code != "" && dept.Code != existingDept.Code {
		var count int64
		if err := initData.GetDB().Model(&models.Department{}).Where("code = ? AND id != ?", dept.Code, dept.ID).Count(&count).Error; err != nil {
			utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新部门失败")
			return
		}

		if count > 0 {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门编码已存在")
			return
		}
	}

	// 如果指定了父部门，检查父部门是否存在且不是自己
	if dept.ParentID != nil && *dept.ParentID != 0 {
		if *dept.ParentID == dept.ID {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "父部门不能是自己")
			return
		}

		var parent models.Department
		if err := initData.GetDB().First(&parent, *dept.ParentID).Error; err != nil {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "父部门不存在")
			return
		}
	}

	// 更新部门
	if err := initData.GetDB().Model(&existingDept).Updates(dept).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新部门失败")
		return
	}

	// 重新查询完整的部门信息
	if err := initData.GetDB().First(&existingDept, dept.ID).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取更新后的部门信息失败")
		return
	}

	utils.Success(w, convertToDepartmentResponse(existingDept))
}

// DeleteDepartment 删除部门
func DeleteDepartments(w http.ResponseWriter, r *http.Request) {
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
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门ID不能为空")
		return
	}

	// 检查部门是否存在
	var dept models.Department
	if err := initData.GetDB().First(&dept, req.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "部门不存在")
		return
	}

	// 检查是否是默认部门
	if dept.Code == "default" {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "不能删除默认部门")
		return
	}

	// 检查是否有子部门
	var childCount int64
	if err := initData.GetDB().Model(&models.Department{}).Where("parent_id = ?", req.ID).Count(&childCount).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除部门失败")
		return
	}

	if childCount > 0 {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "该部门下还有子部门，不能删除")
		return
	}

	// 检查是否有用户
	var userCount int64
	if err := initData.GetDB().Model(&models.User{}).Where("dept_id = ?", req.ID).Count(&userCount).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除部门失败")
		return
	}

	if userCount > 0 {
		utils.Error(w, http.StatusForbidden, utils.FORBIDDEN, "该部门下还有用户，不能删除")
		return
	}

	// 删除部门
	if err := initData.GetDB().Delete(&dept).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除部门失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"id": dept.ID,
	})
}
