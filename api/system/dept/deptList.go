package dept

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// QueryDeptRequest 查询部门请求
type QueryDeptRequest struct {
	Name      string `json:"name" form:"name"`
	Code      string `json:"code" form:"code"`
	IsEnabled *bool  `json:"is_enabled" form:"is_enabled"`
}

// DeptList 部门列表
func DeptList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 查询所有部门
	var depts []system.Department
	if err := database.DB.Order("sort").Find(&depts).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询部门列表失败")
		return
	}

	// 构建部门树
	deptTree := buildDeptTree(depts, 0)

	// 返回结果
	middleware.ResponseSuccess(w, "查询成功", deptTree)
}

// DeptDetail 部门详情
func DeptDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL获取部门ID
	deptID := r.URL.Query().Get("id")
	if deptID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "部门ID不能为空")
		return
	}

	// 查询部门
	var dept system.Department
	if err := database.DB.First(&dept, deptID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "部门不存在")
		return
	}

	// 返回部门信息
	middleware.ResponseSuccess(w, "查询成功", dept)
}

// DeptTreeNode 部门树节点
type DeptTreeNode struct {
	system.Department
	Children []DeptTreeNode `json:"children"`
}

// buildDeptTree 构建部门树
func buildDeptTree(departments []system.Department, parentID int64) []DeptTreeNode {
	var tree []DeptTreeNode
	for _, department := range departments {
		if department.ParentID == parentID {
			node := DeptTreeNode{
				Department: department,
				Children:   buildDeptTree(departments, department.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}
