package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io"
	"net/http"
)

// UpdateDeptRequest 更新部门请求
type UpdateDeptRequest struct {
	ID          int64  `json:"id" validate:"required"`
	Name        string `json:"name" validate:"omitempty,min=2,max=64"`
	Code        string `json:"code" validate:"omitempty,min=2,max=64"`
	ParentID    int64  `json:"parent_id"`
	Description string `json:"description" validate:"max=255"`
	Sort        int    `json:"sort" validate:"min=0"`
}

// DeptUpdate 更新部门
func DeptUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 解析为原始 map 以检查字段是否存在
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 解析请求体到结构体
	var req UpdateDeptRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查部门是否存在
	var dept system.Department
	if err := database.DB.First(&dept, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "部门不存在")
		return
	}

	// 检查是否是系统部门
	if req.ID == SystemDeptID {
		middleware.ResponseError(w, http.StatusBadRequest, "系统部门不能修改")
		return
	}

	// 如果更新名称或编码，检查是否与其他部门冲突
	if req.Name != "" || req.Code != "" {
		var count int64
		query := database.DB.Model(&system.Department{}).Where("id != ?", req.ID)
		if req.Name != "" {
			query = query.Where("name = ?", req.Name)
		}
		if req.Code != "" {
			query = query.Where("code = ?", req.Code)
		}
		query.Count(&count)
		if count > 0 {
			middleware.ResponseError(w, http.StatusBadRequest, "部门名称或编码已存在")
			return
		}
	}

	// 如果有父部门，检查父部门是否存在且不能是自己或自己的子部门
	if req.ParentID != 0 {
		if req.ParentID == req.ID {
			middleware.ResponseError(w, http.StatusBadRequest, "父部门不能是自己")
			return
		}
		var parentCount int64
		database.DB.Model(&system.Department{}).Where("id = ?", req.ParentID).Count(&parentCount)
		if parentCount == 0 {
			middleware.ResponseError(w, http.StatusBadRequest, "父部门不存在")
			return
		}
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	// 只更新提供的字段
	if _, exists := rawData["name"]; exists {
		updates["name"] = req.Name
	}
	if _, exists := rawData["code"]; exists {
		updates["code"] = req.Code
	}
	if _, exists := rawData["parent_id"]; exists {
		updates["parent_id"] = req.ParentID
	}
	if _, exists := rawData["description"]; exists {
		updates["description"] = req.Description
	}
	if _, exists := rawData["sort"]; exists {
		updates["sort"] = req.Sort
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 执行更新
	if err := database.DB.Model(&dept).Updates(updates).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新部门失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
