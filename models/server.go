package models

import "time"

// Server 服务器表
type Server struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Name       string `gorm:"type:varchar(64);not null"`                   // 服务器名称
	IP         string `gorm:"type:varchar(15);uniqueIndex;not null"`       // IP地址
	Port       int    `gorm:"type:int;not null;default:22"`                // SSH端口
	Username   string `gorm:"type:varchar(32);not null"`                   // SSH用户名
	Password   string `gorm:"type:varchar(128)"`                           // SSH密码
	PrivateKey string `gorm:"type:text"`                                   // SSH私钥
	Type       string `gorm:"type:varchar(32);not null"`                   // 服务器类型（如：物理机、虚拟机）
	Status     string `gorm:"type:varchar(32);not null;default:'offline'"` // 服务器状态
	OS         string `gorm:"type:varchar(64)"`                            // 操作系统
	CPU        int    `gorm:"type:int"`                                    // CPU核数
	Memory     int    `gorm:"type:int"`                                    // 内存大小(GB)
	Disk       int    `gorm:"type:int"`                                    // 磁盘大小(GB)
	Comment    string `gorm:"type:varchar(255)"`                           // 备注
}
