package apply

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// 格式化可能为nil的时间指针
func formatDetailTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return middleware.FormatTimeToChina(*t)
}

// SqlApplyDetailResponse 自定义响应结构体
type SqlApplyDetailResponse struct {
	ID                  int64  `json:"id"`
	Project             string `json:"project"`
	SubmitterID         int64  `json:"submitter_id"`
	SubmitterName       string `json:"submitter_name"`
	SQLContent          string `json:"sql_content"`
	DatabaseName        string `json:"database_name"`
	ExecutionTime       string `json:"execution_time"`
	Description         string `json:"description"`
	HasViolation        bool   `json:"has_violation"`
	HasBlocker          bool   `json:"has_blocker"`
	ApplyID             int64  `json:"apply_id"`
	ApplyName           string `json:"apply_name"`
	ApplyComment        string `json:"apply_comment"`
	ApplyTime           string `json:"apply_time"`
	ExecutorID          int64  `json:"executor_id"`
	ExecutorName        string `json:"executor_name"`
	ExecutionResult     string `json:"execution_result"`
	ActualExecutionTime string `json:"actual_execution_time"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Status              int    `json:"status"`
	ApplyRole           string `json:"apply_role"`
}

// getDetailApplyRecord 获取审批记录详情
func getDetailApplyRecord(id int64) (sql.SQLApply, error) {
	var apply sql.SQLApply
	result := database.DB.First(&apply, id)
	if result.Error != nil {
		return sql.SQLApply{}, result.Error
	}
	return apply, nil
}

// buildDetailResponse 构建详情响应
func buildDetailResponse(apply sql.SQLApply, applyRole string) SqlApplyDetailResponse {
	return SqlApplyDetailResponse{
		ID:                  apply.ID,
		Project:             apply.Project,
		SubmitterID:         apply.SubmitterID,
		SubmitterName:       apply.SubmitterName,
		SQLContent:          apply.SQLContent,
		DatabaseName:        apply.DatabaseName,
		ExecutionTime:       formatDetailTime(apply.ExecutionTime),
		Description:         apply.Description,
		HasViolation:        apply.HasViolation == 1,
		HasBlocker:          apply.HasBlocker == 1,
		ApplyID:             apply.ApplyID,
		ApplyName:           apply.ApplyName,
		ApplyComment:        apply.ApplyComment,
		ApplyTime:           formatDetailTime(apply.ApplyTime),
		ExecutorID:          apply.ExecutorID,
		ExecutorName:        apply.ExecutorName,
		ExecutionResult:     apply.ExecutionResult,
		ActualExecutionTime: formatDetailTime(apply.ActualExecutionTime),
		CreatedAt:           middleware.FormatTimeToChina(apply.CreatedAt),
		UpdatedAt:           middleware.FormatTimeToChina(apply.UpdatedAt),
		Status:              apply.Status,
		ApplyRole:           applyRole,
	}
}

// determineApplyRole 确定用户角色
func determineApplyRole(userID int64, apply sql.SQLApply) string {
	if userID == apply.SubmitterID {
		return "submit"
	} else if userID == apply.ApplyID {
		return "apply"
	} else if userID == apply.ExecutorID {
		return "executor"
	}
	return ""
}

// checkUserPermission 检查用户是否有权限查看审批详情
func checkUserPermission(userID int64, apply sql.SQLApply) bool {
	// 检查用户是否是提交人、审批人或执行人
	return userID == apply.SubmitterID || userID == apply.ApplyID || userID == apply.ExecutorID
}

// SqlApplyDetail 获取SQL审批详情的API处理函数
func SqlApplyDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取ID参数
	query := r.URL.Query()
	idStr := query.Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少ID参数")
		return
	}

	// 将ID转换为int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "ID参数无效")
		return
	}

	// 从token获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.Logger.Error("获取用户信息失败")
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取审批记录
	apply, err := getDetailApplyRecord(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "找不到该SQL审批记录")
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取审批记录失败: "+err.Error())
		}
		return
	}

	// 检查用户权限
	if !checkUserPermission(userInfo.UserID, apply) {
		middleware.ResponseError(w, http.StatusForbidden, "您无权查看该审批详情")
		return
	}

	// 确定用户角色
	applyRole := determineApplyRole(userInfo.UserID, apply)

	// 构建并返回响应
	response := buildDetailResponse(apply, applyRole)
	middleware.ResponseSuccess(w, "获取审批记录成功", response)
}
