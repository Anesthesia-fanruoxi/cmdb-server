package dict

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DictUpdateItemReq struct {
	TableName string `json:"tableName"` // 表名
	Key       string `json:"key"`       // 键
	Value     string `json:"value"`     // 值
	Color     string `json:"color"`     // 颜色
}

// UpdateDictItem 更新字典项
func UpdateDictItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req DictUpdateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 参数验证
	if req.TableName == "" || req.Key == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "表名和键名不能为空")
		return
	}

	// 获取字典记录
	var dictRecord system.DictRecord
	if err := database.DB.Where("target_table = ?", req.TableName).First(&dictRecord).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "字典不存在")
		return
	}

	// 构建更新语句
	var updates []string
	var args []interface{}

	if req.Value != "" {
		updates = append(updates, fmt.Sprintf("%s = ?", dictRecord.ValueName))
		args = append(args, req.Value)
	}
	if req.Color != "" {
		// 验证颜色格式
		if !strings.HasPrefix(req.Color, "#") || len(req.Color) != 7 {
			middleware.ResponseError(w, http.StatusBadRequest, "颜色格式错误，应为#RRGGBB格式")
			return
		}
		updates = append(updates, "color = ?")
		args = append(args, req.Color)
	}

	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有需要更新的字段")
		return
	}

	// 添加更新时间和更新人
	updates = append(updates, "created_by = ?", "created_at = CURRENT_TIMESTAMP(3)")
	args = append(args, userInfo.UserID)

	// 添加WHERE条件参数
	args = append(args, req.Key)

	// 构建完整的更新语句
	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s = ?",
		req.TableName,
		strings.Join(updates, ", "),
		dictRecord.KeyName,
	)

	// 执行更新
	result := database.DB.Exec(query, args...)
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新字典项失败")
		return
	}

	// 检查是否有记录被更新
	if result.RowsAffected == 0 {
		middleware.ResponseError(w, http.StatusNotFound, "未找到要更新的字典项")
		return
	}

	middleware.ResponseSuccess(w, "更新字典项成功", nil)
}
