package dict

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"net/http"
)

// DictList 获取字典列表
func DictList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var dictRecords []system.DictRecord
	if err := database.DB.Find(&dictRecords).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取字典列表失败")
		return
	}

	middleware.ResponseSuccess(w, "获取字典列表成功", dictRecords)
}
