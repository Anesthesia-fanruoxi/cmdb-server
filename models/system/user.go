package system

import (
	"cmdb/pkg/middleware"
	"time"
)

// User 用户表
type User struct {
	ID                 int64     `json:"id" gorm:"primarykey"`
	RoleID             int64     `json:"role_id" gorm:"comment:角色ID"`
	IsEnabled          bool      `json:"is_enabled" gorm:"default:true;comment:是否启用"`
	IsDefaultPass      bool      `json:"is_default_pass" gorm:"default:true;comment:是否是默认密码"`
	OnlineAssets       bool      `json:"online_assets" gorm:"default:false;comment:是否开启生成资产关联"`
	TestAssets         bool      `json:"test_assets" gorm:"default:false;comment:是否开启测试资产关联"`
	UserName           string    `json:"user_name" gorm:"size:32;uniqueIndex;comment:用户名"`
	Password           string    `json:"-" gorm:"size:128;comment:密码"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	NickName           string    `json:"nick_name" gorm:"size:32;comment:昵称"`
	Email              string    `json:"email" gorm:"size:64;comment:邮箱"`
	Phone              string    `json:"phone" gorm:"size:11;comment:手机号"`
	DeptID             int64     `json:"dept_id" gorm:"comment:部门ID"`
	OtpSecret          string    `json:"otp_secret" gorm:"size:64;comment:双因子认证密钥"`
	AllowPasswordLogin bool      `json:"allow_password_login" gorm:"default:true;comment:是否允许密码登录"`
	OtpEnabled         bool      `json:"otp_enabled" gorm:"default:false;comment:是否启用双因子认证"`
	OtpSetupCompleted  bool      `json:"otp_setup_completed" gorm:"default:false;comment:是否完成双因子认证设置"`
	OtpBackupCodes     string    `json:"otp_backup_codes" gorm:"type:text;comment:双因子认证备用恢复码"`
	Role               Role      `json:"role" gorm:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return middleware.TableUsers
}
