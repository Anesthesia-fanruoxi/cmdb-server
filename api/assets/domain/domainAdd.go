// This file is auto-generated, don't edit it. Thanks.
package domain

import (
	"encoding/json"
	"fmt"
	"net/http"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"cmdb/pkg/middleware"
	"strings"
)

// DomainAddRequest 域名记录添加请求
type DomainAddRequest struct {
	RR    string `json:"rr"`    // 主机记录，如www
	Type  string `json:"type"`  // 记录类型，如A, CNAME, MX等
	Value string `json:"value"` // 记录值，如IP地址
	TTL   int64  `json:"ttl"`   // TTL值，可选
}

// DomainAddResponse 域名记录添加响应
type DomainAddResponse struct {
	RecordID string `json:"record_id"` // 记录ID
}

// DomainAdd 添加域名解析记录
// @Summary 添加域名解析记录
// @Description 添加域名解析记录
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param request body DomainAddRequest true "添加域名解析记录请求"
// @Success 200 {object} DomainAddResponse
// @Router /api/assets/domain/add [post]
func DomainAdd(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许POST请求")
		return
	}

	// 解析请求体
	var req DomainAddRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	// 验证必填参数
	if req.RR == "" || req.Type == "" || req.Value == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "主机记录(rr)、记录类型(type)和记录值(value)为必填项")
		return
	}

	// 创建阿里云DNS客户端
	client, err := CreateClient()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建DNS客户端失败: "+err.Error())
		return
	}

	// 从配置获取域名
	domainName := GetDomainName()

	// 设置TTL默认值
	ttl := req.TTL
	if ttl <= 0 {
		ttl = 600 // 默认10分钟
	}

	// 创建请求参数
	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String(domainName),
		RR:         tea.String(req.RR),
		Type:       tea.String(req.Type),
		Value:      tea.String(req.Value),
		TTL:        tea.Int64(ttl),
	}

	// 运行时选项
	runtime := &util.RuntimeOptions{}

	// 执行请求
	result, err := client.AddDomainRecordWithOptions(addDomainRecordRequest, runtime)
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

		middleware.ResponseError(w, http.StatusInternalServerError, "添加域名记录失败: "+message)
		return
	}

	// 返回结果
	response := DomainAddResponse{
		RecordID: tea.StringValue(result.Body.RecordId),
	}

	middleware.ResponseSuccess(w, "添加域名记录成功", response)
}
