package utils

import (
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/validator"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword 加密密码
func EncryptPassword(password string) (string, error) {
	cfg := config.GetConfig().Auth

	// 先用盐值加密
	saltedPass := password + cfg.PasswordSalt
	hash := sha256.New()
	hash.Write([]byte(saltedPass))
	hashedPass := hex.EncodeToString(hash.Sum(nil))

	// 再用bcrypt加密
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(hashedPass), cfg.BcryptCost)
	if err != nil {
		return "", err
	}

	return string(bcryptHash), nil
}

// ComparePassword 比较密码
func ComparePassword(hashedPassword, password string) bool {
	// 先用盐值加密
	cfg := config.GetConfig().Auth
	saltedPass := password + cfg.PasswordSalt
	hash := sha256.New()
	hash.Write([]byte(saltedPass))
	hashedPass := hex.EncodeToString(hash.Sum(nil))

	// 再用bcrypt比较
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(hashedPass))
	return err == nil
}

// ValidatePassword 验证密码复杂度
func ValidatePassword(password string) error {
	// 获取系统设置
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		return fmt.Errorf("获取系统设置失败: %v", err)
	}

	rules := validator.PasswordRules{
		MinLength:   setting.PasswordMinLength,
		MaxLength:   setting.PasswordMaxLength,
		NeedNumber:  setting.PasswordNeedNumber,
		NeedLetter:  setting.PasswordNeedLetter,
		NeedCase:    setting.PasswordNeedCase,
		NeedSpecial: setting.PasswordNeedSpecial,
	}

	return validator.ValidatePassword(password, rules)
}
