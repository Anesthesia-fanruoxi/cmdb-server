package models

import (
	"gorm.io/gorm"
)

// Dict 字典模型
type Dict struct {
	gorm.Model
	Type        string `json:"type" gorm:"type:varchar(50);not null;comment:字典类型"`
	Key         string `json:"key" gorm:"type:varchar(50);not null;comment:字典键"`
	Value       string `json:"value" gorm:"type:varchar(255);not null;comment:字典值"`
	Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
	Sort        int    `json:"sort" gorm:"type:int;default:0;comment:排序"`
	IsEnabled   bool   `json:"is_enabled" gorm:"type:tinyint(1);default:1;comment:是否启用"`
}
