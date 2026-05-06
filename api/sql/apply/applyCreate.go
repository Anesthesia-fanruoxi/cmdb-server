package apply

import (
	"encoding/json"
	"net/http"
	"time"

	"cmdb/api/sql/rule"
	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
)

// CreateApplyRequest 创建审批请求
type CreateApplyRequest struct {
	Project       string `json:"project"`
	SQLContent    string `json:"sql_content"`
	DatabaseName  string `json:"database_name"`
	Description   string `json:"description"`
	Remark        string `json:"remark"`         // 备注（兼容前端字段）
	ExecutionTime string `json:"execution_time"` // 可选的执行时间
	ApplyID       int64  `json:"apply_id"`       // 审批人ID
	ApplyName     string `json:"apply_name"`     // 审批人姓名
	ExecutorID    int64  `json:"executor_id"`    // 执行人ID
	ExecutorName  string `json:"executor_name"`  // 执行人姓名
}

// SqlApplyCreate 创建SQL审批
func SqlApplyCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 1. 获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 2. 解析请求
	var req CreateApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 3. 验证必填字段
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目名称不能为空")
		return
	}
	if req.SQLContent == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "SQL内容不能为空")
		return
	}
	if req.DatabaseName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "数据库名称不能为空")
		return
	}
	if req.ApplyID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "审批人ID不能为空")
		return
	}
	if req.ExecutorID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "执行人ID不能为空")
		return
	}

	// 兼容description和remark字段
	description := req.Description
	if description == "" {
		description = req.Remark
	}

	middleware.Logger.Infof("用户 %s 创建SQL审批，项目: %s, 数据库: %s, 审批人: %s, 执行人: %s",
		userInfo.UserName, req.Project, req.DatabaseName, req.ApplyName, req.ExecutorName)

	// 4. 执行规则检查（入库前强制检查，防止绕过前端直接调用接口）
	ruleCheckResult, err := rule.CheckSQLRulesWithEstimate(req.SQLContent, req.Project, req.DatabaseName)
	if err != nil {
		middleware.Logger.Errorf("规则检查失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "规则检查失败: "+err.Error())
		return
	}

	middleware.Logger.Infof("规则检查完成，有违规: %v, 有阻断: %v",
		ruleCheckResult.HasViolation, ruleCheckResult.HasBlocker)

	// 5. 如果有阻断级别的违规，拒绝入库
	if ruleCheckResult.HasBlocker {
		middleware.Logger.Warnf("SQL规则检查未通过，拒绝创建审批，用户: %s", userInfo.UserName)
		middleware.ResponseErrorWithData(w, http.StatusBadRequest, "SQL规则检查未通过，无法创建审批", ruleCheckResult)
		return
	}

	// 5. 序列化规则检查结果
	ruleCheckJSON, err := json.Marshal(ruleCheckResult)
	if err != nil {
		middleware.Logger.Errorf("序列化规则检查结果失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化规则检查结果失败")
		return
	}

	// 6. 解析执行时间（支持多种格式，无效时间置空）
	var executionTime *time.Time
	if req.ExecutionTime != "" {
		// 使用中国时区解析时间
		chinaLoc, _ := time.LoadLocation("Asia/Shanghai")
		formats := []string{
			"2006-01-02T15:04:05.000Z", // ISO 8601 with milliseconds
			"2006-01-02T15:04:05Z",     // ISO 8601
			"2006-01-02 15:04:05",      // 标准格式
		}
		for _, format := range formats {
			if t, err := time.ParseInLocation(format, req.ExecutionTime, chinaLoc); err == nil {
				executionTime = &t
				break
			}
		}
		// 解析失败则置空，不报错
		if executionTime == nil {
			middleware.Logger.Warnf("执行时间格式无效，已置空: %s", req.ExecutionTime)
		}
	}

	// 7. 获取当前中国时区时间
	chinaTime := middleware.GetChinaTime()

	// 8. 判断提交人和审批人是否为同一人
	// 执行审批流程特殊逻辑：如果提交人就是审批人，则跳过审批环节，直接进入待执行状态
	isSelfApproval := userInfo.UserID == req.ApplyID

	// 9. 根据是否自审批设置初始状态
	var initialStatus int
	var applyStatus int
	var executorStatus int
	var submitterStatus int

	if isSelfApproval {
		// 提交人=审批人，跳过审批，直接待执行
		initialStatus = models.ApplyStatusApproved   // 待执行
		applyStatus = models.ApplyRoleCannotProcess  // 审批人不需要处理（已自动通过）
		executorStatus = models.ApplyRoleCanProcess  // 执行人可以处理
		submitterStatus = models.ApplyRoleCanProcess // 提交人可以撤销
		middleware.Logger.Infof("提交人与审批人相同(ID=%d)，跳过审批环节，直接进入待执行状态", userInfo.UserID)
	} else {
		// 正常流程，需要审批
		initialStatus = models.ApplyStatusPending      // 待审批
		applyStatus = models.ApplyRoleCanProcess       // 审批人可以处理
		executorStatus = models.ApplyRoleCannotProcess // 执行人还不能处理
		submitterStatus = models.ApplyRoleCanProcess   // 提交人可以撤销
	}

	// 获取提交人昵称，优先使用NickName，没有则用UserName
	submitterName := userInfo.NickName
	if submitterName == "" {
		submitterName = userInfo.UserName
	}

	// 10. 构建审批记录
	apply := &models.SQLApply{
		Project:         req.Project,
		SubmitterID:     userInfo.UserID,
		SubmitterName:   submitterName,
		SubmitterStatus: submitterStatus,
		SQLContent:      req.SQLContent,
		DatabaseName:    req.DatabaseName,
		ExecutionTime:   executionTime,
		Description:     description,
		RuleCheckResult: string(ruleCheckJSON),
		HasViolation:    boolToInt(ruleCheckResult.HasViolation),
		HasBlocker:      boolToInt(ruleCheckResult.HasBlocker),
		ApplyStatus:     applyStatus,
		ApplyID:         req.ApplyID,
		ApplyName:       req.ApplyName,
		ExecutorStatus:  executorStatus,
		ExecutorID:      req.ExecutorID,
		ExecutorName:    req.ExecutorName,
		Status:          initialStatus,
		CreatedAt:       chinaTime,
		UpdatedAt:       chinaTime,
	}

	// 11. 保存到数据库
	if err := database.DB.Create(apply).Error; err != nil {
		middleware.Logger.Errorf("创建审批记录失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "创建审批记录失败")
		return
	}

	middleware.Logger.Infof("SQL审批创建成功，ID: %d, 有违规: %v, 有阻断: %v, 自审批: %v",
		apply.ID, ruleCheckResult.HasViolation, ruleCheckResult.HasBlocker, isSelfApproval)

	// 12. 发送飞书通知（异步）
	go func() {
		if isSelfApproval {
			// 自审批情况，通知执行人
			if err := SendSQLApplyNotificationCard(apply, "executor"); err != nil {
				middleware.Logger.Errorf("发送执行人飞书通知失败: %v", err)
			} else {
				middleware.Logger.Infof("已发送飞书通知给执行人: %s（自审批跳过审批环节）", apply.ExecutorName)
			}
		} else {
			// 正常流程，通知审批人
			if err := SendSQLApplyNotificationCard(apply, "apply"); err != nil {
				middleware.Logger.Errorf("发送审批人飞书通知失败: %v", err)
			} else {
				middleware.Logger.Infof("已发送飞书通知给审批人: %s", apply.ApplyName)
			}
		}
	}()

	// 13. 返回结果
	response := map[string]interface{}{
		"id":                apply.ID,
		"has_violation":     ruleCheckResult.HasViolation,
		"has_blocker":       ruleCheckResult.HasBlocker,
		"rule_check_result": ruleCheckResult,
		"status":            apply.Status,
		"status_text":       models.GetApplyStatusText(apply.Status),
		"is_self_approval":  isSelfApproval,
	}

	middleware.ResponseSuccess(w, "创建审批成功", response)
}

// boolToInt 将bool转换为int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
