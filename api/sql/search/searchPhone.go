package search

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/sql"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// SearchPhoneRequest 手机号查询请求
type SearchPhoneRequest struct {
	Project        string   `json:"project"`         // 项目名称
	SqlSearch      string   `json:"sql_search"`      // SQL查询语句
	DecryptColumns []string `json:"decrypt_columns"` // 需要解密的列名列表（可选，不传则默认解密第一列）
}

// SearchPhoneResponse 手机号查询响应（字符串数组）
type SearchPhoneResponse []string

// SearchPhone 查询并解密手机号
func SearchPhone(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 获取加密密钥
	cfg := config.GetConfig()
	phoneSecret := cfg.Phone.Secret
	if phoneSecret == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "手机号查询接口加密密钥未配置")
		return
	}

	// 解析加密请求体 {data: "加密数据"}
	var encryptedReq struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &encryptedReq); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败")
		return
	}

	if encryptedReq.Data == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "data字段不能为空")
		return
	}

	// 解密请求数据
	decryptedData, err := middleware.DecryptAndDecompress(encryptedReq.Data, phoneSecret)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("解密请求失败: %v", err))
		return
	}

	middleware.Logger.Infof("解密后的请求数据: %s", string(decryptedData))

	// 解析解密后的请求参数
	var req SearchPhoneRequest
	if err := json.Unmarshal(decryptedData, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析解密后的请求参数失败")
		return
	}

	// 验证必填字段
	if req.Project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project字段不能为空")
		return
	}
	if req.SqlSearch == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "sql_search字段不能为空")
		return
	}

	// 检查是否使用了 SELECT *
	if containsSelectStar(req.SqlSearch) {
		middleware.ResponseError(w, http.StatusBadRequest, "不允许使用 SELECT * 查询，请明确指定需要查询的列")
		return
	}

	// 第1步：执行SQL查询
	middleware.Logger.Infof("开始查询手机号，项目: %s, SQL: %s", req.Project, req.SqlSearch)

	queryResult, err := executePhoneQuery(req.Project, req.SqlSearch)
	if err != nil {
		middleware.Logger.Errorf("查询失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("查询失败: %v", err))
		return
	}

	middleware.Logger.Infof("查询成功，获取到 %d 行数据", len(queryResult.Rows))

	// 第2步：识别并解密手机号字段
	decryptedRows, err := decryptPhoneRows(r, queryResult.Columns, queryResult.Rows, req.Project, req.DecryptColumns)
	if err != nil {
		middleware.Logger.Errorf("解密失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("解密失败: %v", err))
		return
	}

	middleware.Logger.Infof("解密成功，共解密 %d 行数据", len(decryptedRows))

	// 构建响应数据结构
	var responseData interface{}

	if len(req.DecryptColumns) == 0 {
		// 兼容旧逻辑：不传解密列参数时，只返回第一列的字符串数组
		dataList := make(SearchPhoneResponse, 0, len(decryptedRows))
		for _, row := range decryptedRows {
			if len(row) > 0 && row[0] != nil {
				if val, ok := row[0].(string); ok {
					dataList = append(dataList, val)
				}
			}
		}
		responseData = dataList
		middleware.Logger.Infof("旧逻辑：返回第一列数据，共 %d 条", len(dataList))
	} else {
		// 新逻辑：传了解密列参数时，将二维数组转为对象数组返回
		objRows := make([]map[string]interface{}, 0, len(decryptedRows))
		for _, row := range decryptedRows {
			rowMap := make(map[string]interface{}, len(queryResult.Columns))
			for i, col := range queryResult.Columns {
				if i < len(row) {
					rowMap[col] = row[i]
				}
			}
			objRows = append(objRows, rowMap)
		}
		responseData = map[string]interface{}{
			"rows":  objRows,
			"total": len(objRows),
		}
		middleware.Logger.Infof("新逻辑：返回对象数组，列: %v，行数: %d", queryResult.Columns, len(objRows))
	}

	// 记录审计日志（无用户信息，传空context）
	go RecordSqlAudit(r.Clone(context.Background()), req.Project, queryResult.DBName,
		req.SqlSearch, "", time.Since(startTime).Milliseconds(),
		"success", "", "phone_query", 1, "", len(decryptedRows))
	//middleware.ResponseSuccess(w, "查询成功", responseData)
	middleware.ResponseEncrypted(w, "查询成功", responseData, phoneSecret)
}

// executePhoneQuery 执行SQL查询
func executePhoneQuery(project string, sqlQuery string) (*sql.QueryResult, error) {
	// 构建请求数据
	requestData := map[string]interface{}{
		"query":    sqlQuery,
		"dbName":   "",    // 从SQL语句中自动识别
		"is_count": false, // 不需要COUNT
	}

	// 调用 Agent 接口
	pluginPath := "/proxy/sql-plugs/api/searchPhone"
	decryptedData, err := middleware.CallAgentWithBody(project, pluginPath, requestData)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %v", err)
	}

	// 解析响应
	decoder := json.NewDecoder(bytes.NewReader(decryptedData))
	decoder.UseNumber()

	var agentResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Results []struct {
				Columns []string        `json:"columns"`
				Rows    [][]interface{} `json:"rows"`
			} `json:"results"`
		} `json:"data"`
	}

	if err := decoder.Decode(&agentResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查返回码
	if agentResponse.Code != 200 {
		return nil, fmt.Errorf("Agent返回错误: %s", agentResponse.Msg)
	}

	// 检查结果
	if len(agentResponse.Data.Results) == 0 {
		return &sql.QueryResult{
			Columns: []string{},
			Rows:    [][]interface{}{},
			Total:   0,
		}, nil
	}

	// 提取第一个结果
	firstResult := agentResponse.Data.Results[0]
	return &sql.QueryResult{
		DBName:  "",
		Columns: firstResult.Columns,
		Rows:    firstResult.Rows,
		Total:   int64(len(firstResult.Rows)),
	}, nil
}

// decryptPhoneRows 对查询结果的指定列进行批量解密
func decryptPhoneRows(r *http.Request, columns []string, rows [][]interface{}, project string, decryptColumns []string) ([][]interface{}, error) {
	if len(rows) == 0 {
		return rows, nil
	}

	// 确定需要解密的列索引
	var columnIndexes []int
	if len(decryptColumns) == 0 {
		// 兼容旧逻辑：不传参数则默认解密第一列
		columnIndexes = []int{0}
		middleware.Logger.Info("未指定解密列，默认解密第一列")
	} else {
		// 根据列名找到对应的列索引
		columnMap := make(map[string]int)
		for i, col := range columns {
			columnMap[col] = i
		}

		for _, colName := range decryptColumns {
			if idx, exists := columnMap[colName]; exists {
				columnIndexes = append(columnIndexes, idx)
			} else {
				middleware.Logger.Warnf("列名 '%s' 不存在，跳过", colName)
			}
		}

		if len(columnIndexes) == 0 {
			middleware.Logger.Warn("没有找到需要解密的列")
			return rows, nil
		}

		middleware.Logger.Infof("需要解密的列: %v (索引: %v)", decryptColumns, columnIndexes)
	}

	// 1. 收集所有需要解密的列的数据（去重）
	encryptedData := make(map[string]bool)
	for _, row := range rows {
		for _, colIdx := range columnIndexes {
			if colIdx < len(row) && row[colIdx] != nil {
				if data, ok := row[colIdx].(string); ok && data != "" {
					encryptedData[data] = true
				}
			}
		}
	}

	if len(encryptedData) == 0 {
		middleware.Logger.Info("没有数据需要解密")
		return rows, nil
	}

	middleware.Logger.Infof("共收集到 %d 个待解密的唯一数据", len(encryptedData))

	// 2. 批量解密（每批10000条）
	dataList := make([]string, 0, len(encryptedData))
	for data := range encryptedData {
		dataList = append(dataList, data)
	}

	decryptedMap, err := batchDecryptPhones(r, dataList, project)
	if err != nil {
		return nil, fmt.Errorf("批量解密失败: %v", err)
	}

	middleware.Logger.Infof("成功解密 %d 个数据", len(decryptedMap))

	// 3. 替换指定列的加密数据
	decryptedRows := make([][]interface{}, len(rows))
	for i, row := range rows {
		newRow := make([]interface{}, len(row))
		copy(newRow, row)

		// 解密指定的列
		for _, colIdx := range columnIndexes {
			if colIdx < len(newRow) && newRow[colIdx] != nil {
				if encryptedValue, ok := newRow[colIdx].(string); ok {
					if decryptedValue, exists := decryptedMap[encryptedValue]; exists {
						newRow[colIdx] = decryptedValue
					}
				}
			}
		}

		decryptedRows[i] = newRow
	}

	return decryptedRows, nil
}

// batchDecryptPhones 批量解密手机号（处理批量限制）
func batchDecryptPhones(r *http.Request, phones []string, project string) (map[string]string, error) {
	result := make(map[string]string)

	// 分批处理，每批最多10000条
	batchSize := 10000
	for i := 0; i < len(phones); i += batchSize {
		end := i + batchSize
		if end > len(phones) {
			end = len(phones)
		}

		batch := phones[i:end]
		middleware.Logger.Infof("解密第 %d 批，共 %d 条", i/batchSize+1, len(batch))

		// 调用批量解密接口
		decrypted, err := callBatchDecryptAPI(r, batch, project)
		if err != nil {
			return nil, fmt.Errorf("解密第 %d 批失败: %v", i/batchSize+1, err)
		}

		// 合并结果
		for k, v := range decrypted {
			result[k] = v
		}
	}

	return result, nil
}

// callBatchDecryptAPI 调用agent批量解密接口
func callBatchDecryptAPI(_ *http.Request, phones []string, project string) (map[string]string, error) {
	// 调用 Agent 接口
	pluginPath := "/proxy/key-plugs/batchDecrypt"
	decryptedData, err := middleware.CallAgentWithBody(project, pluginPath, phones)
	if err != nil {
		return nil, fmt.Errorf("调用Agent失败: %v", err)
	}

	// 解析批量解密结果
	var batchResults []map[string]interface{}
	if err := json.Unmarshal(decryptedData, &batchResults); err != nil {
		// 尝试二次解析
		var stringResult string
		if err2 := json.Unmarshal(decryptedData, &stringResult); err2 != nil {
			return nil, fmt.Errorf("解析批量解密结果失败: %v", err)
		}
		if err3 := json.Unmarshal([]byte(stringResult), &batchResults); err3 != nil {
			return nil, fmt.Errorf("二次解析批量解密结果失败: %v", err3)
		}
	}

	// 构建加密->解密映射
	decryptedMap := make(map[string]string)
	for _, item := range batchResults {
		encrypted, okEncrypted := item["encrypted"].(string)
		decrypted, okDecrypted := item["decrypted"].(string)

		if okEncrypted && okDecrypted {
			decryptedMap[encrypted] = decrypted
		}
	}

	return decryptedMap, nil
}

// containsSelectStar 检查 SQL 是否包含 SELECT *
func containsSelectStar(sqlQuery string) bool {
	// 转为大写并去除多余空格
	upperSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))

	// 使用正则表达式匹配 SELECT * 模式
	// 匹配: SELECT * 或 SELECT  * 或 SELECT	* (多个空格/制表符)
	// 但不匹配: SELECT *abc 或 SELECT a.* 等
	selectStarPattern := `SELECT\s+\*(\s|,|FROM|$)`
	matched, _ := regexp.MatchString(selectStarPattern, upperSQL)

	return matched
}
