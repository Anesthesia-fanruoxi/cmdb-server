package task

import (
	"cmdb/models/task"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"time"
)

// CreateTaskRecord 创建任务记录到MySQL
func CreateTaskRecord(taskID, taskType, userName string, userID int64, title string) error {
	record := &task.TaskCenter{
		TaskID:   taskID,
		Type:     taskType,
		Status:   task.StatusPending,
		UserID:   userID,
		UserName: userName,
		Title:    title,
	}

	if err := database.DB.Create(record).Error; err != nil {
		middleware.Logger.Errorf("创建任务记录失败: %v", err)
		return err
	}

	middleware.Logger.Infof("创建任务记录成功: taskID=%s, type=%s", taskID, taskType)
	return nil
}

// UpdateTaskSuccess 更新任务为成功，同时更新文件信息
func UpdateTaskSuccess(taskID, filePath, fileName string, fileSize int64, totalCount int) error {
	now := time.Now()
	expireAt := now.AddDate(0, 0, 30) // 30天后过期

	updates := map[string]interface{}{
		"status":       task.StatusSuccess,
		"file_path":    filePath,
		"file_name":    fileName,
		"file_size":    fileSize,
		"total_count":  totalCount,
		"completed_at": now,
		"expire_at":    expireAt,
	}

	if err := database.DB.Model(&task.TaskCenter{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error; err != nil {
		middleware.Logger.Errorf("更新任务成功状态失败: %v", err)
		return err
	}

	middleware.Logger.Infof("任务完成: taskID=%s, file=%s, count=%d", taskID, fileName, totalCount)
	return nil
}

// UpdateTaskFailed 更新任务为失败
func UpdateTaskFailed(taskID, errorMessage string) error {
	now := time.Now()

	updates := map[string]interface{}{
		"status":        task.StatusFailed,
		"error_message": errorMessage,
		"completed_at":  now,
	}

	if err := database.DB.Model(&task.TaskCenter{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error; err != nil {
		middleware.Logger.Errorf("更新任务失败状态失败: %v", err)
		return err
	}

	return nil
}

// GetTaskByTaskID 根据任务ID获取任务记录
func GetTaskByTaskID(taskID string) (*task.TaskCenter, error) {
	var record task.TaskCenter
	if err := database.DB.Where("task_id = ?", taskID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetTaskByTaskIDAndUser 根据任务ID和用户ID获取任务记录
func GetTaskByTaskIDAndUser(taskID string, userID int64) (*task.TaskCenter, error) {
	var record task.TaskCenter
	if err := database.DB.Where("task_id = ? AND user_id = ?", taskID, userID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetTaskList 获取用户任务列表（只返回最近30天）
// 如果 userID 为 0，表示管理员查看所有任务
// keyword 支持模糊查询 username 和 nickname
func GetTaskList(userID int64, taskType, status, keyword string) ([]task.TaskCenter, error) {
	// 定义查询结果结构体（包含昵称）
	type TaskCenterWithNick struct {
		task.TaskCenter
		NickName string `gorm:"column:nick_name"`
	}

	var records []task.TaskCenter

	// 管理员查看时，使用 JOIN 查询获取昵称
	if userID == 0 {
		var withNick []TaskCenterWithNick

		query := database.DB.Table(middleware.TableTaskCenter + " as tc").
			Select("tc.*, u.nick_name").
			Joins("LEFT JOIN " + middleware.TableUsers + " as u ON tc.user_id = u.id")

		// 只查询可展示的任务类型
		query = query.Where("tc.type IN ?", []string{task.TypeAnalysis, task.TypeESExport, task.TypeSQLExport})

		// 只查询最近30天的任务（因为Redis详情保留30天）
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		query = query.Where("tc.created_at >= ?", thirtyDaysAgo)

		if taskType != "" {
			query = query.Where("tc.type = ?", taskType)
		}
		if status != "" {
			query = query.Where("tc.status = ?", status)
		}

		// 模糊查询：支持 username 和 nickname
		if keyword != "" {
			fuzzyPattern := buildFuzzyPattern(keyword)
			query = query.Where("tc.user_name LIKE ? OR u.nick_name LIKE ?", fuzzyPattern, fuzzyPattern)
		}

		// 按创建时间倒序
		if err := query.Order("tc.created_at DESC").Scan(&withNick).Error; err != nil {
			middleware.Logger.Errorf("查询任务列表失败: %v", err)
			return nil, err
		}

		// 转换并填充昵称
		records = make([]task.TaskCenter, len(withNick))
		for i, item := range withNick {
			records[i] = item.TaskCenter
			records[i].NickName = item.NickName
		}

		return records, nil
	}

	// 普通用户查询
	query := database.DB.Model(&task.TaskCenter{})

	query = query.Where("user_id = ?", userID)

	// 只查询可展示的任务类型
	query = query.Where("type IN ?", []string{task.TypeAnalysis, task.TypeESExport, task.TypeSQLExport})

	// 只查询最近30天的任务
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	query = query.Where("created_at >= ?", thirtyDaysAgo)

	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 模糊查询 username
	if keyword != "" {
		fuzzyPattern := buildFuzzyPattern(keyword)
		query = query.Where("user_name LIKE ?", fuzzyPattern)
	}

	// 按创建时间倒序
	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		middleware.Logger.Errorf("查询任务列表失败: %v", err)
		return nil, err
	}

	return records, nil
}

// buildFuzzyPattern 构建模糊匹配模式
// 例如：cs -> %c%s%, 测试 -> %测%试%
func buildFuzzyPattern(keyword string) string {
	if keyword == "" {
		return ""
	}

	runes := []rune(keyword)
	pattern := "%"
	for _, r := range runes {
		pattern += string(r) + "%"
	}
	return pattern
}

// isDisplayableType 判断是否需要在任务中心展示的类型
func isDisplayableType(taskType string) bool {
	displayTypes := map[string]bool{
		task.TypeAnalysis:  true,
		task.TypeESExport:  true,
		task.TypeSQLExport: true,
	}
	return displayTypes[taskType]
}

// getTaskTitle 根据任务类型生成标题
func getTaskTitle(taskType string, _ []byte) string {
	return task.GetTypeText(taskType)
}
