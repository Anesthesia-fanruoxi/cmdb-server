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

// DomainStatusUpdateRequest 域名记录状态更新请求
type DomainStatusUpdateRequest struct {
	RecordId string `json:"record_id"` // 解析记录ID
	Status   string `json:"status"`    // 状态，Enable/Disable
}

// DomainUpdateRequest 域名记录内容更新请求
type DomainUpdateRequest struct {
	RecordId string `json:"record_id"` // 解析记录ID
	RR       string `json:"rr"`        // 主机记录
	Type     string `json:"type"`      // 记录类型
	Value    string `json:"value"`     // 记录值
	TTL      int64  `json:"ttl"`       // TTL值，可选
}

// DomainUpdateResponse 域名记录更新响应
type DomainUpdateResponse struct {
	RecordId string `json:"record_id"` // 记录ID
}

// DomainUpdateStatus 更新域名解析记录状态（启用/禁用）
// @Summary 更新域名解析记录状态
// @Description 更新域名解析记录状态（启用/禁用）
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param request body DomainStatusUpdateRequest true "更新域名解析记录状态请求"
// @Success 200 {object} DomainUpdateResponse
// @Router /api/assets/domain/update/status [post]
func DomainUpdateStatus(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许POST请求")
		return
	}

	// 解析请求体
	var req DomainStatusUpdateRequest
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

	if req.Status == "" || (req.Status != "Enable" && req.Status != "Disable") {
		middleware.ResponseError(w, http.StatusBadRequest, "状态必须为Enable或Disable")
		return
	}

	// 创建阿里云DNS客户端
	client, err := CreateClient()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建DNS客户端失败: "+err.Error())
		return
	}

	// 创建请求参数
	statusRequest := &alidns20150109.SetDomainRecordStatusRequest{
		RecordId: tea.String(req.RecordId),
		Status:   tea.String(req.Status),
	}

	// 运行时选项
	runtime := &util.RuntimeOptions{}

	// 执行请求
	result, err := client.SetDomainRecordStatusWithOptions(statusRequest, runtime)
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

		middleware.ResponseError(w, http.StatusInternalServerError, "更新域名记录状态失败: "+message)
		return
	}

	// 返回结果
	response := DomainUpdateResponse{
		RecordId: tea.StringValue(result.Body.RecordId),
	}

	middleware.ResponseSuccess(w, "更新域名记录状态成功", response)
}

// DomainUpdate 更新域名解析记录内容
// @Summary 更新域名解析记录内容
// @Description 更新域名解析记录的主机记录、类型和值
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param request body DomainUpdateRequest true "更新域名解析记录请求"
// @Success 200 {object} DomainUpdateResponse
// @Router /api/assets/domain/update [post]
func DomainUpdate(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只允许POST请求")
		return
	}

	// 解析请求体
	var req DomainUpdateRequest
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

	// 创建请求参数
	updateRequest := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: tea.String(req.RecordId),
		RR:       tea.String(req.RR),
		Type:     tea.String(req.Type),
		Value:    tea.String(req.Value),
	}

	// 设置可选参数
	if req.TTL > 0 {
		updateRequest.TTL = tea.Int64(req.TTL)
	}

	// 运行时选项
	runtime := &util.RuntimeOptions{}

	// 执行请求
	result, err := client.UpdateDomainRecordWithOptions(updateRequest, runtime)
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

		middleware.ResponseError(w, http.StatusInternalServerError, "更新域名记录失败: "+message)
		return
	}

	// 返回结果
	response := DomainUpdateResponse{
		RecordId: tea.StringValue(result.Body.RecordId),
	}

	middleware.ResponseSuccess(w, "更新域名记录成功", response)
}
