package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// DeptProjectDetail 获取部门关联的项目列表
func DeptProjectDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取部门ID
	deptID := r.URL.Query().Get("dept_id")
	if deptID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "部门ID不能为空")
		return
	}

	// 检查部门是否存在
	var dept system.Department
	if err := database.DB.First(&dept, deptID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "部门不存在")
		return
	}

	// 查询部门关联的项目
	var deptProjects []system.DeptProject
	if err := database.DB.Where("dept_id = ?", deptID).Find(&deptProjects).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询部门项目关联失败")
		return
	}

	// 提取项目代码列表
	projects := make([]string, 0, len(deptProjects))
	for _, dp := range deptProjects {
		projects = append(projects, dp.Project)
	}

	// 返回结果
	result := map[string]interface{}{
		"dept_id": dept.ID,
		"project": projects,
	}

	middleware.ResponseSuccess(w, "查询成功", result)
}
