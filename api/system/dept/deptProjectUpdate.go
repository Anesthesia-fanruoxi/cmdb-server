package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
)

// UpdateDeptProject 更新部门项目关联请求
type UpdateDeptProject struct {
	DeptID  int64    `json:"dept_id" validate:"required"`
	Project []string `json:"project" validate:"required"`
}

// DeptProjectUpdate 更新部门项目关联
func DeptProjectUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var req UpdateDeptProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	if err := database.DB.First(&dept, req.DeptID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "部门不存在")
		return
	}

	// 检查是否是系统部门
	if req.DeptID == SystemDeptID {
		middleware.ResponseError(w, http.StatusForbidden, "系统部门的项目关联不能修改")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 删除原有的部门项目关联
	if err := tx.Where("dept_id = ?", req.DeptID).Delete(&system.DeptProject{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除原有项目关联失败")
		return
	}

	// 创建新的部门项目关联
	for _, projectCode := range req.Project {
		deptProject := system.DeptProject{
			DeptID:  req.DeptID,
			Project: projectCode,
		}
		if err := tx.Create(&deptProject).Error; err != nil {
			tx.Rollback()
			middleware.ResponseError(w, http.StatusInternalServerError, "创建项目关联失败")
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新项目关联失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
