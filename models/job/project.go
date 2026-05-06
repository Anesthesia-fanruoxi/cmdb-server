package job

import (
	"cmdb/pkg/middleware"
	"database/sql/driver"
	"encoding/json"
)

// JobProject 任务-项目关联表
type JobProject struct {
	JobID        int64   `json:"job_id" gorm:"primary_key"`      // 任务ID
	Project      string  `json:"project" gorm:"primary_key"`     // 项目标识
	ScriptParams JSONMap `json:"script_params" gorm:"type:json"` // 脚本参数
}

// JSONMap 自定义类型用于处理JSON字段
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// TableName 指定表名
func (JobProject) TableName() string {
	return middleware.TableJobProject
}
