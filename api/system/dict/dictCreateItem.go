package dict

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/initialize"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DictItemRequest 创建字典项请求
type DictItemRequest struct {
	TableName string `json:"tableName"` // 表名
	Key       string `json:"key"`       // 键
	Value     string `json:"value"`     // 值
	Color     string `json:"color"`     // 颜色
}

// CreateDictItem 创建字典项
func CreateDictItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req DictItemRequest
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber() // 使用 Number 类型处理数字
	if err := decoder.Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 使用 userInfo.UserID 作为创建人ID
	userID := userInfo.UserID

	// 验证必填字段
	if req.TableName == "" || req.Key == "" || req.Value == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "表名、键和值不能为空")
		return
	}

	// 获取字典记录，以获取key_name和value_name
	var dictRecord system.DictRecord
	if err := database.DB.Where("target_table = ?", req.TableName).First(&dictRecord).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "字典不存在")
		return
	}

	// 检查键是否已存在
	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` = ?",
		req.TableName,
		dictRecord.KeyName,
	)
	var count int64
	if err := database.DB.Raw(checkSQL, req.Key).Count(&count).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "检查键是否存在失败")
		return
	}
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "该键已存在")
		return
	}

	// 使用原生SQL插入数据
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (`%s`, `%s`, `color`, `created_by`) VALUES (?, ?, ?, ?)",
		req.TableName,
		dictRecord.KeyName,
		dictRecord.ValueName,
	)

	if err := database.DB.Exec(insertSQL, req.Key, req.Value, req.Color, userID).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			middleware.ResponseError(w, http.StatusBadRequest, "该键已存在")
			return
		}
		middleware.ResponseError(w, http.StatusInternalServerError, "创建字典项失败")
		return
	}

	// 如果是项目相关的字典，需要初始化系统部门的项目关联
	if req.TableName == middleware.TableProjectDict {
		if err := initialize.InitSystemDeptProjects(); err != nil {
			fmt.Printf("初始化系统部门项目关联失败: %v\n", err)
		}
	}

	result := map[string]interface{}{
		dictRecord.KeyName:   req.Key,
		dictRecord.ValueName: req.Value,
	}
	middleware.ResponseSuccess(w, "创建字典项成功", result)
}
