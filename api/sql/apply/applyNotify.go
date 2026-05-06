package apply

import (
	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/pkg/notice"
	"fmt"
)

// SendSQLApplyNotificationCard 发送SQL执行审批卡片通知
func SendSQLApplyNotificationCard(apply *models.SQLApply, role string) error {
	var targetUserID int64

	// 根据角色确定目标用户
	switch role {
	case "apply":
		targetUserID = apply.ApplyID
	case "executor":
		targetUserID = apply.ExecutorID
	default:
		return fmt.Errorf("不支持的角色: %s", role)
	}

	// 获取目标用户邮箱
	targetEmail := getUserEmail(targetUserID)
	if targetEmail == "" {
		return fmt.Errorf("无法获取用户 %d 的邮箱", targetUserID)
	}

	// 获取项目中文名称
	projectName := getProjectName(apply.Project)
	if projectName == "" {
		projectName = apply.Project
	}

	// 格式化执行时间
	var executionTimeStr string
	if apply.ExecutionTime != nil {
		executionTimeStr = middleware.FormatTimeToChina(*apply.ExecutionTime)
	}

	// 构建SQL执行审批卡片信息
	cardInfo := notice.SQLApplyCardInfo{
		ApplyID:       apply.ID,
		SubmitterName: apply.SubmitterName,
		ProjectName:   projectName,
		DatabaseName:  apply.DatabaseName,
		Description:   apply.Description,
		ExecutionTime: executionTimeStr,
		CreatedAt:     middleware.FormatTimeToChina(apply.CreatedAt),
		Role:          role,
	}

	// 发送私信卡片
	return notice.SendSQLApplyCard(targetEmail, cardInfo)
}

// getUserEmail 根据用户ID获取用户邮箱
func getUserEmail(userID int64) string {
	if userID <= 0 {
		return ""
	}

	var email string
	err := database.DB.Table(middleware.TableUsers).
		Select("email").
		Where("id = ?", userID).
		Row().
		Scan(&email)

	if err != nil {
		middleware.Logger.Errorf("获取用户 %d 邮箱失败: %v", userID, err)
		return ""
	}

	return email
}

// getProjectName 根据项目标识获取项目中文名称
func getProjectName(projectID string) string {
	if projectID == "" {
		return ""
	}

	var projectName string
	err := database.DB.Table(middleware.TableProjectDict).
		Select("project_name").
		Where("project = ?", projectID).
		Row().
		Scan(&projectName)

	if err != nil {
		middleware.Logger.Errorf("获取项目 %s 中文名称失败: %v", projectID, err)
		return ""
	}

	return projectName
}
