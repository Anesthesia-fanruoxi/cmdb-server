package audit

import (
	"cmdb/pkg/middleware"
	"time"
)

type OperationResponse struct {
	Total      int64          `json:"total"`
	List       []OperationLog `json:"list"`
	Page       int            `json:"page"`       // 当前页
	PageSize   int            `json:"pageSize"`   // 每页大小
	TotalPages int64          `json:"totalPages"` // 总页数
}

// OperationLog 操作日志结构体
type OperationLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null" json:"userId"`
	UserName  string    `gorm:"size:50;not null" json:"userName"`
	Module    string    `gorm:"size:50;not null" json:"module"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Method    string    `gorm:"size:10;not null" json:"method"`
	URL       string    `gorm:"size:255;not null" json:"url"`
	IP        string    `gorm:"size:50;not null" json:"ip"`
	Request   string    `gorm:"type:text" json:"request"`
	Response  string    `gorm:"type:text" json:"response"`
	Status    int       `gorm:"not null" json:"status"`
	ErrorMsg  string    `gorm:"type:text" json:"errorMsg"`
	Duration  int       `gorm:"not null" json:"duration"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 设置表名
func (OperationLog) TableName() string {
	return middleware.TableOperationLog
}

// OperationLogQuery 操作日志查询参数
type OperationLogQuery struct {
	UserName  string    `json:"userName"`
	Module    string    `json:"module"`
	Action    string    `json:"action"`
	Status    *int      `json:"status"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Page      int64     `json:"page"`     // 当前页
	PageSize  int64     `json:"pageSize"` // 每页大小
}
