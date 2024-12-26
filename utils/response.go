package utils

import (
	"encoding/json"
	"net/http"
)

// 定义统一的状态码
const (
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400
	UNAUTHORIZED   = 401
	FORBIDDEN      = 403
	NOT_FOUND      = 404
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// WriteJSON 写入JSON响应
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteError 写入错误响应
func WriteError(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := Response{
		Code:    ERROR,
		Message: message,
	}
	if err != nil {
		resp.Data = err.Error()
	}
	WriteJSON(w, statusCode, resp)
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	resp := Response{
		Code:    SUCCESS,
		Message: "success",
		Data:    data,
	}
	WriteJSON(w, http.StatusOK, resp)
}

// Error 错误响应
func Error(w http.ResponseWriter, statusCode, code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	WriteJSON(w, statusCode, resp)
}
