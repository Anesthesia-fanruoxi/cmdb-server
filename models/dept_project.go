package models

// DeptProject 部门-项目关联模型
type DeptProject struct {
	DeptID  uint   `json:"dept_id" gorm:"primaryKey;not null"`
	Project string `json:"project" gorm:"primaryKey;not null"`
}

// TableName 指定表名
func (DeptProject) TableName() string {
	return "dept_projects"
}
