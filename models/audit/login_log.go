package audit

import (
	"cmdb/pkg/middleware"
	"time"
)

// LoginLog 登录日志结构体
type LoginLog struct {
	ID        int64     `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null" json:"userId"`
	UserName  string    `gorm:"column:user_name;size:50;not null" json:"userName"`
	IP        string    `gorm:"column:ip;size:50;not null" json:"ip"`
	City      string    `gorm:"column:city;size:50;not null" json:"city"`
	Country   string    `gorm:"column:country;size:50;not null" json:"country"`
	Region    string    `gorm:"column:region;size:50;not null" json:"region"`
	District  string    `gorm:"column:district;size:50;not null" json:"district"`
	Browser   string    `gorm:"column:browser;size:50" json:"browser"`
	OS        string    `gorm:"column:os;size:50" json:"os"`
	Status    int       `gorm:"column:status;not null" json:"status"` // 0: 失败, 1: 成功
	Msg       string    `gorm:"column:msg;size:255" json:"msg"`       // 提示消息
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName 设置表名
func (LoginLog) TableName() string {
	return middleware.TableLoginLog
}

type LoginResponse struct {
	Total      int64      `json:"total"`
	List       []LoginLog `json:"list"`
	Page       int        `json:"page"`       // 当前页
	PageSize   int        `json:"pageSize"`   // 每页大小
	TotalPages int64      `json:"totalPages"` // 总页数
}

// LoginLogQuery 操作日志查询参数
type LoginLogQuery struct {
	UserName  string    `json:"userName"`
	Status    *int      `json:"status"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Page      int64     `json:"page"`     // 当前页
	PageSize  int64     `json:"pageSize"` // 每页大小
}
