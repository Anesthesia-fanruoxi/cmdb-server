package process

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
	"strconv"
)

// SqlProcessDelete 删除SQL审批流程
func SqlProcessDelete(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查请求方法是否为DELETE
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许DELETE方法")
		return
	}

	// 从查询参数获取ID
	idStr := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的流程ID")
		return
	}

	// 检查流程是否存在
	var flow sql.SQLApplyFlow
	result := database.DB.First(&flow, id)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		return
	}

	// 删除流程
	result = database.DB.Delete(&flow)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除流程失败: "+result.Error.Error())
		return
	}

	// 返回响应
	middleware.ResponseSuccess(w, "流程删除成功", nil)
}
