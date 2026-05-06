package proUpdate

import (
	"cmdb/config"
	"cmdb/models/assets"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProUpdateRecords 接收第三方推送的发版记录数据
// POST /api/assets/proUpdate/records
func ProUpdateRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}
	// 读取原始请求数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求数据失败")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)
	//middleware.Logger.Infof("body: %s", string(body))
	// 解析请求数据
	var requestData assets.RecordRequest
	if err := json.Unmarshal(body, &requestData); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据格式错误")
		return
	}
	//middleware.Logger.Infof("开始接受到数据：%s", requestData.Data)

	// 参数验证
	if requestData.Data == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	// 获取任务加密盐
	cfg := config.GetConfig()
	CICDSalt := cfg.CICD.Secret
	if CICDSalt == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "任务加密盐未配置")
		return
	}

	// 解密并解压数据
	decompressedData, err := middleware.DecryptAndDecompress(requestData.Data, CICDSalt)
	if err != nil {
		middleware.Logger.Errorf("解密解压失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "解密解压失败: "+err.Error())
		return
	}

	// 解密成功后再打印数据
	middleware.Logger.Infof("解密后的数据为: %s", decompressedData)

	// 解析解密后的数据
	var recordReq assets.ReleaseRecordRequest
	if err := json.Unmarshal(decompressedData, &recordReq); err != nil {
		middleware.Logger.Errorf("解析解密后的数据失败: %v，数据内容: %s", err, decompressedData)
		middleware.ResponseError(w, http.StatusBadRequest, "解析解密后的数据失败: "+err.Error())
		return
	}

	// 获取任务ID
	id := recordReq.ID
	if id == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 确保ID长度不超过32个字符
	if len(id) > 32 {
		middleware.Logger.Warnf("任务ID过长，原始ID: %s", id)
		id = id[:32]
		middleware.Logger.Warnf("任务ID已截断为32个字符: %s", id)
	}

	ctx := context.Background()

	if !recordReq.IsSet {
		// 处理任务信息（isset=false）
		middleware.Logger.Infof("处理任务信息，ID: %s, Status: %s", id, recordReq.Status)

		// 验证必要字段
		if recordReq.Name == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "项目名称不能为空")
			return
		}

		// 处理任务信息
		if err := handleTaskInfo(ctx, id, &recordReq); err != nil {
			middleware.Logger.Errorf("处理任务信息失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
			return
		}

	} else {
		// 处理步骤信息（isset=true）
		middleware.Logger.Infof("处理步骤信息，ID: %s, Step: %d, StepName: %s", id, recordReq.Step, recordReq.StepName)

		// 处理步骤信息
		if err := handleStepInfo(ctx, id, &recordReq); err != nil {
			middleware.Logger.Errorf("处理步骤信息失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// 返回成功响应
	middleware.ResponseSuccess(w, "接收发版记录成功", nil)
}

// 处理任务信息
func handleTaskInfo(ctx context.Context, id string, recordReq *assets.ReleaseRecordRequest) error {
	// 处理开始时间
	var startedAt time.Time
	if recordReq.StartedAt != "" {
		var err error
		startedAt, err = time.ParseInLocation("2006-01-02 15:04:05", recordReq.StartedAt, time.Local)
		if err != nil {
			middleware.Logger.Errorf("解析开始时间失败: %v, 时间字符串: %s", err, recordReq.StartedAt)
			return err
		}
	} else {
		startedAt = time.Now()
	}

	// 处理结束时间（可为空）
	var completedAt time.Time
	if recordReq.FinishedAt != "" {
		var err error
		completedAt, err = time.ParseInLocation("2006-01-02 15:04:05", recordReq.FinishedAt, time.Local)
		if err != nil {
			middleware.Logger.Warnf("解析结束时间失败: %v, 时间字符串: %s，使用零值", err, recordReq.FinishedAt)
			completedAt = time.Time{}
		}
	}

	// 处理Redis任务记录
	if err := updateRedisTaskRecord(ctx, id, recordReq); err != nil {
		return err
	}

	// MySQL操作：只有isset=false时才操作数据库
	// 确定MySQL状态
	var mysqlStatus string
	if recordReq.FinishedAt != "" {
		// 任务完成，状态与Redis保持一致
		if recordReq.Status == "complete" {
			mysqlStatus = "success"
		} else if recordReq.Status == "failed" {
			mysqlStatus = "failed"
		} else if recordReq.Status == "cancel" {
			mysqlStatus = "cancel"
		} else {
			mysqlStatus = "failed" // 默认为失败
		}
	} else {
		mysqlStatus = "running" // MySQL中任务运行状态（包括排队中的任务）
	}

	var existingRecord assets.ReleaseRecord
	result := database.DB.Where("task_id = ?", id).First(&existingRecord)

	if result.Error != nil && result.Error.Error() == "record not found" {
		// 创建新记录
		releaseRecord := assets.ReleaseRecord{
			TaskID:       id,
			Project:      recordReq.Name,
			Type:         recordReq.Type,
			Category:     recordReq.Category,
			Status:       mysqlStatus,
			StartedAt:    startedAt,
			CompletedAt:  completedAt,
			GitURL:       recordReq.GitURL,
			Description:  recordReq.Description,
			UpdateFeishu: recordReq.UpdateFeishu,
			NotifyFeishu: recordReq.NotifyFeishu,
		}

		middleware.Logger.Infof("创建MySQL任务记录，ID: %s, Status: %s", id, mysqlStatus)
		if err := database.DB.Create(&releaseRecord).Error; err != nil {
			middleware.Logger.Errorf("创建MySQL记录失败: %v", err)
			return err
		}
	} else if result.Error == nil {
		// 记录已存在，更新记录（为空不覆盖）
		updateFields := map[string]interface{}{
			"status": mysqlStatus,
		}
		if recordReq.Name != "" {
			updateFields["project"] = recordReq.Name
		}
		if recordReq.Type != "" {
			updateFields["type"] = recordReq.Type
		}
		if recordReq.Category != "" {
			updateFields["category"] = recordReq.Category
		}
		if startedAt != (time.Time{}) {
			updateFields["started_at"] = startedAt
		}
		if recordReq.GitURL != "" {
			updateFields["git_url"] = recordReq.GitURL
		}
		if recordReq.Description != "" {
			updateFields["description"] = recordReq.Description
		}
		if recordReq.UpdateFeishu != "" {
			updateFields["update_feishu"] = recordReq.UpdateFeishu
		}
		if recordReq.NotifyFeishu != "" {
			updateFields["notify_feishu"] = recordReq.NotifyFeishu
		}
		if !completedAt.IsZero() {
			updateFields["completed_at"] = completedAt
		}

		middleware.Logger.Infof("更新MySQL任务记录，ID: %s, Status: %s", id, mysqlStatus)
		if err := database.DB.Model(&existingRecord).Updates(updateFields).Error; err != nil {
			middleware.Logger.Errorf("更新MySQL记录失败: %v", err)
			return err
		}
	}

	return nil
}

// 处理步骤信息
func handleStepInfo(ctx context.Context, id string, recordReq *assets.ReleaseRecordRequest) error {
	// 更新Redis任务记录，添加步骤信息
	return updateRedisTaskRecordWithStep(ctx, id, recordReq)
}

// 更新Redis任务记录（任务信息）
func updateRedisTaskRecord(ctx context.Context, id string, recordReq *assets.ReleaseRecordRequest) error {
	// 尝试获取现有记录
	existingJSON, err := GetTaskRecord(ctx, id)
	var taskRecord assets.RedisTaskRecord

	if err != nil && err.Error() != "redis: nil" {
		return err
	}

	if existingJSON != "" {
		// 解析现有记录（数组格式）
		var taskRecords []assets.RedisTaskRecord
		if err := json.Unmarshal([]byte(existingJSON), &taskRecords); err != nil {
			middleware.Logger.Errorf("解析现有Redis记录失败: %v", err)
			return err
		}
		if len(taskRecords) > 0 {
			taskRecord = taskRecords[0]
		}
		middleware.Logger.Infof("更新现有Redis任务记录，ID: %s", id)
	} else {
		// 创建新记录
		middleware.Logger.Infof("创建新的Redis任务记录，ID: %s", id)
		taskRecord = assets.RedisTaskRecord{
			StepDetail: []assets.RedisStepInfo{}, // 初始化空的步骤数组
		}
	}

	// 更新任务基本信息（仅当新值不为空时才更新，避免覆盖已有数据）
	if recordReq.Name != "" {
		taskRecord.ProjectName = recordReq.Name
	}
	if recordReq.Description != "" {
		taskRecord.Description = recordReq.Description
	}
	if recordReq.GitURL != "" {
		taskRecord.GitURL = recordReq.GitURL
	}
	if recordReq.UpdateFeishu != "" {
		taskRecord.UpdateFeishu = recordReq.UpdateFeishu
	}
	if recordReq.NotifyFeishu != "" {
		taskRecord.NotifyFeishu = recordReq.NotifyFeishu
	}
	if recordReq.StartedAt != "" {
		taskRecord.StartedAt = recordReq.StartedAt
	}
	if recordReq.Type != "" {
		taskRecord.Type = recordReq.Type
	}
	if recordReq.Category != "" {
		taskRecord.Category = recordReq.Category
	}
	taskRecord.Timestamp = fmt.Sprintf("%d", time.Now().UnixMilli())

	// 设置Redis状态
	if recordReq.FinishedAt != "" {
		// 任务完成，状态与MySQL保持一致
		taskRecord.FinishedAt = recordReq.FinishedAt
		if recordReq.Status == "complete" {
			taskRecord.Status = "success"
		} else if recordReq.Status == "failed" {
			taskRecord.Status = "failed"
		} else if recordReq.Status == "cancel" {
			taskRecord.Status = "cancel"
		} else {
			taskRecord.Status = "failed" // 默认为失败
		}
	} else if len(taskRecord.StepDetail) > 0 {
		// 有步骤信息，状态为running
		taskRecord.Status = "running"
	} else {
		// 无步骤信息，状态为pending
		taskRecord.Status = "pending"
		taskRecord.Step = 0
	}

	// 序列化并保存
	recordJSON, err := json.MarshalIndent([]assets.RedisTaskRecord{taskRecord}, "", "  ")
	if err != nil {
		middleware.Logger.Errorf("序列化Redis任务记录失败: %v", err)
		return err
	}

	if err := SetTaskRecord(ctx, id, string(recordJSON)); err != nil {
		middleware.Logger.Errorf("保存Redis任务记录失败: %v", err)
		return err
	}

	middleware.Logger.Infof("Redis任务记录已保存，ID: %s, Status: %s", id, taskRecord.Status)

	return nil
}

// 更新Redis任务记录（添加步骤信息）
func updateRedisTaskRecordWithStep(ctx context.Context, id string, recordReq *assets.ReleaseRecordRequest) error {
	// 获取现有记录
	existingJSON, err := GetTaskRecord(ctx, id)
	if err != nil {
		middleware.Logger.Errorf("获取现有Redis记录失败: %v", err)
		return err
	}

	var taskRecords []assets.RedisTaskRecord
	if err := json.Unmarshal([]byte(existingJSON), &taskRecords); err != nil {
		middleware.Logger.Errorf("解析现有Redis记录失败: %v", err)
		return err
	}

	if len(taskRecords) == 0 {
		return fmt.Errorf("未找到任务记录: %s", id)
	}

	taskRecord := &taskRecords[0] // 只有一条记录

	// 查找是否已存在相同步骤
	existingStepIndex := -1
	for i, step := range taskRecord.StepDetail {
		if step.Step == recordReq.Step && step.StepType == recordReq.StepType {
			existingStepIndex = i
			break
		}
	}

	if existingStepIndex >= 0 {
		// 更新现有步骤（完成通知）
		taskRecord.StepDetail[existingStepIndex].StepFinishedAt = recordReq.StepFinishedAt
		taskRecord.StepDetail[existingStepIndex].StepStatus = recordReq.StepStatus
		taskRecord.StepDetail[existingStepIndex].Duration = recordReq.Duration
		taskRecord.StepDetail[existingStepIndex].EstimatedEnd = recordReq.EstimatedEnd
		taskRecord.StepDetail[existingStepIndex].LastDuration = recordReq.LastDuration
		middleware.Logger.Infof("更新现有步骤，ID: %s, Step: %d, StepType: %s, Status: %s, EstimatedEnd: %s, LastDuration: %.2f",
			id, recordReq.Step, recordReq.StepType, recordReq.StepStatus, recordReq.EstimatedEnd, recordReq.LastDuration)
	} else {
		// 创建新步骤（开始通知）
		newStep := assets.RedisStepInfo{
			Step:           recordReq.Step,
			StepType:       recordReq.StepType,
			StepStartedAt:  recordReq.StepStartedAt,
			StepFinishedAt: recordReq.StepFinishedAt,
			StepName:       recordReq.StepName,
			StepStatus:     recordReq.StepStatus,
			Remote:         recordReq.Remote,
			Duration:       recordReq.Duration,
			EstimatedEnd:   recordReq.EstimatedEnd,
			LastDuration:   recordReq.LastDuration,
		}
		// 追加新步骤到数组
		taskRecord.StepDetail = append(taskRecord.StepDetail, newStep)
		middleware.Logger.Infof("创建新步骤，ID: %s, Step: %d, StepType: %s, Status: %s, EstimatedEnd: %s, LastDuration: %.2f",
			id, recordReq.Step, recordReq.StepType, recordReq.StepStatus, recordReq.EstimatedEnd, recordReq.LastDuration)
	}

	// 更新任务当前步骤和状态
	taskRecord.Step = recordReq.Step
	if taskRecord.Status == "pending" {
		taskRecord.Status = "running" // 从pending转为running
	}

	// 序列化并保存
	recordJSON, err := json.MarshalIndent(taskRecords, "", "  ")
	if err != nil {
		middleware.Logger.Errorf("序列化Redis任务记录失败: %v", err)
		return err
	}

	if err := SetTaskRecord(ctx, id, string(recordJSON)); err != nil {
		middleware.Logger.Errorf("保存Redis任务记录失败: %v", err)
		return err
	}

	middleware.Logger.Infof("Redis任务记录已更新步骤，ID: %s, Step: %d, Status: %s", id, recordReq.Step, taskRecord.Status)

	// 第一步开始时发送任务开始通知（带手机端链接，只发一次）
	if taskRecord.NotifyFeishu != "" && recordReq.Step == 1 && recordReq.StepStatus == "running" {
		if err := SendTaskStartNotify(id, taskRecord, taskRecord.NotifyFeishu); err != nil {
			middleware.Logger.Errorf("发送飞书开始通知失败: %v", err)
		}
	}

	return nil
}

// 任务记录相关操作
const (
	// TaskRecordExpireTime 任务记录在 Redis 中的过期时间（30天）
	TaskRecordExpireTime = 30 * 24 * time.Hour
)

// SetTaskRecord 设置任务记录到 Redis
func SetTaskRecord(ctx context.Context, taskID string, recordJSON string) error {
	redisKey := middleware.GetTaskRecordKey(taskID)
	return database.RDB.Set(ctx, redisKey, recordJSON, TaskRecordExpireTime).Err()
}

// GetTaskRecord 从 Redis 获取任务记录
func GetTaskRecord(ctx context.Context, taskID string) (string, error) {
	redisKey := middleware.GetTaskRecordKey(taskID)
	return database.RDB.Get(ctx, redisKey).Result()
}
