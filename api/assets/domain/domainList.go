package domain

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"net/http"
	"strconv"
	"strings"
)

// DomainListResponse 域名记录列表响应
type DomainListResponse struct {
	Total int                      `json:"total"`
	List  []map[string]interface{} `json:"list"`
}

// DomainList 获取域名解析列表
// @Summary 获取域名解析列表
// @Description 获取域名解析列表，支持分页和过滤条件
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param rrKeyWord query string false "主机记录关键字"
// @Param typeKeyWord query string false "记录类型关键字"
// @Param valueKeyWord query string false "记录值关键字"
// @Param status query string false "记录状态，如ENABLE, DISABLE"
// @Success 200 {object} DomainListResponse
// @Router /api/v1/assets/domain/list [get]
func DomainList(w http.ResponseWriter, r *http.Request) {
	// 获取页码参数
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 获取过滤条件
	rrKeyWord := r.URL.Query().Get("rrKeyWord")
	typeKeyWord := r.URL.Query().Get("typeKeyWord")
	valueKeyWord := r.URL.Query().Get("valueKeyWord")
	status := r.URL.Query().Get("status")

	// 固定每页大小为20条
	pageSize := 20

	// 创建Alibaba Cloud客户端
	client, err := CreateClient()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建客户端失败: "+err.Error())
		return
	}

	// 从manager获取域名
	domainName := GetDomainName()

	// 创建请求参数
	request := &alidns20150109.DescribeDomainRecordsRequest{
		DomainName: tea.String(domainName),
		PageNumber: tea.Int64(int64(page)),
		PageSize:   tea.Int64(int64(pageSize)),
	}

	// 设置过滤条件
	if rrKeyWord != "" {
		request.RRKeyWord = tea.String(rrKeyWord)
	}
	if typeKeyWord != "" {
		request.TypeKeyWord = tea.String(typeKeyWord)
	}
	if valueKeyWord != "" {
		request.ValueKeyWord = tea.String(valueKeyWord)
	}
	if status != "" {
		request.Status = tea.String(status)
	}

	// 运行时选项
	runtime := &util.RuntimeOptions{}

	// 执行请求
	result, err := client.DescribeDomainRecordsWithOptions(request, runtime)
	if err != nil {
		// 处理错误
		var message string
		sdkErr, ok := err.(*tea.SDKError)
		if ok {
			message = tea.StringValue(sdkErr.Message)

			// 尝试解析错误数据，获取推荐信息
			var data interface{}
			if sdkErr.Data != nil {
				d := json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
				if d.Decode(&data) == nil {
					if m, ok := data.(map[string]interface{}); ok {
						if recommend, exists := m["Recommend"]; exists {
							message += fmt.Sprintf(", 推荐: %v", recommend)
						}
					}
				}
			}
		} else {
			message = err.Error()
		}

		middleware.ResponseError(w, http.StatusInternalServerError, "获取域名记录失败: "+message)
		return
	}

	// 准备响应数据
	recordList := make([]map[string]interface{}, 0)
	total := 0

	if result != nil && result.Body != nil {
		// 获取总记录数
		total = int(tea.Int64Value(result.Body.TotalCount))

		// 转换记录列表
		if result.Body.DomainRecords != nil && result.Body.DomainRecords.Record != nil {
			for _, record := range result.Body.DomainRecords.Record {
				item := map[string]interface{}{
					"recordId":        tea.StringValue(record.RecordId),
					"rr":              tea.StringValue(record.RR),
					"type":            tea.StringValue(record.Type),
					"value":           tea.StringValue(record.Value),
					"ttl":             tea.Int64Value(record.TTL),
					"status":          tea.StringValue(record.Status),
					"domainName":      tea.StringValue(record.DomainName),
					"createTimestamp": tea.Int64Value(record.CreateTimestamp),
					"updateTimestamp": tea.Int64Value(record.UpdateTimestamp),
				}
				recordList = append(recordList, item)
			}
		}
	}

	// 返回结果
	response := DomainListResponse{
		Total: total,
		List:  recordList,
	}
	middleware.ResponseSuccess(w, "获取域名记录成功", response)
}
