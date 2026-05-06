package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"errors"
)

// CheckExportPermission 检查用户对导出流程的操作权限
// role: "submitter", "approver", "reviewer", "executor"
func CheckExportPermission(exportID int64, userID int64, role string) error {
	db := database.DB
	var export sql.SQLExport

	if err := db.Where("id = ?", exportID).First(&export).Error; err != nil {
		return errors.New("导出流程不存在")
	}

	// 验证当前用户是否为对应角色
	switch role {
	case "submitter":
		if export.SubmitterID != userID {
			return errors.New("您不是该流程的发起人")
		}
		if export.SubmitterStatus != sql.StatusCanProcess {
			return errors.New("当前状态不允许发起人操作")
		}

		// 检查流程状态是否允许撤销
		if export.Status == sql.StatusCompleted || export.Status == sql.StatusFailed ||
			export.Status == sql.StatusRejected || export.Status == sql.StatusCanceled {
			return errors.New("流程已结束，不能再操作")
		}

	case "approver":
		if export.ApplyID != userID {
			return errors.New("您不是该流程的审批人")
		}
		if export.ApplyStatus != sql.StatusCanProcess {
			return errors.New("当前状态不允许进行审批操作")
		}

		// 检查流程整体状态
		if export.Status != sql.StatusDraft {
			return errors.New("流程不在审批阶段")
		}

	case "reviewer":
		if export.ReviewerID != userID {
			return errors.New("您不是该流程的审核人")
		}
		if export.ReviewerStatus != sql.StatusCanProcess {
			return errors.New("当前状态不允许进行审核操作")
		}

		// 检查流程整体状态
		if export.Status != sql.StatusAudit {
			return errors.New("流程不在审核阶段")
		}

	case "executor":
		if export.ExecutorID != userID {
			return errors.New("您不是该流程的执行人")
		}
		if export.ExecutorStatus != sql.StatusCanProcess {
			return errors.New("当前状态不允许进行执行操作")
		}

		// 检查流程整体状态
		if export.Status != sql.StatusExec {
			return errors.New("流程不在执行阶段")
		}

	default:
		return errors.New("未知的角色类型")
	}

	return nil
}
