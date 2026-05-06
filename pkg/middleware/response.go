package middleware

import (
	"encoding/json"
	"net/http"
)

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorType 错误类型
type ErrorType struct {
	Code    int
	Message string
}

// CustomError 自定义错误
type CustomError struct {
	ErrorType
	Err error
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// ResponseSuccess 成功响应
func ResponseSuccess(w http.ResponseWriter, message string, data interface{}) {
	resp := Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")

	// 使用自定义设置的Encoder，不自动转换数值类型
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(resp); err != nil {
		// 如果编码失败，返回错误信息
		w.WriteHeader(http.StatusInternalServerError)
		ResponseError(w, http.StatusInternalServerError, "响应编码失败")
		Logger.Errorf("响应编码失败: %v", err)
	}
}

// ResponseError 错误响应
func ResponseError(w http.ResponseWriter, code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

// ResponseErrorWithData 带数据的错误响应
func ResponseErrorWithData(w http.ResponseWriter, code int, message string, data interface{}) {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

// ResponseEncrypted 加密响应（用于需要加密返回的接口）
// secret: 加密密钥
func ResponseEncrypted(w http.ResponseWriter, message string, data interface{}, secret string) {
	// 构建标准响应结构
	resp := Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}

	// 序列化响应数据
	responseJSON, err := json.Marshal(resp)
	if err != nil {
		ResponseError(w, http.StatusInternalServerError, "序列化响应数据失败")
		Logger.Errorf("序列化响应失败: %v", err)
		return
	}

	// 加密并压缩响应数据
	encryptedResponse, err := CompressAndEncrypt(responseJSON, secret)
	if err != nil {
		ResponseError(w, http.StatusInternalServerError, "加密响应数据失败")
		Logger.Errorf("加密响应失败: %v", err)
		return
	}

	// 返回加密后的响应 {data: "加密数据"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"data": encryptedResponse,
	})
}
