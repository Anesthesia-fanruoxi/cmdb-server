package search

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// HandleSLSPagedSearch 处理SLS分页查询
// 直接调用SLS API进行查询，而非从缓存获取
func HandleSLSPagedSearch(
	pageReq ESSearchPageRequest,
	project, indexPattern, startTime, endTime, keyword string,
) (*ESSearchResult, error) {
	// 验证参数
	if pageReq.QueryID == "" {
		return nil, fmt.Errorf("查询ID不能为空")
	}

	// 如果关键参数缺失，则无法进行查询
	if project == "" || indexPattern == "" || startTime == "" || endTime == "" {
		return nil, fmt.Errorf("查询参数不完整，无法执行SLS分页查询")
	}

	// 设置默认分页参数
	if pageReq.Page <= 0 {
		pageReq.Page = 1
	}

	// SLS固定页大小为100条
	// SLS API限制每次最多返回100条记录
	pageSize := 20

	// 记录日志
	middleware.Logger.Infof("执行SLS直接查询分页: QID=%s, Project=%s, Page=%d, Size=%d",
		pageReq.QueryID, project, pageReq.Page, pageSize)

	// 解析时间字符串
	parsedStartTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %v", err)
	}

	parsedEndTime, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("解析结束时间失败: %v", err)
	}

	// 计算偏移量
	offset := int64((pageReq.Page - 1) * pageSize)

	// 使用SLSManager获取对应项目的客户端
	client, err := GlobalSLSManager.GetClient(project)
	if err != nil {
		return nil, fmt.Errorf("获取SLS客户端失败: %v", err)
	}

	// 构建查询条件
	query := keyword
	if query == "" {
		query = "*" // 如果没有关键词，查询所有日志
	}

	// 解析日志库名
	logstore := indexPattern
	if strings.HasSuffix(logstore, "*") {
		logstore = logstore[:len(logstore)-1]
	}

	// 获取符合条件的总数
	totalHits, err := CountSLSLogs(client, project, logstore, parsedStartTime.Unix(), parsedEndTime.Unix(), query)
	if err != nil {
		middleware.Logger.Warnf("获取SLS总数失败: %v", err)
		totalHits = 0
	}

	// 查询日志数据
	getLogsV2Request := &sls20201230.GetLogsV2Request{
		From:    tea.Int32(int32(parsedStartTime.Unix())),
		To:      tea.Int32(int32(parsedEndTime.Unix())),
		Query:   tea.String(query + "|with_pack_meta"),
		Line:    tea.Int64(int64(pageSize)),
		Offset:  tea.Int64(offset),
		Reverse: tea.Bool(true),
	}

	// 执行查询
	runtime := &util.RuntimeOptions{}
	headers := &sls20201230.GetLogsV2Headers{}
	response, err := client.GetLogsV2WithOptions(
		tea.String(project),
		tea.String(logstore),
		getLogsV2Request,
		headers,
		runtime,
	)
	if err != nil {
		return nil, fmt.Errorf("查询SLS日志失败: %v", err)
	}

	// 构建初始结果对象
	esResult := &ESSearchResult{
		Project:   project,
		IndexName: indexPattern,
		TotalHits: totalHits,
		Page:      pageReq.Page,
		Size:      pageSize,
		QueryID:   pageReq.QueryID,
		Hits:      []map[string]interface{}{},
	}

	// 计算总页数
	esResult.Pages = (int(totalHits) + pageSize - 1) / pageSize

	// 解析返回的日志数据
	if response != nil && response.Body != nil {
		bodyBytes, err := json.Marshal(response.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化SLS响应失败: %v", err)
		}

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			return nil, fmt.Errorf("解析SLS响应失败: %v", err)
		}

		if dataVal, ok := bodyMap["data"]; ok && dataVal != nil {
			if dataArray, ok := dataVal.([]interface{}); ok {
				for _, item := range dataArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						esResult.Hits = append(esResult.Hits, itemMap)
					}
				}
				middleware.Logger.Infof("SLS分页查询返回日志: 当前页%d条, 偏移量%d", len(dataArray), offset)
			}
		}
	}

	middleware.Logger.Infof("SLS分页查询完成: 总数:%d, 页大小:%d, 偏移量:%d, 返回结果数:%d",
		totalHits, pageSize, offset, len(esResult.Hits))

	return esResult, nil
}
