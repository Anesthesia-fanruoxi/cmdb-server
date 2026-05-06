package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// DictRecord 项目字典表
type DictProject struct {
	ID          int64     `json:"id" gorm:"primarykey"`
	Project     string    `json:"project" gorm:"size:64;uniqueIndex;comment:项目code"`
	ProjectName string    `json:"project_name" gorm:"size:64;uniqueIndex;comment:项目名称"`
	KeyName     string    `json:"key_name" gorm:"size:64;comment:键字段名"`
	ValueName   string    `json:"value_name" gorm:"size:64;comment:值字段名"`
	CreatedBy   int64     `json:"created_by" gorm:"comment:创建人ID"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (DictProject) TableName() string {
	return middleware.TableDictRecords
}
