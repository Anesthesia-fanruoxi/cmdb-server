package models

// DictRecord 字典记录表
type DictRecord struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
	DictName  string `json:"dict_name" gorm:"type:varchar(64);not null;uniqueIndex;comment:字典名称"`
	DictTable string `json:"table_name" gorm:"column:table_name;type:varchar(64);not null;uniqueIndex;comment:表名"`
	KeyName   string `json:"key_name" gorm:"type:varchar(64);not null;comment:键字段名"`
	ValueName string `json:"value_name" gorm:"type:varchar(64);not null;comment:值字段名"`
	CreatedBy uint   `json:"created_by" gorm:"not null;comment:创建人ID"`
	CreatedAt string `json:"created_at" gorm:"type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间"`
}

// TableName 指定表名
func (DictRecord) TableName() string {
	return "dict_records"
}
