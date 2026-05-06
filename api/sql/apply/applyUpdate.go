package apply

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"

	"gorm.io/gorm"
)

// SqlApplyUpdateNew 新的SQL审批更新接口
func SqlApplyUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受PUT请求
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持PUT请求")
		return
	}

	// 从上下文中获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "无法获取用户信息")
		return
	}

	// 解析请求体
	var req models.UpdateApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 验证流程ID
	if req.ID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID不能为空")
		return
	}

	// 获取当前流程信息
	var apply models.SQLApply
	if err := database.DB.Where("id = ?", req.ID).First(&apply).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询流程失败")
		}
		return
	}

	// 开启数据库事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 判断当前用户角色
	// 优先级：提交人 > 审批人 > 执行人
	// 当提交人和审批人是同一人时，优先识别为提交人（可以撤回）
	userRole := ""
	if apply.SubmitterID == userInfo.UserID {
		userRole = "submit"
	} else if apply.ApplyID == userInfo.UserID {
		userRole = "apply"
	} else if apply.ExecutorID == userInfo.UserID {
		userRole = "executor"
	}

	// 如果找不到有效角色，直接返回错误
	if userRole == "" {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, "您不是该流程的相关人员")
		return
	}

	// 检查权限
	if err := CheckApplyPermission(req.ID, userInfo.UserID, userRole); err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, err.Error())
		return
	}

	// 根据角色分发处理
	var err error
	switch userRole {
	case "submit":
		err = submitterApplyOperator(tx, &req, &apply, w)
	case "apply":
		err = approverApplyOperator(tx, &req, &apply, w, "web")
	case "executor":
		err = executorApplyOperator(tx, &req, &apply, w, "web")
	}

	if err != nil {
		tx.Rollback()
		return // 错误已经在各处理函数中响应给客户端
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "提交事务失败")
		return
	}

	middleware.ResponseSuccess(w, "操作成功", nil)
}

// CheckApplyPermission 检查审批权限
func CheckApplyPermission(applyID int64, _ int64, role string) error {
	var apply models.SQLApply
	if err := database.DB.Where("id = ?", applyID).First(&apply).Error; err != nil {
		return errors.New("查询流程失败")
	}

	// 根据角色检查状态
	switch role {
	case "submit":
		if apply.SubmitterStatus != models.ApplyRoleCanProcess {
			return errors.New("提交人当前不能处理该流程")
		}
	case "apply":
		if apply.ApplyStatus != models.ApplyRoleCanProcess {
			return errors.New("审批人当前不能处理该流程")
		}
	case "executor":
		if apply.ExecutorStatus != models.ApplyRoleCanProcess {
			return errors.New("执行人当前不能处理该流程")
		}
	}

	return nil
}

// submitterApplyOperator 处理提交人的操作
func submitterApplyOperator(tx *gorm.DB, req *models.UpdateApplyRequest, apply *models.SQLApply, w http.ResponseWriter) error {
	middleware.Logger.Infof("提交人操作参数: ID=%d, ProcessType=%d", req.ID, req.ProcessType)

	// 提交人只能撤销流程（ProcessType=0表示撤销）
	if req.ProcessType != 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "提交人只能撤销流程")
		return errors.New("提交人只能撤销流程")
	}

	// 撤销操作：关闭所有人的处理权限
	if err := tx.Model(&models.SQLApply{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":           models.ApplyStatusCanceled,
		"submitter_status": models.ApplyRoleCannotProcess,
		"apply_status":     models.ApplyRoleCannotProcess,
		"executor_status":  models.ApplyRoleCannotProcess,
	}).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "撤销操作失败: "+err.Error())
		return err
	}

	// 获取更新后的记录
	if err := tx.Where("id = ?", req.ID).First(apply).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	middleware.Logger.Infof("提交人 %s 撤销了审批流程 ID: %d", apply.SubmitterName, apply.ID)
	return nil
}

// approverApplyOperator 处理审批人的操作
func approverApplyOperator(tx *gorm.DB, req *models.UpdateApplyRequest, apply *models.SQLApply, w http.ResponseWriter, source string) error {
	var statusValue int
	var executorCanProcess int

	currentTime := time.Now()

	// 打印接收到的参数，用于调试
	middleware.Logger.Infof("审批人操作参数: ID=%d, ProcessType=%d, Comment=%s", req.ID, req.ProcessType, req.Comment)

	if req.ProcessType == 1 {
		// 审批通过
		statusValue = models.ApplyStatusApproved
		executorCanProcess = models.ApplyRoleCanProcess
		middleware.Logger.Infof("审批操作: 通过")
	} else {
		// 审批拒绝
		statusValue = models.ApplyStatusRejected
		executorCanProcess = models.ApplyRoleCannotProcess
		middleware.Logger.Infof("审批操作: 拒绝 (ProcessType=%d)", req.ProcessType)
	}

	// 准备更新字段
	// 如果提交人和审批人是同一人，审批通过后提交人仍可撤回
	submitterCanProcess := models.ApplyRoleCannotProcess
	if req.ProcessType == 1 && apply.SubmitterID == apply.ApplyID {
		submitterCanProcess = models.ApplyRoleCanProcess
	}

	updateFields := map[string]interface{}{
		"apply_time":       currentTime,
		"apply_status":     models.ApplyRoleCannotProcess,
		"apply_source":     source,
		"apply_comment":    req.Comment,
		"submitter_status": submitterCanProcess,
		"executor_status":  executorCanProcess,
		"status":           statusValue,
	}

	if err := tx.Model(&models.SQLApply{}).Where("id = ?", req.ID).Updates(updateFields).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "审批操作失败: "+err.Error())
		return err
	}

	// 获取更新后的记录
	if err := tx.Where("id = ?", req.ID).First(apply).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	// 审批通过后，发送飞书通知给执行人
	if req.ProcessType == 1 {
		go func() {
			if err := SendSQLApplyNotificationCard(apply, "executor"); err != nil {
				middleware.Logger.Errorf("发送执行人飞书通知失败: %v", err)
			} else {
				middleware.Logger.Infof("已发送飞书通知给执行人: %s", apply.ExecutorName)
			}
		}()
		middleware.Logger.Infof("审批人 %s 通过了审批流程 ID: %d", apply.ApplyName, apply.ID)
	} else {
		middleware.Logger.Infof("审批人 %s 拒绝了审批流程 ID: %d", apply.ApplyName, apply.ID)
	}

	return nil
}

// executorApplyOperator 处理执行人的操作
func executorApplyOperator(tx *gorm.DB, req *models.UpdateApplyRequest, apply *models.SQLApply, w http.ResponseWriter, source string) error {
	middleware.Logger.Infof("执行人操作参数: ID=%d, ProcessType=%d, Comment=%s", req.ID, req.ProcessType, req.Comment)

	// ProcessType=1 执行SQL，ProcessType=0 驳回（不执行SQL）
	if req.ProcessType == 1 {
		// 执行SQL操作
		// 更新状态为执行中
		if err := tx.Model(&models.SQLApply{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"executor_status":  models.ApplyRoleCannotProcess,
			"submitter_status": models.ApplyRoleCannotProcess,
			"apply_status":     models.ApplyRoleCannotProcess,
			"executor_source":  source,
			"status":           models.ApplyStatusExecuting, // 执行中
		}).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "执行操作失败: "+err.Error())
			return err
		}

		// 获取更新后的记录
		if err := tx.Where("id = ?", req.ID).First(apply).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
			return err
		}

		middleware.Logger.Infof("执行人 %s 开始执行SQL，流程 ID: %d", apply.ExecutorName, apply.ID)

		// 判断是否有定时执行时间
		if apply.ExecutionTime != nil && apply.ExecutionTime.After(time.Now()) {
			// 有定时时间且未到期，等待定时任务执行
			middleware.Logger.Infof("SQL设置了定时执行，执行时间: %v，流程 ID: %d", apply.ExecutionTime, apply.ID)
			return nil
		}

		// 没有定时时间或已到期，立即执行
		go executeApplySQL(apply.ID, apply.SQLContent, apply.DatabaseName, apply.Project)

		return nil
	} else {
		// 驳回操作：不执行SQL
		if err := tx.Model(&models.SQLApply{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"executor_status":  models.ApplyRoleCannotProcess,
			"submitter_status": models.ApplyRoleCannotProcess,
			"apply_status":     models.ApplyRoleCannotProcess,
			"executor_source":  source,
			"status":           models.ApplyStatusRejected, // 驳回
		}).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "驳回操作失败: "+err.Error())
			return err
		}

		// 获取更新后的记录
		if err := tx.Where("id = ?", req.ID).First(apply).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
			return err
		}

		middleware.Logger.Infof("执行人 %s 驳回了申请，流程 ID: %d，不执行SQL", apply.ExecutorName, apply.ID)
		return nil
	}
}

// executeApplySQL 执行SQL并更新结果
func executeApplySQL(applyID int64, sqlContent, dbName, project string) {
	middleware.Logger.Infof("========== 开始真实SQL执行 ==========")
	middleware.Logger.Infof("流程 ID: %d", applyID)
	middleware.Logger.Infof("项目: %s", project)
	middleware.Logger.Infof("数据库: %s", dbName)
	middleware.Logger.Infof("SQL内容: %s", sqlContent)
	middleware.Logger.Infof("正在调用Agent执行接口...")

	// 调用Agent执行SQL
	result, err := executeSQLViaAgent(project, dbName, sqlContent)

	executionTime := time.Now()
	var finalStatus int
	var executionResult string

	if err != nil {
		finalStatus = models.ApplyStatusFailed
		executionResult = err.Error()
		middleware.Logger.Errorf("========== SQL执行失败 ==========")
		middleware.Logger.Errorf("流程 ID: %d", applyID)
		middleware.Logger.Errorf("错误信息: %v", err)
	} else {
		finalStatus = models.ApplyStatusCompleted
		executionResult = result
		middleware.Logger.Infof("========== SQL执行成功 ==========")
		middleware.Logger.Infof("流程 ID: %d", applyID)
		middleware.Logger.Infof("执行结果: %s", result)
	}

	// 更新执行结果
	if err := database.DB.Model(&models.SQLApply{}).Where("id = ?", applyID).Updates(map[string]interface{}{
		"status":                finalStatus,
		"execution_result":      executionResult,
		"actual_execution_time": executionTime,
	}).Error; err != nil {
		middleware.Logger.Errorf("更新执行结果失败: %v", err)
		return
	}

	// 获取最新记录
	var updatedApply models.SQLApply
	if err := database.DB.Where("id = ?", applyID).First(&updatedApply).Error; err != nil {
		middleware.Logger.Errorf("获取更新后的记录失败: %v", err)
		return
	}

	middleware.Logger.Infof("========== SQL执行流程完成 ==========")
	middleware.Logger.Infof("流程 ID: %d, 最终状态: %s", applyID, models.GetApplyStatusText(finalStatus))
}
