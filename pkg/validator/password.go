package validator

import (
	"fmt"
	"unicode"
)

// PasswordRules 密码规则
type PasswordRules struct {
	MinLength   int  // 最小长度
	MaxLength   int  // 最大长度
	NeedNumber  bool // 是否需要数字
	NeedLetter  bool // 是否需要字母
	NeedCase    bool // 是否需要大小写
	NeedSpecial bool // 是否需要特殊字符
}

// ValidatePassword 验证密码复杂度
func ValidatePassword(password string, rules PasswordRules) error {
	// 检查长度
	length := len(password)
	if length < rules.MinLength || length > rules.MaxLength {
		return fmt.Errorf("密码长度必须在%d-%d之间", rules.MinLength, rules.MaxLength)
	}

	if rules.NeedNumber {
		hasNumber := false
		for _, char := range password {
			if unicode.IsNumber(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return fmt.Errorf("密码必须包含数字")
		}
	}

	if rules.NeedLetter {
		hasLetter := false
		for _, char := range password {
			if unicode.IsLetter(char) {
				hasLetter = true
				break
			}
		}
		if !hasLetter {
			return fmt.Errorf("密码必须包含字母")
		}
	}

	if rules.NeedCase {
		hasUpper := false
		hasLower := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
			}
			if unicode.IsLower(char) {
				hasLower = true
			}
		}
		if !hasUpper || !hasLower {
			return fmt.Errorf("密码必须包含大小写字母")
		}
	}

	if rules.NeedSpecial {
		hasSpecial := false
		for _, char := range password {
			if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			return fmt.Errorf("密码必须包含特殊字符")
		}
	}

	return nil
}
