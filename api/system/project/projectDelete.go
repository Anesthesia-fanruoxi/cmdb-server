package project

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// ProjectDelete 删除项目
// DELETE /api/system/project/delete?project=xxx&force=true
func ProjectDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持DELETE请求")
		return
	}

	// 获取项目标识
	projectID := r.URL.Query().Get("project")
	if projectID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 获取强制删除参数
	forceDelete := r.URL.Query().Get("force") == "true"

	// 检查项目是否存在
	var project system.ProjectDict
	if err := database.DB.Where("project = ?", projectID).First(&project).Error; err != nil {
		middleware.Logger.Errorf("查询项目失败: %v", err)
		middleware.ResponseError(w, http.StatusNotFound, "项目不存在")
		return
	}

	// 检查是否有关联的部门
	var deptCount int64
	if err := database.DB.Model(&system.DeptProject{}).Where("project = ?", projectID).Count(&deptCount).Error; err != nil {
		middleware.Logger.Errorf("检查部门关联失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "检查部门关联失败")
		return
	}

	if deptCount > 0 {
		if !forceDelete {
			// 非强制删除，查询并返回关联的部门信息
			var deptProjects []system.DeptProject
			if err := database.DB.Where("project = ?", projectID).Find(&deptProjects).Error; err != nil {
				middleware.Logger.Errorf("查询关联部门失败: %v", err)
				middleware.ResponseError(w, http.StatusInternalServerError, "查询关联部门失败")
				return
			}

			// 获取部门ID列表
			deptIDs := make([]int64, len(deptProjects))
			for i, dp := range deptProjects {
				deptIDs[i] = dp.DeptID
			}

			// 查询部门名称
			var depts []system.Department
			if err := database.DB.Where("id IN ?", deptIDs).Find(&depts).Error; err != nil {
				middleware.Logger.Errorf("查询部门信息失败: %v", err)
				middleware.ResponseError(w, http.StatusInternalServerError, "查询部门信息失败")
				return
			}

			// 构建部门信息列表
			deptList := make([]map[string]interface{}, len(depts))
			for i, d := range depts {
				deptList[i] = map[string]interface{}{
					"id":        d.ID,
					"dept_name": d.Name,
				}
			}

			// 返回关联部门信息
			middleware.ResponseSuccess(w, "该项目已关联部门，请确认是否强制删除", map[string]interface{}{
				"has_relation": true,
				"dept_count":   deptCount,
				"depts":        deptList,
			})
			return
		}
		// 强制删除：先删除关联关系
		if err := database.DB.Where("project = ?", projectID).Delete(&system.DeptProject{}).Error; err != nil {
			middleware.Logger.Errorf("删除部门关联失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "删除部门关联失败")
			return
		}
	}

	// 删除项目
	if err := database.DB.Delete(&project).Error; err != nil {
		middleware.Logger.Errorf("删除项目失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "删除项目失败")
		return
	}

	middleware.ResponseSuccess(w, "删除项目成功", map[string]interface{}{
		"project": projectID,
	})
}
