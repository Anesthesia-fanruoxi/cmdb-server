package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

const (
	// SystemDeptID 系统部门ID
	SystemDeptID = 1
)

// DeptDelete 删除部门
func DeptDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取部门ID
	deptID := r.URL.Query().Get("id")
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

	// 检查是否是系统部门
	if dept.ID == SystemDeptID {
		middleware.ResponseError(w, http.StatusBadRequest, "系统部门不能删除")
		return
	}

	// 获取要删除的部门及其所有子部门的ID列表
	var deptIDs []int64
	recursiveQuery := `WITH RECURSIVE dept_tree AS (
		SELECT id FROM %s WHERE id = ?
		UNION ALL
		SELECT d.id FROM %s d
		INNER JOIN dept_tree dt ON d.parent_id = dt.id
	) SELECT id FROM dept_tree`

	if err := database.DB.Raw(
		recursiveQuery,
		middleware.TableDepartments,
		middleware.TableDepartments,
	).Scan(&deptIDs).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取部门树失败")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 1. 将这些部门下的所有用户转移到默认部门
	if err := tx.Table(middleware.TableUsers).
		Where("dept_id IN ?", deptIDs).
		Update("dept_id", 99).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "转移用户失败")
		return
	}

	// 2. 删除这些部门在部门项目关联表中的关联关系
	if err := tx.Table(middleware.TableDeptProjects).
		Where("dept_id IN ?", deptIDs).
		Delete(&system.DeptProject{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除部门项目关联失败")
		return
	}

	// 3. 删除这些部门
	if err := tx.Table(middleware.TableDepartments).
		Where("id IN ?", deptIDs).
		Delete(&system.Department{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除部门失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除部门失败")
		return
	}

	middleware.ResponseSuccess(w, "删除成功", nil)
}
