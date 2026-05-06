package task

import (
	"cmdb/api/job/scheduler"
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// JobDelete 删除任务
func JobDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取任务ID
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的任务ID")
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 删除任务
	if err := tx.Delete(&job.JobInfo{}, id).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除任务失败")
		return
	}

	// 删除关联的项目绑定
	if err := tx.Where("job_id = ?", id).Delete(&job.JobProject{}).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除项目绑定失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	// 从调度器移除任务
	jobScheduler := scheduler.GetScheduler()
	jobScheduler.RemoveJob(id)

	middleware.ResponseSuccess(w, "删除成功", nil)
}
