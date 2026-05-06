package system

import (
	"cmdb/pkg/middleware"
)

// DeptProject 部门项目关联表
type DeptProject struct {
	DeptID  int64  `json:"dept_id" gorm:"primaryKey;comment:部门ID"`
	Project string `json:"project" gorm:"primaryKey;size:64;comment:项目标识"`
}

// TableName 指定表名
func (DeptProject) TableName() string {
	return middleware.TableDeptProjects
}
