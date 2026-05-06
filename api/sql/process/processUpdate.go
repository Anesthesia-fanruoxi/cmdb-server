package process

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io"
	"net/http"
)

// SqlProcessUpdate 更新SQL审批流程
func SqlProcessUpdate(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查HTTP方法
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持PUT请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭请求体失败: %v", closeErr)
		}
	}()

	// 解析JSON
	var updatedFlow sql.SQLApplyFlow
	if err := json.Unmarshal(body, &updatedFlow); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 检查流程是否存在
	var existingFlow sql.SQLApplyFlow
	result := database.DB.First(&existingFlow, updatedFlow.ID)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		return
	}

	// 设置ID，确保更新的是正确的记录
	updatedFlow.ID = existingFlow.ID

	// 准备更新数据
	updates := make(map[string]interface{})

	if updatedFlow.ApplyID != nil {
		updates["apply_id"] = updatedFlow.ApplyID
	}
	if updatedFlow.ExecutorID != nil {
		updates["executor_id"] = updatedFlow.ExecutorID
	}

	// 检查是否有字段需要更新
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有提供需要更新的字段")
		return
	}

	// 更新流程
	result = database.DB.Model(&existingFlow).Updates(updates)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新流程失败: "+result.Error.Error())
		return
	}

	// 获取更新后的流程
	var flow sql.SQLApplyFlow
	database.DB.First(&flow, updatedFlow.ID)

	// 返回响应
	middleware.ResponseSuccess(w, "流程更新成功", flow)
}
