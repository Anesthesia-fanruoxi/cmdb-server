package process

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// ProcessUserItem 审批/执行人员信息
type ProcessUserItem struct {
	ID       int64  `json:"id" gorm:"column:id"`
	UserName string `json:"user_name" gorm:"column:user_name"`
	NickName string `json:"nick_name" gorm:"column:nick_name"`
}

// ProcessUsersResponse 审批人员响应
type ProcessUsersResponse struct {
	Approvers []ProcessUserItem `json:"approvers"` // 审批人列表
	Executors []ProcessUserItem `json:"executors"` // 执行人列表
}

// SqlProcessUsers 获取审批人和执行人列表
func SqlProcessUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 查询审批人：角色ID = 1(超级管理员) 或 3(主管)
	var approvers []ProcessUserItem
	err := database.DB.Table(middleware.TableUsers).
		Select("id, user_name, nick_name").
		Where("is_enabled = ? AND role_id IN ?", true, []int64{1, 3}).
		Order("role_id ASC, id ASC").
		Find(&approvers).Error

	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询审批人失败: "+err.Error())
		return
	}

	// 查询执行人：角色ID = 1(超级管理员) 或 2(运维)
	var executors []ProcessUserItem
	err = database.DB.Table(middleware.TableUsers).
		Select("id, user_name, nick_name").
		Where("is_enabled = ? AND role_id IN ?", true, []int64{1, 2}).
		Order("role_id ASC, id ASC").
		Find(&executors).Error

	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询执行人失败: "+err.Error())
		return
	}

	response := ProcessUsersResponse{
		Approvers: approvers,
		Executors: executors,
	}

	middleware.ResponseSuccess(w, "获取成功", response)
}
