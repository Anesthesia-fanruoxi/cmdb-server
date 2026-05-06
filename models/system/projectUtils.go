package system

import (
	"cmdb/pkg/database"
	"fmt"
)

// GetProjectAgentURL 根据项目标识获取Agent代理地址
func GetProjectAgentURL(project string) (string, error) {
	var projectDict ProjectDict
	if err := database.DB.Where("project = ?", project).First(&projectDict).Error; err != nil {
		return "", fmt.Errorf("获取项目信息失败: %v", err)
	}

	if projectDict.AgentURL == "" {
		return "", fmt.Errorf("项目 %s 未配置agent地址", project)
	}

	return projectDict.AgentURL, nil
}

// GetProjectInfo 获取完整的项目信息
func GetProjectInfo(project string) (*ProjectDict, error) {
	var projectDict ProjectDict
	if err := database.DB.Where("project = ?", project).First(&projectDict).Error; err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %v", err)
	}
	return &projectDict, nil
}
