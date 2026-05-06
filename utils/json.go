package utils

import (
	"cmdb/pkg/middleware"
	"encoding/json"
)

// ToJSON 将对象转换为JSON字符串
func ToJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		middleware.Logger.Error("转换JSON失败: %v\n", err)
		return "{}"
	}
	return string(bytes)
}

// FromJSON 将JSON字符串转换为对象
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
