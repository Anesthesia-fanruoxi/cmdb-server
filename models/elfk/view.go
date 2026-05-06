package elfk

import (
	"cmdb/pkg/middleware"
	"time"
)

// View 视图表结构
type View struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null;comment:视图名称"`
	Project      string    `json:"project" gorm:"type:varchar(50);not null;comment:所属项目"`
	Category     string    `json:"category" gorm:"type:varchar(50);not null;comment:所属分类"`
	LogType      string    `json:"log_type" gorm:"type:varchar(50);not null;comment:日志类型"`
	IndexPattern string    `json:"index_pattern" gorm:"type:varchar(255);not null;comment:索引模式"`
	TimeField    string    `json:"time_field" gorm:"type:varchar(50);not null;comment:时间字段名称"`
	TimeFormat   string    `json:"time_format" gorm:"type:varchar(50);not null;comment:时间格式"`
	Description  string    `json:"description" gorm:"type:varchar(255);comment:视图描述"`
	CreateTime   time.Time `json:"create_time" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime   time.Time `json:"update_time" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

type ViewResponse struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null;comment:视图名称"`
	Project      string    `json:"project" gorm:"type:varchar(50);not null;comment:所属项目"`
	Category     string    `json:"category" gorm:"type:varchar(50);not null;comment:所属分类"`
	LogType      string    `json:"log_type" gorm:"type:varchar(50);not null;comment:日志类型"`
	IndexPattern string    `json:"index_pattern" gorm:"type:varchar(255);not null;comment:索引模式"`
	TimeField    string    `json:"time_field" gorm:"type:varchar(50);not null;comment:时间字段名称"`
	TimeFormat   string    `json:"time_format" gorm:"type:varchar(50);not null;comment:时间格式"`
	Description  string    `json:"description" gorm:"type:varchar(255);comment:视图描述"`
	UpdateTime   time.Time `json:"update_time" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

// TableName 指定表名
func (v *View) TableName() string {
	return middleware.TableView
}
