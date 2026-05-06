package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// IPLocationAudit IP位置查询审计记录
type IPLocationAudit struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	RequestIP string    `json:"request_ip" gorm:"type:varchar(45);not null;index:idx_request_ip"` // 请求来源IP
	QueryIP   string    `json:"query_ip" gorm:"type:varchar(45);not null;index:idx_query_ip"`     // 查询的IP（入参）
	Status    int       `json:"status" gorm:"not null;default:200;index:idx_status"`              // 响应状态 200:成功 400:失败
	CreatedAt time.Time `json:"created_at" gorm:"not null;index:idx_created_at"`                  // 请求时间
	Date      string    `json:"date" gorm:"type:varchar(10);not null;index:idx_date"`             // 日期(YYYYMMDD)便于统计
}

// TableName 指定表名
func (IPLocationAudit) TableName() string {
	return middleware.TableIPLocationAudit
}
