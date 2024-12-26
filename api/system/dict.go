package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// 验证表名是否合法
func validateTableName(tableName string) bool {
	return strings.HasSuffix(tableName, "_dict")
}

// ListDicts 获取字典列表
func ListDicts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	dictType := query.Get("type")

	// 构建查询
	db := initData.GetDB()
	var dicts []models.DictRecord
	tx := db.Table("dict_records")
	if dictType != "" {
		tx = tx.Where("type = ?", dictType)
	}

	if err := tx.Find(&dicts).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取字典列表失败")
		return
	}

	utils.Success(w, dicts)
}

// CreateDict 创建字典
func CreateDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var dict models.DictRecord
	if err := json.NewDecoder(r.Body).Decode(&dict); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证必填字段
	if dict.DictName == "" || dict.DictTable == "" || dict.KeyName == "" || dict.ValueName == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "字典名称、表名、键名和值名不能为空")
		return
	}

	// 验证表名格式
	if !strings.HasSuffix(dict.DictTable, "_dict") {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名必须以_dict结尾")
		return
	}

	// 创建字典记录
	if err := initData.GetDB().Create(&dict).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "dict_name") {
				utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "字典名称已存在")
			} else if strings.Contains(err.Error(), "table_name") {
				utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名已存在")
			} else {
				utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "字典记录已存在")
			}
		} else {
			utils.Error(w, http.StatusInternalServerError, utils.ERROR, fmt.Sprintf("创建字典失败: %v", err))
		}
		return
	}

	// 创建对应的数据表
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
			%s VARCHAR(64) NOT NULL COMMENT '键',
			%s VARCHAR(64) NOT NULL COMMENT '值',
			created_by BIGINT UNSIGNED NOT NULL COMMENT '创建人ID',
			created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
			PRIMARY KEY (id),
			UNIQUE KEY uk_%s (%s)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='%s'
	`, dict.DictTable, dict.KeyName, dict.ValueName, dict.KeyName, dict.KeyName, dict.DictName)

	if err := initData.GetDB().Exec(createTableSQL).Error; err != nil {
		// 如果创建表失败，删除刚才创建的记录
		initData.GetDB().Delete(&dict)
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, fmt.Sprintf("创建字典表失败: %v", err))
		return
	}

	utils.Success(w, dict)
}

// DeleteDict 删除字典
func DeleteDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req struct {
		ID   uint   `json:"id"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if req.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "字典ID不能为空")
		return
	}

	// 验证表名
	if !validateTableName(req.Type) {
		utils.Error(w, http.StatusForbidden, utils.ERROR, "无权操作该表")
		return
	}

	// 检查字典是否存在
	var dict models.Dict
	if err := initData.GetDB().First(&dict, req.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "字典不存在")
		return
	}

	// 删除字典
	if err := initData.GetDB().Delete(&dict).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除字典失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"dict_id": dict.ID,
	})
}

// QueryDict 查询字典内容
func QueryDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 获取查询参数
	query := r.URL.Query()
	tableName := query.Get("table_name")
	if tableName == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名不能为���")
		return
	}

	// 验证表名
	if !validateTableName(tableName) {
		utils.Error(w, http.StatusForbidden, utils.ERROR, "无权操作该表")
		return
	}

	// 获取字典记录，以获取key_name和value_name
	var dictRecord models.DictRecord
	if err := initData.GetDB().Where("table_name = ?", tableName).First(&dictRecord).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "字典不存在")
		return
	}

	// 使用原生SQL查询
	rows, err := initData.GetDB().Raw(fmt.Sprintf("SELECT %s, %s FROM %s",
		dictRecord.KeyName,
		dictRecord.ValueName,
		tableName,
	)).Rows()
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, fmt.Sprintf("查询%s失败", tableName))
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			utils.Error(w, http.StatusInternalServerError, utils.ERROR, "扫描数据失败")
			return
		}
		items = append(items, map[string]interface{}{
			dictRecord.KeyName:   key,
			dictRecord.ValueName: value,
		})
	}

	utils.Success(w, items)
}

// CreateDictItem 创建字典项
func CreateDictItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req struct {
		TableName string `json:"table_name"`
		Key       string `json:"key"`
		Value     string `json:"value"`
		CreatedBy uint   `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证表名
	if !validateTableName(req.TableName) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名必须以_dict结尾")
		return
	}

	// 验证必填字段
	if req.TableName == "" || req.Key == "" || req.Value == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名、键和值不能为空")
		return
	}

	// 获取字典记录，以获取key_name
	var dictRecord models.DictRecord
	if err := initData.GetDB().Where("table_name = ?", req.TableName).First(&dictRecord).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "字典不存在")
		return
	}

	// 使用原生SQL插入数据
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s, %s, created_by) VALUES (?, ?, ?)",
		req.TableName,
		dictRecord.KeyName,
		dictRecord.ValueName,
	)

	if err := initData.GetDB().Exec(insertSQL, req.Key, req.Value, req.CreatedBy).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "该键已存在")
			return
		}
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建字典项失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		dictRecord.KeyName:   req.Key,
		dictRecord.ValueName: req.Value,
	})
}

// DeleteDictItem 删除字典项
func DeleteDictItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	var req struct {
		TableName string `json:"table_name"`
		Key       string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	// 验证表名
	if !validateTableName(req.TableName) {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名必须以_dict结尾")
		return
	}

	if req.TableName == "" || req.Key == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "表名和键不能为空")
		return
	}

	// 获取字典记录，以获取key_name
	var dictRecord models.DictRecord
	if err := initData.GetDB().Where("table_name = ?", req.TableName).First(&dictRecord).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "字典不存在")
		return
	}

	// 使用原生SQL删除数据
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s = ?",
		req.TableName,
		dictRecord.KeyName,
	)

	result := initData.GetDB().Exec(deleteSQL, req.Key)
	if result.Error != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除字典项失败")
		return
	}

	if result.RowsAffected == 0 {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "字典项不存在")
		return
	}

	utils.Success(w, map[string]interface{}{
		"table_name": req.TableName,
		"key":        req.Key,
	})
}
