package search

import "fmt"

// extractErrorMessage 从响应中提取错误信息
func extractErrorMessage(result map[string]interface{}, statusCode int) string {
	if msg, ok := result["msg"].(string); ok {
		return msg
	}
	if errStr, ok := result["error"].(string); ok {
		return errStr
	}
	if message, ok := result["message"].(string); ok {
		return message
	}
	return fmt.Sprintf("未知错误，状态码: %d", statusCode)
}
