package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// OperationLog 操作日志
type OperationLog struct {
	ID        int64     `json:"id" gorm:"primaryKey;column:id;comment:日志ID"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;comment:用户ID"`
	UserName  string    `json:"user_name" gorm:"column:user_name;comment:用户名"`
	Module    string    `json:"module" gorm:"column:module;comment:操作模块"`
	Action    string    `json:"action" gorm:"column:action;comment:操作类型"`
	Method    string    `json:"method" gorm:"column:method;comment:请求方法"`
	URL       string    `json:"url" gorm:"column:url;comment:请求URL"`
	IP        string    `json:"ip" gorm:"column:ip;comment:请求IP"`
	Country   string    `json:"country" gorm:"column:country;comment:国家"`
	Region    string    `json:"region" gorm:"column:region;comment:省份"`
	City      string    `json:"city" gorm:"column:city;comment:城市"`
	District  string    `json:"district" gorm:"column:district;comment:区县"`
	Request   string    `json:"request" gorm:"column:request;type:text;comment:请求参数"`
	Response  string    `json:"response" gorm:"column:response;type:text;comment:响应结果"`
	Status    int       `json:"status" gorm:"column:status;comment:响应状态码"`
	ErrorMsg  string    `json:"error_msg" gorm:"column:error_msg;type:text;comment:错误信息"`
	Duration  int       `json:"duration" gorm:"column:duration;comment:执行时长(ms)"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return middleware.TableOperationLog
}

// LoginLog 登录日志
type LoginLog struct {
	ID        int64     `json:"id" gorm:"primaryKey;column:id;comment:日志ID"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;comment:用户ID"`
	UserName  string    `json:"user_name" gorm:"column:user_name;comment:用户名"`
	IP        string    `json:"ip" gorm:"column:ip;comment:登录IP"`
	Country   string    `json:"country" gorm:"column:country;comment:国家"`
	Region    string    `json:"region" gorm:"column:region;comment:省份"`
	City      string    `json:"city" gorm:"column:city;comment:城市"`
	District  string    `json:"district" gorm:"column:district;comment:区县"`
	Browser   string    `json:"browser" gorm:"column:browser;comment:浏览器类型"`
	OS        string    `json:"os" gorm:"column:os;comment:操作系统"`
	Status    int       `json:"status" gorm:"column:status;comment:登录状态(0:失败,1:成功)"`
	Msg       string    `json:"msg" gorm:"column:msg;comment:提示消息"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:登录时间"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return middleware.TableLoginLog
}
