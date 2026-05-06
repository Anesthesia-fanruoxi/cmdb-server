package key

import (
	"cmdb/models/system"
)

// GetAgentURL 根据项目获取agent地址
func GetAgentURL(project string) (string, error) {
	return system.GetProjectAgentURL(project)
}
