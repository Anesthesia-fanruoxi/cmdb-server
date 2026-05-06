package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// DictRecord 字典记录
type DictRecord struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	DictName    string    `json:"dictName" gorm:"column:dict_name"`
	TargetTable string    `json:"targetTable" gorm:"column:target_table"`
	KeyName     string    `json:"keyName" gorm:"column:key_name"`
	ValueName   string    `json:"valueName" gorm:"column:value_name"`
	Color       string    `json:"color" gorm:"column:color"`
	CreatedBy   int64     `json:"createdBy" gorm:"column:created_by"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
}

// TableName 指定表名
func (DictRecord) TableName() string {
	return middleware.TableDictRecords
}
