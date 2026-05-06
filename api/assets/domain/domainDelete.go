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

// DomainDeleteRequest 域名记录删除请求
type DomainDeleteRequest struct {
	RecordId string `json:"record_id"` // 解析记录ID
}

// DomainDeleteResponse 域名记录删除响应
type DomainDeleteResponse struct {
	RecordId string `json:"record_id"` // 被删除的记录ID
}

// DomainDelete 删除域名解析记录
// @Summary 删除域名解析记录
// @Description 根据记录ID删除域名解析记录
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param request body DomainDeleteRequest true "删除域名解析记录请求"
// @Success 200 {object} DomainDeleteResponse
// @Router /api/assets/domain/delete [post]
func DomainDelete(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许POST请求")
		return
	}

	// 解析请求体
	var req DomainDeleteRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求参数失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	// 验证必填参数
	if req.RecordId == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "解析记录ID为必填项")
		return
	}

	// 创建阿里云DNS客户端
	client, err := CreateClient()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建DNS客户端失败: "+err.Error())
		return
	}

	// 创建请求参数
	deleteRequest := &alidns20150109.DeleteDomainRecordRequest{
		RecordId: tea.String(req.RecordId),
	}

	// 运行时选项
	runtime := &util.RuntimeOptions{}

	// 执行请求
	result, err := client.DeleteDomainRecordWithOptions(deleteRequest, runtime)
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

		middleware.ResponseError(w, http.StatusInternalServerError, "删除域名记录失败: "+message)
		return
	}

	// 返回结果
	response := DomainDeleteResponse{
		RecordId: req.RecordId,
	}

	// 当删除成功但没有返回记录ID时，使用请求中的ID
	if result != nil && result.Body != nil && result.Body.RecordId != nil {
		response.RecordId = tea.StringValue(result.Body.RecordId)
	}

	middleware.ResponseSuccess(w, "删除域名记录成功", response)
}
