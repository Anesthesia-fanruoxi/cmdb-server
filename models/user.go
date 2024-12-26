package models

import (
	"cmdb/config"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var cfg *config.Config

// SetConfig 设置配置
func SetConfig(c *config.Config) {
	cfg = c
}

type User struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string `gorm:"type:varchar(32);uniqueIndex;not null"`
	Password  string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string `gorm:"type:varchar(32)"`
	Email     string `gorm:"type:varchar(128)"`
	Phone     string `gorm:"type:varchar(11)"`
	RoleID    uint   `gorm:"not null;default:2"`
	IsEnabled bool   `gorm:"not null;default:true"`
	DeptID    uint   `gorm:"not null" json:"dept_id"`
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	// 将密码和盐值组合
	saltedPassword := password + cfg.Security.PasswordSalt

	// 使用bcrypt加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	if cfg == nil {
		log.Printf("警告: 配置未初始化，使用空盐值")
		// 使用bcrypt验证密码（不加盐）
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		return err == nil
	}

	// 将密码和盐值组合
	saltedPassword := password + cfg.Security.PasswordSalt

	// 使用bcrypt验证密码
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(saltedPassword))
	return err == nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
