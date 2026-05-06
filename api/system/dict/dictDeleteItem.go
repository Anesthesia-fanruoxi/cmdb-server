package dict

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"fmt"
	"net/http"
)

// DeleteDictItem 删除字典项
func DeleteDictItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	tableName := r.URL.Query().Get("target_table")
	key := r.URL.Query().Get("key")

	if tableName == "" || key == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "表名和键不能为空")
		return
	}

	// 获取字典记录，以获取key_name
	var dictRecord system.DictRecord
	if err := database.DB.Where("target_table = ?", tableName).First(&dictRecord).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "字典不存在")
		return
	}

	// 使用原生SQL删除数据
	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?",
		tableName,
		dictRecord.KeyName,
	)

	if err := database.DB.Exec(deleteSQL, key).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "删除字典项失败")
		return
	}

	middleware.ResponseSuccess(w, "删除字典项成功", nil)
}
