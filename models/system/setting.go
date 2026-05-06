package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// Setting 系统设置
type Setting struct {
	ID                  int64     `json:"id" gorm:"column:id;primary_key;auto_increment"`                               // 主键ID
	SystemName          string    `json:"system_name" gorm:"column:system_name;not null"`                               // 系统名称
	SystemShortName     string    `json:"system_short_name" gorm:"column:system_short_name;not null"`                   // 系统简称
	SystemUrl           string    `json:"system_url" gorm:"column:system_url;not null"`                                 // 系统访问地址
	FeiShuWebhookUrl    string    `json:"fs_webhook_url" gorm:"column:fs_webhook_url;not null"`                         // 飞书webhook地址
	PasswordMinLength   int       `json:"password_min_length" gorm:"column:password_min_length;not null;default:8"`     // 密码最小长度
	PasswordMaxLength   int       `json:"password_max_length" gorm:"column:password_max_length;not null;default:32"`    // 密码最大长度
	PasswordNeedNumber  bool      `json:"password_need_number" gorm:"column:password_need_number;not null;default:1"`   // 密码是否需要数字
	PasswordNeedLetter  bool      `json:"password_need_letter" gorm:"column:password_need_letter;not null;default:1"`   // 密码是否需要字母
	PasswordNeedCase    bool      `json:"password_need_case" gorm:"column:password_need_case;not null;default:1"`       // 密码是否需要大小写
	PasswordNeedSpecial bool      `json:"password_need_special" gorm:"column:password_need_special;not null;default:0"` // 密码是否需要特殊字符
	LoginProtection     bool      `json:"login_protection" gorm:"column:login_protection;not null;default:1"`           // 是否开启登录保护
	LoginFailCount      int       `json:"login_fail_count" gorm:"column:login_fail_count;not null;default:5"`           // 登录失败次数限制
	LoginLockTime       int       `json:"login_lock_time" gorm:"column:login_lock_time;not null;default:30"`            // 登录锁定时间(分钟)
	DefaultPassword     string    `json:"default_password" gorm:"column:default_password;not null;default:'Cmdb@123'"`  // 默认密码
	TokenExpireTime     int       `json:"token_expire_time" gorm:"column:token_expire_time;not null;default:24"`        // token过期时间(小时)
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at;default:null"`                             // 创建时间
	UpdatedAt           time.Time `json:"updated_at" gorm:"column:updated_at;default:null"`                             // 更新时间
}

// TableName 指定表名
func (Setting) TableName() string {
	return middleware.TableSettings
}
