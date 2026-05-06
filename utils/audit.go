package utils

import (
	"bytes"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ResponseWriter 自定义ResponseWriter来捕获响应
type ResponseWriter struct {
	http.ResponseWriter
	Body   *bytes.Buffer
	Status int
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// ParseModuleAndAction 解析模块和操作类型
func ParseModuleAndAction(path string) (action string) {
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return "unknown"
	}
	if parts[len(parts)-1] == "update" {
		action = "更新"
	}
	if parts[len(parts)-1] == "create" {
		action = "创建"
	}
	if parts[len(parts)-1] == "delete" {
		action = "删除"
	}

	return action
}

// FilterSensitiveInfo 过滤敏感信息
func FilterSensitiveInfo(data string) string {
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return data
	}

	sensitiveFields := []string{"password", "token", "secret", "key"}
	for _, field := range sensitiveFields {
		if _, exists := jsonData[field]; exists {
			jsonData[field] = "******"
		}
	}

	result, err := json.Marshal(jsonData)
	if err != nil {
		return data
	}

	return string(result)
}

// SaveOperationLog 异步保存操作日志
func SaveOperationLog(log *system.OperationLog) {
	if err := database.DB.Create(log).Error; err != nil {
		middleware.Logger.Errorf("保存操作日志失败: %v", err)
	}
}

// SaveLoginLog 异步保存登录日志
func SaveLoginLog(log *system.LoginLog) {
	if err := database.DB.Create(log).Error; err != nil {
		middleware.Logger.Errorf("保存登录日志失败: %v", err)
	}
}

// RecordLoginLog 记录登录日志
func RecordLoginLog(r *http.Request, userID int64, userName string, status bool, msg string) {
	// 获取IP
	clientIP := GetClientIP(r)

	country, region, city, district := GetIPLocationDetails(GetClientIP(r))

	// 将bool类型转换为int
	statusInt := 0
	if status {
		statusInt = 1
	}

	log := system.LoginLog{
		UserID:    userID,
		UserName:  userName,
		IP:        clientIP,
		Country:   country,
		Region:    region,
		City:      city,
		District:  district,
		Browser:   GetBrowser(r.UserAgent()),
		OS:        GetOS(r.UserAgent()),
		Status:    statusInt,
		Msg:       msg,
		CreatedAt: time.Now(),
	}
	// 异步保存日志
	go SaveLoginLog(&log)
}
