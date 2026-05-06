package rule

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// SqlRuleUpdate 更新规则状态
func SqlRulesUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 从URL获取ID参数
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少ID参数")
		return
	}

	// 转换ID为整数
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的ID参数")
		return
	}

	var req struct {
		IsEnabled int `json:"is_enable"` // 前端字段名是 is_enable
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据无效")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if req.IsEnabled != 0 && req.IsEnabled != 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "is_enabled只能为0或1")
		return
	}

	// 先检查规则是否存在
	var count int64
	if err := database.DB.Table(middleware.TableSqlRule).Where("id = ?", id).Count(&count).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询规则失败")
		return
	}

	if count == 0 {
		middleware.ResponseError(w, http.StatusNotFound, "规则不存在")
		return
	}

	// 更新指定ID的规则
	result := database.DB.Table(middleware.TableSqlRule).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_enabled": req.IsEnabled,
		})

	if result.Error != nil {
		middleware.Logger.Errorf("更新规则失败: %v", result.Error)
		middleware.ResponseError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	middleware.Logger.Infof("更新规则 ID=%d, is_enabled=%d, 影响行数=%d", id, req.IsEnabled, result.RowsAffected)

	if result.RowsAffected == 0 {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新失败，未影响任何行")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", map[string]interface{}{
		"id":            id,
		"is_enabled":    req.IsEnabled,
		"rows_affected": result.RowsAffected,
	})
}
