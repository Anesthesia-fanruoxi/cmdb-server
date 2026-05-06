package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// FeiShuUpdateRequest 飞书请求结构体
type FeiShuUpdateRequest struct {
	ExportID     int64  `json:"export_id" binding:"required"`    // 流程ID
	ProcessType  int    `json:"process_type" binding:"required"` // 处理类型：1通过，0拒绝
	Action       string `json:"action" binding:"required"`       // 操作类型：approve, review_pass, review_reject, execute等
	UserEmail    string `json:"user_email,omitempty"`            // 操作用户邮箱（可选，用于验证）
	DatabaseName string `json:"database_name,omitempty"`         // 数据库名称（仅审批人需要）
	SQLContent   string `json:"sql_content,omitempty"`           // SQL内容（仅审批人需要）
}

// FeiShuExportUpdate 飞书机器人专用的导出流程更新接口
func FeiShuExportUpdate(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 解析请求体
	var req FeiShuUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	middleware.Logger.Infof("飞书更新请求 - ExportID: %d, Action: %s, ProcessType: %d", req.ExportID, req.Action, req.ProcessType)

	// 验证流程ID是否存在
	if req.ExportID <= 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID不能为空")
		return
	}

	// 获取当前流程信息
	var export sql.SQLExport
	if err := database.DB.Where("id = ?", req.ExportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询流程失败")
		}
		return
	}

	// 使用分布式锁防止重复操作
	lockKey := fmt.Sprintf("sql_export_feishu_lock:%d", req.ExportID)
	success, err := database.RDB.SetNX(context.Background(), lockKey, "1", 30*time.Second).Result()
	if err != nil {
		middleware.Logger.Errorf("获取分布式锁失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "系统繁忙，请稍后再试")
		return
	}
	if !success {
		middleware.ResponseError(w, http.StatusConflict, "该流程正在处理中，请稍后再试")
		return
	}
	defer database.RDB.Del(context.Background(), lockKey)

	// 验证流程状态是否允许操作
	if err := validateFeiShuOperation(&req, &export); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 开启数据库事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var operationErr error
	var userRole string
	var source = "飞书机器人"

	// 根据action确定角色和操作
	switch req.Action {
	case "approve", "reject":
		userRole = "approver"
		// 转换为标准的UpdateExportRequest格式
		updateReq := &sql.UpdateExportRequest{
			ID:           req.ExportID,
			ProcessType:  req.ProcessType,
			DatabaseName: req.DatabaseName,
			SQLContent:   req.SQLContent,
		}
		operationErr = approverOperator(tx, updateReq, &export, w, source)

	case "review_pass", "review_reject":
		userRole = "reviewer"
		updateReq := &sql.UpdateExportRequest{
			ID:          req.ExportID,
			ProcessType: req.ProcessType,
		}
		operationErr = reviewerOperator(tx, updateReq, &export, w, source)

	case "execute":
		userRole = "executor"
		updateReq := &sql.UpdateExportRequest{
			ID:          req.ExportID,
			ProcessType: 1, // 执行操作固定为通过
		}
		operationErr = executorOperator(tx, updateReq, &export, w, source)

	default:
		tx.Rollback()
		middleware.ResponseError(w, http.StatusBadRequest, "不支持的操作类型: "+req.Action)
		return
	}

	// 检查操作是否成功
	if operationErr != nil {
		tx.Rollback()
		middleware.Logger.Errorf("飞书更新操作失败 - ExportID: %d, Action: %s, Error: %v", req.ExportID, req.Action, operationErr)
		return // 错误响应已在operator函数中处理
	}

	middleware.Logger.Infof("飞书更新操作成功 - ExportID: %d, Action: %s, Role: %s", req.ExportID, req.Action, userRole)

	// 返回成功结果
	response := map[string]interface{}{
		"export_id": req.ExportID,
		"action":    req.Action,
		"status":    "success",
		"message":   "操作成功",
		"role":      userRole,
	}

	middleware.ResponseSuccess(w, "飞书操作成功", response)

	// 如果是执行操作，异步发送查询结果到提交人邮箱
	if req.Action == "execute" && req.ProcessType == 1 {
		go processSQLExecution(req.ExportID)
	}
}

// validateFeiShuOperation 验证飞书操作是否合法
func validateFeiShuOperation(req *FeiShuUpdateRequest, export *sql.SQLExport) error {
	switch req.Action {
	case "approve", "reject":
		// 验证审批人操作
		if export.Status != sql.StatusDraft || export.ApplyStatus != sql.StatusCanProcess {
			return errors.New("当前流程状态不允许审批操作")
		}

		// 如果没有SQL内容，需要提供
		if (export.SQLContent == "" || export.DatabaseName == "") &&
			(req.SQLContent == "" || req.DatabaseName == "") &&
			req.ProcessType == 1 {
			return errors.New("审批通过需要提供数据库名称和SQL内容")
		}

	case "review_pass", "review_reject":
		// 验证审核人操作
		if export.Status != sql.StatusAudit || export.ReviewerStatus != sql.StatusCanProcess {
			return errors.New("当前流程状态不允许审核操作")
		}

	case "execute":
		// 验证执行人操作
		if export.Status != sql.StatusExec || export.ExecutorStatus != sql.StatusCanProcess {
			return errors.New("当前流程状态不允许执行操作")
		}

		// 验证是否有SQL内容
		if export.SQLContent == "" || export.DatabaseName == "" {
			return errors.New("缺少SQL内容或数据库名称，无法执行")
		}

	default:
		return errors.New("不支持的操作类型")
	}

	return nil
}

// processSQLExecution 异步处理SQL执行和结果发送
func processSQLExecution(exportID int64) {
	// 使用recover防止goroutine崩溃
	defer func() {
		if r := recover(); r != nil {
			middleware.Logger.Errorf("飞书执行SQL异步处理崩溃: %v", r)
		}
	}()

	middleware.Logger.Infof("开始异步处理飞书执行SQL操作，ExportID: %d", exportID)

	// 等待一秒，确保主流程已完成
	time.Sleep(1 * time.Second)

	// 获取最新的导出信息
	var updatedExport sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&updatedExport).Error; err != nil {
		middleware.Logger.Errorf("获取导出信息失败: %v", err)
		return
	}

	// 检查SQL内容
	if updatedExport.SQLContent == "" || updatedExport.DatabaseName == "" {
		middleware.Logger.Errorf("SQL内容或数据库名为空，无法执行，ExportID: %d", exportID)
		return
	}

	// 获取提交人邮箱
	submitterEmail := sql.GetUserEmail(updatedExport.SubmitterID)
	if submitterEmail == "" {
		middleware.Logger.Errorf("获取提交人邮箱失败，ID: %d", updatedExport.SubmitterID)
		return
	}

	// 获取项目中文名称
	projectName := sql.GetProjectName(updatedExport.Project)
	if projectName == "" {
		projectName = updatedExport.Project // 如果获取不到中文名，使用原项目标识
	}

	middleware.Logger.Infof("开始执行SQL并发送结果到提交人邮箱: %s, ExportID: %d", submitterEmail, exportID)

	// 准备邮件信息
	emailInfo := &sql.EmailInfo{
		To:      []string{submitterEmail},
		Subject: fmt.Sprintf("SQL导出结果 - %s - %s", projectName, updatedExport.DatabaseName),
	}

	// 添加审批人和审核人到抄送列表
	if updatedExport.ApplyID > 0 {
		approverEmail := sql.GetUserEmail(updatedExport.ApplyID)
		if approverEmail != "" && approverEmail != submitterEmail {
			emailInfo.Cc = append(emailInfo.Cc, approverEmail)
		}
	}

	if updatedExport.ReviewerID > 0 {
		reviewerEmail := sql.GetUserEmail(updatedExport.ReviewerID)
		if reviewerEmail != "" && reviewerEmail != submitterEmail && !containsEmail(emailInfo.Cc, reviewerEmail) {
			emailInfo.Cc = append(emailInfo.Cc, reviewerEmail)
		}
	}

	// 执行SQL并保存文件到服务器
	middleware.Logger.Infof("开始执行SQL，项目: %s, 数据库: %s, ExportID: %d",
		updatedExport.Project, updatedExport.DatabaseName, exportID)

	result, err := ExecuteSQL(updatedExport.Project, updatedExport.DatabaseName, updatedExport.SQLContent, updatedExport.SubmitterID, exportID)
	if err != nil {
		middleware.Logger.Errorf("执行SQL失败: %v, ExportID: %d", err, exportID)
		return
	}

	// 记录执行结果
	if len(result.Errors) > 0 {
		middleware.Logger.Errorf("SQL执行有错误: %v, ExportID: %d", result.Errors, exportID)
	} else {
		middleware.Logger.Infof("SQL执行成功，ExportID: %d, 行数: %d, 文件: %s", exportID, result.ResultCount, result.ExcelFilePath)
	}

	middleware.Logger.Infof("SQL执行完成，文件已保存到服务器, 结果行数: %d, ExportID: %d",
		result.ResultCount, exportID)
}

// containsEmail 检查邮件列表中是否包含指定邮箱
func containsEmail(emails []string, email string) bool {
	for _, e := range emails {
		if e == email {
			return true
		}
	}
	return false
}
