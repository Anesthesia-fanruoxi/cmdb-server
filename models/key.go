package models

import (
	"cmdb/pkg/middleware"
	"time"
)

// EncryptRequest 加密请求结构体
type EncryptRequest struct {
	Data string `json:"data"` // 要加密的数据
}

// DecryptRequest 解密请求结构体
type DecryptRequest struct {
	EncryptedData string `json:"encryptedData"` // 要解密的数据
}

// BatchDecryptRequest 批量解密请求结构体
type BatchDecryptRequest struct {
	Project string   `json:"project"` // 项目标识
	Data    []string `json:"data"`    // 要解密的数据数组
}

// EncryptResponse 加密响应结构体
type EncryptResponse struct {
	EncryptedData string `json:"encryptedData"`
}

// DecryptResponse 解密响应结构体
type DecryptResponse struct {
	Data string `json:"data"`
}

// BatchDecryptItem 批量解密单项结构体
type BatchDecryptItem struct {
	Encrypted string `json:"encrypted"`
	Decrypted string `json:"decrypted"`
}

// BatchDecryptResponse 批量解密响应结构体
type BatchDecryptResponse []BatchDecryptItem

// EncryptDecryptAudit 加解密操作审计结构体
type EncryptDecryptAudit struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        int       `json:"user_id" gorm:"column:user_id;not null;comment:用户ID"`
	UserName      string    `json:"user_name" gorm:"column:user_name;not null;comment:用户名"`
	OperationType string    `json:"operation_type" gorm:"column:operation_type;not null;comment:操作类型: encrypt/decrypt/batchDecrypt"`
	Project       string    `json:"project" gorm:"column:project;not null;comment:项目标识"`
	BatchCount    int       `json:"batch_count" gorm:"column:batch_count;default:1;comment:批量操作数量"`
	Status        string    `json:"status" gorm:"column:status;default:success;comment:操作状态: success/failed"`
	IPAddress     string    `json:"ip_address" gorm:"column:ip_address;comment:IP地址"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:创建时间"`
}

// TableName 指定表名
func (EncryptDecryptAudit) TableName() string {
	return middleware.TableEncryptDecryptLog
}
