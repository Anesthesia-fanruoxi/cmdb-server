package process

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io"
	"net/http"
)

// SqlProcessCreate 创建SQL审批流程 - 使用GORM
func SqlProcessCreate(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "关闭请求体失败: "+closeErr.Error())
		}
	}()

	// 解析JSON
	var flow sql.SQLApplyFlow
	if err := json.Unmarshal(body, &flow); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析JSON失败")
		return
	}

	// 简单验证
	if flow.ApplyID == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "默认审批人不能为空")
		return
	}

	if flow.ExecutorID == nil {
		middleware.ResponseError(w, http.StatusBadRequest, "默认执行人不能为空")
		return
	}

	// 使用GORM创建记录
	result := database.DB.Create(&flow)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建SQL审批流程失败: "+result.Error.Error())
		return
	}

	// 返回响应
	middleware.ResponseSuccess(w, "创建审批记录成功", nil)
}
