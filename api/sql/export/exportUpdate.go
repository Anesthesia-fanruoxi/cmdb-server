package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// SqlExportUpdate 处理导出流程的更新请求
func SqlExportUpdate(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
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
	var req sql.UpdateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 验证流程ID是否存在
	if req.ID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID不能为空")
		return
	}

	// 获取当前流程信息
	var export sql.SQLExport
	if err := database.DB.Where("id = ?", req.ID).First(&export).Error; err != nil {
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

	// 判断当前用户角色并处理
	var err error

	// 判断当前用户是哪个角色
	userRole := ""
	if export.SubmitterID == userInfo.UserID {
		userRole = "submitter"
	} else if export.ApplyID == userInfo.UserID {
		userRole = "approver"
	} else if export.ReviewerID == userInfo.UserID {
		userRole = "reviewer"
	} else if export.ExecutorID == userInfo.UserID {
		userRole = "executor"
	}

	// 如果找不到有效角色，直接返回错误
	if userRole == "" {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, "您不是该流程的相关人员")
		return
	}

	// 使用CheckExportPermission检查权限
	if err := CheckExportPermission(req.ID, userInfo.UserID, userRole); err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, err.Error())
		return
	}

	// 根据角色和请求进行不同的处理
	switch userRole {
	case "submitter":
		err = submitterOperator(tx, &req, &export, w)
	case "approver":
		err = approverOperator(tx, &req, &export, w, "pc")
	case "reviewer":
		err = reviewerOperator(tx, &req, &export, w, "pc")
	case "executor":
		err = executorOperator(tx, &req, &export, w, "pc")
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

func SqlExportUrlUpdate(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
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
	var req sql.UpdateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 验证流程ID是否存在
	if req.ID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID不能为空")
		return
	}

	// 获取当前流程信息
	var export sql.SQLExport
	if err := database.DB.Where("id = ?", req.ID).First(&export).Error; err != nil {
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

	// 判断当前用户角色并处理
	var err error

	// 判断当前用户是哪个角色
	userRole := ""
	if export.SubmitterID == userInfo.UserID {
		userRole = "submitter"
	} else if export.ApplyID == userInfo.UserID {
		userRole = "approver"
	} else if export.ReviewerID == userInfo.UserID {
		userRole = "reviewer"
	} else if export.ExecutorID == userInfo.UserID {
		userRole = "executor"
	}

	// 如果找不到有效角色，直接返回错误
	if userRole == "" {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, "您不是该流程的相关人员")
		return
	}

	// 使用CheckExportPermission检查权限
	if err := CheckExportPermission(req.ID, userInfo.UserID, userRole); err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusForbidden, err.Error())
		return
	}

	// 根据角色和请求进行不同的处理
	switch userRole {
	case "submitter":
		err = submitterOperator(tx, &req, &export, w)
	case "approver":
		err = approverOperator(tx, &req, &export, w, "mobile")
	case "reviewer":
		err = reviewerOperator(tx, &req, &export, w, "mobile")
	case "executor":
		err = executorOperator(tx, &req, &export, w, "mobile")
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

// submitterOperator 处理提交人的操作
func submitterOperator(tx *gorm.DB, req *sql.UpdateExportRequest, export *sql.SQLExport, w http.ResponseWriter) error {
	// 提交人只能撤销流程
	if req.ProcessType != 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "提交人只能撤销流程")
		return errors.New("提交人只能撤销流程")
	}

	// 提交人撤销操作
	if err := tx.Model(&sql.SQLExport{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":           sql.StatusCanceled,
		"submitter_status": sql.StatusCannotProcess,
		"apply_status":     sql.StatusCannotProcess,
		"reviewer_status":  sql.StatusCannotProcess,
		"executor_status":  sql.StatusCannotProcess,
	}).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "撤销操作失败: "+err.Error())
		return err
	}

	// 获取更新后的记录并更新Redis
	if err := tx.Where("id = ?", req.ID).First(export).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	if err := sql.UpdateExportRedis(export); err != nil {
		middleware.Logger.Errorf("更新Redis失败: %v", err)
	}

	return nil
}

// approverOperator 处理审批人的操作
func approverOperator(tx *gorm.DB, req *sql.UpdateExportRequest, export *sql.SQLExport, w http.ResponseWriter, source string) error {
	// 检查是否已有SQL和数据库名称，如果没有则需要提供
	if export.SQLContent == "" || export.DatabaseName == "" {
		if req.DatabaseName == "" || req.SQLContent == "" {
			middleware.ResponseError(w, http.StatusBadRequest, "审批需要提供数据库名称和SQL内容")
			return errors.New("审批需要提供数据库名称和SQL内容")
		}
	}

	var statusValue int
	var reviewerCanProcess int
	var executorCanProcess int

	// 检查审批人和审核人是否是同一个人
	isSamePerson := export.ApplyID == export.ReviewerID

	// 根据处理类型设置状态
	if req.ProcessType == 1 {
		// 审批通过
		if isSamePerson {
			// 如果审批人和审核人是同一个人，直接跳过审核环节
			middleware.Logger.Infof("审批人和审核人是同一个人(ID:%d)，直接跳过审核环节", export.ApplyID)
			statusValue = sql.StatusExec // 直接进入待执行状态
			reviewerCanProcess = sql.StatusCannotProcess
			executorCanProcess = sql.StatusCanProcess
		} else {
			// 正常流程，进入审核环节
			statusValue = sql.StatusAudit
			reviewerCanProcess = sql.StatusCanProcess
			executorCanProcess = sql.StatusCannotProcess
		}
	} else {
		// 审批拒绝
		statusValue = sql.StatusRejected
		reviewerCanProcess = sql.StatusCannotProcess
		executorCanProcess = sql.StatusCannotProcess
	}

	// 准备更新字段
	updateFields := map[string]interface{}{
		"apply_time":       time.Now(),
		"apply_status":     sql.StatusCannotProcess,
		"apply_source":     source,
		"submitter_status": sql.StatusCannotProcess,
		"reviewer_status":  reviewerCanProcess,
		"executor_status":  executorCanProcess,
		"status":           statusValue,
	}

	// 如果审批人和审核人是同一个人，自动设置审核时间
	if isSamePerson && req.ProcessType == 1 {
		updateFields["reviewer_time"] = time.Now()
		updateFields["reviewer_source"] = source
	}

	// 如果拒绝，关闭所有后续角色的处理权限
	if req.ProcessType == 0 {
		updateFields["executor_status"] = sql.StatusCannotProcess
	}

	// 如果请求中提供了数据库名称和SQL内容，则更新这些字段
	if req.DatabaseName != "" && req.SQLContent != "" {
		updateFields["database_name"] = req.DatabaseName
		updateFields["sql_content"] = req.SQLContent
	}

	if err := tx.Model(&sql.SQLExport{}).Where("id = ?", req.ID).Updates(updateFields).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "审批操作失败: "+err.Error())
		return err
	}

	// 获取更新后的记录并更新Redis
	if err := tx.Where("id = ?", req.ID).First(export).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	if err := sql.UpdateExportRedis(export); err != nil {
		middleware.Logger.Errorf("更新Redis失败: %v", err)
	}

	// 审批通过后，根据情况发送通知
	if req.ProcessType == 1 {
		if isSamePerson {
			// 如果审批人和审核人是同一个人，直接通知执行人
			go func() {
				if err := sql.SendApprovalNotificationCard(export, "executor"); err != nil {
					middleware.Logger.Errorf("发送执行人卡片通知失败: %v", err)
				} else {
					middleware.Logger.Infof("已发送执行卡片通知给执行人: %s", export.ExecutorName)
				}
			}()
		} else {
			// 正常流程，通知审核人
			go func() {
				if err := sql.SendApprovalNotificationCard(export, "reviewer"); err != nil {
					middleware.Logger.Errorf("发送审核人卡片通知失败: %v", err)
				} else {
					middleware.Logger.Infof("已发送审核卡片通知给审核人: %s", export.ReviewerName)
				}
			}()
		}
	}

	return nil
}

// reviewerOperator 处理审核人的操作
func reviewerOperator(tx *gorm.DB, req *sql.UpdateExportRequest, export *sql.SQLExport, w http.ResponseWriter, source string) error {
	var err error
	if req.ProcessType == 1 {
		// 审核通过
		if err = tx.Model(&sql.SQLExport{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"reviewer_time":   time.Now(),
			"reviewer_status": sql.StatusCannotProcess,
			"reviewer_source": source,
			"executor_status": sql.StatusCanProcess,
			"status":          sql.StatusExec,
		}).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "审核操作失败: "+err.Error())
			return err
		}
	} else {
		// 审核拒绝
		if err = tx.Model(&sql.SQLExport{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"reviewer_time":    time.Now(),
			"reviewer_status":  sql.StatusCannotProcess,
			"reviewer_source":  source,
			"submitter_status": sql.StatusCannotProcess,
			"apply_status":     sql.StatusCannotProcess,
			"executor_status":  sql.StatusCannotProcess,
			"status":           sql.StatusRejected,
		}).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "拒绝操作失败: "+err.Error())
			return err
		}
	}

	// 获取更新后的记录并更新Redis
	if err := tx.Where("id = ?", req.ID).First(export).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	if err := sql.UpdateExportRedis(export); err != nil {
		middleware.Logger.Errorf("更新Redis失败: %v", err)
	}

	// 审核通过后，发送卡片给执行人
	if req.ProcessType == 1 {
		go func() {
			if err := sql.SendApprovalNotificationCard(export, "executor"); err != nil {
				middleware.Logger.Errorf("发送执行人卡片通知失败: %v", err)
			} else {
				middleware.Logger.Infof("已发送执行卡片通知给执行人: %s", export.ExecutorName)
			}
		}()
	}

	return nil
}

// executorOperator 处理执行人的操作
func executorOperator(tx *gorm.DB, req *sql.UpdateExportRequest, export *sql.SQLExport, w http.ResponseWriter, resource string) error {
	// 更新状态为SQL执行中
	if err := tx.Model(&sql.SQLExport{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"executor_status":  sql.StatusCannotProcess,
		"submitter_status": sql.StatusCannotProcess,
		"apply_status":     sql.StatusCannotProcess,
		"reviewer_status":  sql.StatusCannotProcess,
		"executor_source":  resource,
		"status":           sql.StatusExecuting, // 执行中
	}).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "执行操作失败: "+err.Error())
		return err
	}

	// 获取更新后的记录并更新Redis
	if err := tx.Where("id = ?", req.ID).First(export).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的记录失败: "+err.Error())
		return err
	}

	if err := sql.UpdateExportRedis(export); err != nil {
		middleware.Logger.Errorf("更新Redis失败: %v", err)
	}

	// 获取流程信息用于执行SQL
	var exportInfo sql.SQLExport
	if err := tx.Where("id = ?", req.ID).First(&exportInfo).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取流程信息失败: "+err.Error())
		return err
	}

	// 获取项目代码
	var projectCode string
	err := database.DB.Table("sys_project_dict").
		Select("project").
		Where("id = ?", exportInfo.Project).
		Row().
		Scan(&projectCode)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取项目代码失败: "+err.Error())
		return err
	}

	if projectCode == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "项目代码为空")
		return errors.New("项目代码为空")
	}

	// 只有在明确传递ProcessType=1时才真正执行SQL
	if req.ProcessType == 1 {
		middleware.Logger.Infof("开始执行SQL导出任务, ID: %d, Project: %s, DB: %s",
			exportInfo.ID, projectCode, exportInfo.DatabaseName)

		// 提交事务后异步执行SQL
		go func(project, dbName, sqlContent string, submitterID, exportID int64) {
			result, err := ExecuteSQL(project, dbName, sqlContent, submitterID, exportID)
			if err != nil {
				middleware.Logger.Errorf("SQL执行失败: %v", err)
				return
			}

			if len(result.Errors) > 0 {
				middleware.Logger.Errorf("SQL执行有错误: %v", result.Errors)
			} else {
				middleware.Logger.Infof("SQL执行成功，流程ID: %d, 行数: %d, 文件: %s",
					exportID, result.ResultCount, result.ExcelFilePath)
			}

			// 获取最新记录并更新Redis
			var updatedExport sql.SQLExport
			if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err != nil {
				middleware.Logger.Errorf("获取更新后的记录失败: %v", err)
				return
			}

			if err := sql.UpdateExportRedis(&updatedExport); err != nil {
				middleware.Logger.Errorf("更新Redis失败: %v", err)
			}
		}(projectCode, exportInfo.DatabaseName, exportInfo.SQLContent, exportInfo.SubmitterID, exportInfo.ID)
	} else {
		middleware.Logger.Infof("不执行SQL，直接标记为成功, ID: %d", exportInfo.ID)

		// 异步更新状态为成功
		go func(exportID int64) {
			time.Sleep(time.Second)

			// 更新执行结果状态为成功
			if updateErr := database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).Update("status", sql.StatusCompleted).Error; updateErr != nil {
				middleware.Logger.Errorf("更新执行结果状态失败: %v", updateErr)
				return
			}

			// 获取最新记录并更新Redis
			var updatedExport sql.SQLExport
			if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err != nil {
				middleware.Logger.Errorf("获取更新后的记录失败: %v", err)
				return
			}

			if err := sql.UpdateExportRedis(&updatedExport); err != nil {
				middleware.Logger.Errorf("更新Redis失败: %v", err)
			}
		}(exportInfo.ID)
	}

	return nil
}
