package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// UserDevice 用户设备绑定表
type UserDevice struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64      `json:"user_id" gorm:"column:user_id;not null;comment:用户ID"`
	UserName     string     `json:"user_name" gorm:"column:user_name;size:64;not null;comment:用户名"`
	MachineID    string     `json:"machine_id" gorm:"column:machine_id;size:128;not null;comment:设备硬件指纹"`
	DeviceSecret string     `json:"-" gorm:"column:device_secret;size:128;not null;comment:设备密钥"`
	DeviceName   string     `json:"device_name" gorm:"column:device_name;size:64;comment:设备名称"`
	Version      string     `json:"version" gorm:"column:version;size:32;comment:客户端版本号"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"column:last_login_at;comment:最后登录时间"`
	ExpireAt     *time.Time `json:"expire_at" gorm:"column:expire_at;comment:凭据到期时间"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:绑定时间"`
	Status       int8       `json:"status" gorm:"column:status;default:1;comment:状态:1-正常,0-禁用"`
}

// TableName 指定表名
func (UserDevice) TableName() string {
	return middleware.TableUserDevice
}
