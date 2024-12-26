package models

type ProjectDict struct {
	Project     string `gorm:"primarykey;type:varchar(64)" json:"project"` // 项目代码
	ProjectName string `gorm:"type:varchar(128)" json:"project_name"`      // 项目中文名称
}
