package apply

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"

	"gorm.io/gorm"
)

// getUserIDByFeiShuID 通过飞书ID获取系统用户ID
// 方案：通过飞书API获取用户邮箱，再通过邮箱匹配系统用户
func getUserIDByFeiShuID(userID, openID string) (int64, error) {
	// 方案1：通过邮箱映射表反向查找
	// emailToUserIDMap 在 sendFeiShuCard.go 中定义，这里需要反向查找
	for email, fsUserID := range emailToFeiShuUserIDMap {
		if fsUserID == userID || fsUserID == openID {
			// 通过邮箱查找系统用户ID
			var sysUserID int64
			err := database.DB.Table(middleware.TableUsers).
				Select("id").
				Where("email = ?", email).
				Row().Scan(&sysUserID)
			if err == nil && sysUserID > 0 {
				return sysUserID, nil
			}
		}
	}

	// 方案2：直接通过user_id在映射表中查找对应的邮箱
	email := getFeiShuUserEmail(userID)
	if email != "" {
		var sysUserID int64
		err := database.DB.Table(middleware.TableUsers).
			Select("id").
			Where("email = ?", email).
			Row().Scan(&sysUserID)
		if err == nil && sysUserID > 0 {
			return sysUserID, nil
		}
	}

	return 0, errors.New("未找到对应的系统用户，请确认飞书账号已关联")
}

// emailToFeiShuUserIDMap 邮箱到飞书用户ID的映射（与sendFeiShuCard.go中的映射保持一致）
var emailToFeiShuUserIDMap = map[string]string{
	"zhanghuajia@hzbxhd.com": "1gcd7cd8",
	"zhaowenda@hzbxhd.com":   "7a8d37gb",
	"guokai@hzbxhd.com":      "72bf1bg3",
	"wangxin@hzbxhd.com":     "3bdade88",
	"huxiaojun@hzbxhd.com":   "6696f6fd",
	"liyu@hzbxhd.com":        "47b555ad",
	"liyang@hzbxhd.com":      "5b5eb6bg",
	"liurong@hzbxhd.com":     "f6d2773d",
	"jiangqi@hzbxhd.com":     "bbbe9c64",
}

// getFeiShuUserEmail 通过飞书用户ID获取邮箱
func getFeiShuUserEmail(feiShuUserID string) string {
	for email, userID := range emailToFeiShuUserIDMap {
		if userID == feiShuUserID {
			return email
		}
	}
	return ""
}

// validateCallbackPermission 验证回调操作权限
func validateCallbackPermission(apply *models.SQLApply, operatorID int64, action string) error {
	switch action {
	case "approve", "reject":
		// 审批操作 - 验证是否是审批人且状态允许
		if apply.ApplyID != operatorID {
			return errors.New("您不是该流程的审批人")
		}
		if apply.Status != models.ApplyStatusPending || apply.ApplyStatus != models.ApplyRoleCanProcess {
			return errors.New("当前流程状态不允许审批操作")
		}

	case "execute":
		// 执行操作 - 验证是否是执行人且状态允许
		if apply.ExecutorID != operatorID {
			return errors.New("您不是该流程的执行人")
		}
		if apply.Status != models.ApplyStatusApproved || apply.ExecutorStatus != models.ApplyRoleCanProcess {
			return errors.New("当前流程状态不允许执行操作")
		}

	default:
		return errors.New("不支持的操作类型")
	}

	return nil
}

// executeApplyAction 执行审批操作并返回更新后的卡片
func executeApplyAction(apply *models.SQLApply, action string) (map[string]interface{}, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var err error
	currentTime := time.Now()

	switch action {
	case "approve":
		err = executeApprove(tx, apply, currentTime)
	case "reject":
		err = executeReject(tx, apply, currentTime)
	case "execute":
		err = executeSQL(tx, apply, currentTime)
	default:
		tx.Rollback()
		return nil, errors.New("不支持的操作类型")
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事务失败")
	}

	// 返回更新后的卡片
	return createResultCard(apply, action), nil
}

// executeApprove 执行审批通过
func executeApprove(tx *gorm.DB, apply *models.SQLApply, currentTime time.Time) error {
	submitterCanProcess := models.ApplyRoleCannotProcess
	if apply.SubmitterID == apply.ApplyID {
		submitterCanProcess = models.ApplyRoleCanProcess
	}

	updateFields := map[string]interface{}{
		"apply_time":       currentTime,
		"apply_status":     models.ApplyRoleCannotProcess,
		"apply_source":     "feishu",
		"apply_comment":    "通过（飞书快捷审批）",
		"submitter_status": submitterCanProcess,
		"executor_status":  models.ApplyRoleCanProcess,
		"status":           models.ApplyStatusApproved,
	}

	if err := tx.Model(&models.SQLApply{}).Where("id = ?", apply.ID).Updates(updateFields).Error; err != nil {
		return err
	}

	// 更新内存中的apply对象
	apply.Status = models.ApplyStatusApproved
	apply.ApplyStatus = models.ApplyRoleCannotProcess
	apply.ExecutorStatus = models.ApplyRoleCanProcess

	// 重新获取更新后的记录
	tx.Where("id = ?", apply.ID).First(apply)

	// 异步通知执行人
	go func() {
		if err := SendSQLApplyNotificationCard(apply, "executor"); err != nil {
			middleware.Logger.Errorf("发送执行人通知失败: %v", err)
		}
	}()

	return nil
}

// executeReject 执行审批拒绝
func executeReject(tx *gorm.DB, apply *models.SQLApply, currentTime time.Time) error {
	updateFields := map[string]interface{}{
		"apply_time":       currentTime,
		"apply_status":     models.ApplyRoleCannotProcess,
		"apply_source":     "feishu",
		"apply_comment":    "拒绝（飞书快捷审批）",
		"submitter_status": models.ApplyRoleCannotProcess,
		"executor_status":  models.ApplyRoleCannotProcess,
		"status":           models.ApplyStatusRejected,
	}

	if err := tx.Model(&models.SQLApply{}).Where("id = ?", apply.ID).Updates(updateFields).Error; err != nil {
		return err
	}

	apply.Status = models.ApplyStatusRejected

	// 重新获取更新后的记录
	tx.Where("id = ?", apply.ID).First(apply)

	return nil
}

// executeSQL 执行SQL
func executeSQL(tx *gorm.DB, apply *models.SQLApply, _ time.Time) error {
	updateFields := map[string]interface{}{
		"executor_status":  models.ApplyRoleCannotProcess,
		"submitter_status": models.ApplyRoleCannotProcess,
		"apply_status":     models.ApplyRoleCannotProcess,
		"executor_source":  "feishu",
		"status":           models.ApplyStatusExecuting,
	}

	if err := tx.Model(&models.SQLApply{}).Where("id = ?", apply.ID).Updates(updateFields).Error; err != nil {
		return err
	}

	apply.Status = models.ApplyStatusExecuting

	// 重新获取更新后的记录
	tx.Where("id = ?", apply.ID).First(apply)

	// 判断是否有定时执行
	if apply.ExecutionTime != nil && apply.ExecutionTime.After(time.Now()) {
		middleware.Logger.Infof("SQL设置了定时执行: %v", apply.ExecutionTime)
		return nil
	}

	// 异步执行SQL
	go executeApplySQL(apply.ID, apply.SQLContent, apply.DatabaseName, apply.Project)

	return nil
}

// respondWithCard 响应卡片内容
func respondWithCard(w http.ResponseWriter, card map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// createErrorCard 创建错误提示卡片
func createErrorCard(title, message string) map[string]interface{} {
	return map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"template": "red",
			"title": map[string]interface{}{
				"content": fmt.Sprintf("❌ %s", title),
				"tag":     "plain_text",
			},
		},
		"elements": []map[string]interface{}{
			{
				"tag": "div",
				"text": map[string]interface{}{
					"content": message,
					"tag":     "plain_text",
				},
			},
		},
	}
}

// createResultCard 创建操作结果卡片
func createResultCard(apply *models.SQLApply, action string) map[string]interface{} {
	var headerColor, title, statusText string

	switch action {
	case "approve":
		headerColor = "green"
		title = "✅ 审批已通过"
		statusText = fmt.Sprintf("已通过，等待执行人 %s 执行", apply.ExecutorName)
	case "reject":
		headerColor = "red"
		title = "❌ 审批已拒绝"
		statusText = "已拒绝"
	case "execute":
		headerColor = "blue"
		title = "⚡ 开始执行"
		statusText = "SQL正在执行中..."
	default:
		headerColor = "grey"
		title = "操作完成"
		statusText = "操作已完成"
	}

	projectName := getProjectName(apply.Project)
	if projectName == "" {
		projectName = apply.Project
	}

	return map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"template": headerColor,
			"title": map[string]interface{}{
				"content": title,
				"tag":     "plain_text",
			},
		},
		"elements": []map[string]interface{}{
			{
				"tag": "div",
				"fields": []map[string]interface{}{
					{
						"is_short": true,
						"text": map[string]interface{}{
							"content": fmt.Sprintf("**审批ID：** %d", apply.ID),
							"tag":     "lark_md",
						},
					},
					{
						"is_short": true,
						"text": map[string]interface{}{
							"content": fmt.Sprintf("**项目：** %s", projectName),
							"tag":     "lark_md",
						},
					},
				},
			},
			{
				"tag": "hr",
			},
			{
				"tag": "note",
				"elements": []map[string]interface{}{
					{
						"content": statusText,
						"tag":     "plain_text",
					},
				},
			},
		},
	}
}
