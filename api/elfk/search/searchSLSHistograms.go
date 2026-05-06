package search

import (
	"encoding/json"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"strconv"
)

// SLSProjectInfo SLS项目信息
type SLSProjectInfo struct {
	Project  string // SLS项目名称
	Endpoint string // SLS服务端点
}

// CountSLSLogs 根据查询条件计算SLS日志总数
func CountSLSLogs(client *sls20201230.Client, project, logstore string, from, to int64, query string) (int64, error) {
	// 构造直方图请求
	getHistogramsRequest := &sls20201230.GetHistogramsRequest{
		From:  tea.Int64(from),
		To:    tea.Int64(to),
		Query: tea.String(query),
	}

	// 执行查询
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	response, err := client.GetHistogramsWithOptions(
		tea.String(project),
		tea.String(logstore),
		getHistogramsRequest,
		headers,
		runtime,
	)
	if err != nil {
		return 0, err
	}

	// 将响应体转换为JSON字符串
	bodyBytes, err := json.Marshal(response.Body)
	if err != nil {
		return 0, err
	}

	// 解析为JSON数组
	var histogramArray []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &histogramArray); err != nil {
		return 0, err
	}

	// 计算总数
	var totalCount int64 = 0
	for _, hist := range histogramArray {
		if countVal, ok := hist["count"]; ok {
			var count int64
			switch v := countVal.(type) {
			case string:
				count, _ = strconv.ParseInt(v, 10, 64)
			case float64:
				count = int64(v)
			case json.Number:
				count, _ = v.Int64()
			}
			totalCount += count
		}
	}

	return totalCount, nil
}
