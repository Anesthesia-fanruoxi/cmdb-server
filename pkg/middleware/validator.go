package middleware

import (
	"cmdb/pkg/database"
	"cmdb/pkg/validator"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	playgroundValidator "github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"strconv"
	"strings"
)

// RequestValidator 请求验证器接口
type RequestValidator interface {
	Validate() error
}

// Validator 全局验证器实例
var Validator *playgroundValidator.Validate

var (
	Trans ut.Translator
)

// InitValidator 初始化验证器
func InitValidator() {
	Validator = playgroundValidator.New()

	// 注册密码验证函数
	Validator.RegisterValidation("password", validatePassword)

	// 注册翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	var ok bool
	Trans, ok = uni.GetTranslator("zh")
	if !ok {
		Logger.Warn("获取中文翻译器失败")
	}

	// 注册翻译
	if err := zh_translations.RegisterDefaultTranslations(Validator, Trans); err != nil {
		Logger.Error("注册默认翻译失败: %v\n", err)
	}

	// 注册密码验证错误信息翻译
	if err := Validator.RegisterTranslation("password", Trans, func(ut ut.Translator) error {
		return ut.Add("password", "{0}不符合密码规则", true)
	}, func(ut ut.Translator, fe playgroundValidator.FieldError) string {
		t, _ := ut.T("password", fe.Field())
		return t
	}); err != nil {
		Logger.Errorf("注册密码翻译失败: %v", err)
	}

	// 加载密码规则
	LoadPasswordRulesFromDB()
}

// 密码规则缓存
var cachedPasswordRules validator.PasswordRules

// UpdatePasswordRules 更新密码规则
func UpdatePasswordRules(minLength, maxLength int, needNumber, needLetter, needCase, needSpecial bool) {
	cachedPasswordRules = validator.PasswordRules{
		MinLength:   minLength,
		MaxLength:   maxLength,
		NeedNumber:  needNumber,
		NeedLetter:  needLetter,
		NeedCase:    needCase,
		NeedSpecial: needSpecial,
	}
	Logger.Infof("密码规则已更新: 长度(%d-%d), 需要数字:%v, 需要字母:%v, 需要大小写:%v, 需要特殊字符:%v",
		minLength, maxLength, needNumber, needLetter, needCase, needSpecial)
}

// 存储密码验证失败的错误信息
var passwordErrors = make(map[string]string)

// validatePassword 密码验证函数
func validatePassword(fl playgroundValidator.FieldLevel) bool {
	password := fl.Field().String()
	fieldName := fl.FieldName()
	Logger.Infof("开始验证密码字段 [%s]，密码长度: %d", fieldName, len(password))

	// 确保密码规则已加载
	if cachedPasswordRules.MinLength == 0 {
		Logger.Info("密码规则未加载，尝试从数据库加载")
		// 如果规则未加载，尝试从数据库加载
		LoadPasswordRulesFromDB()

		// 如果仍未加载成功，使用安全的默认规则
		if cachedPasswordRules.MinLength == 0 {
			Logger.Warn("无法从数据库加载密码规则，使用默认规则")
			cachedPasswordRules = validator.PasswordRules{
				MinLength:   8,
				MaxLength:   32,
				NeedNumber:  true,
				NeedLetter:  true,
				NeedCase:    false, // 默认不要求大小写
				NeedSpecial: false,
			}
		}
	}

	Logger.Infof("当前密码规则: 长度(%d-%d), 需要数字:%v, 需要字母:%v, 需要大小写:%v, 需要特殊字符:%v",
		cachedPasswordRules.MinLength, cachedPasswordRules.MaxLength,
		cachedPasswordRules.NeedNumber, cachedPasswordRules.NeedLetter,
		cachedPasswordRules.NeedCase, cachedPasswordRules.NeedSpecial)

	// 使用缓存的规则进行验证
	err := validator.ValidatePassword(password, cachedPasswordRules)
	if err != nil {
		Logger.Warnf("密码验证失败 [%s]: %v", fieldName, err)
		// 存储具体的错误信息
		passwordErrors[fieldName] = err.Error()
		return false
	}

	// 验证成功时删除错误信息
	delete(passwordErrors, fieldName)
	Logger.Infof("密码验证成功 [%s]", fieldName)
	return true
}

// LoadPasswordRulesFromDB 从数据库加载密码规则
func LoadPasswordRulesFromDB() {
	Logger.Info("开始从数据库加载密码规则")

	var setting struct {
		PasswordMinLength   int  `json:"password_min_length"`
		PasswordMaxLength   int  `json:"password_max_length"`
		PasswordNeedNumber  bool `json:"password_need_number"`
		PasswordNeedLetter  bool `json:"password_need_letter"`
		PasswordNeedCase    bool `json:"password_need_case"`
		PasswordNeedSpecial bool `json:"password_need_special"`
	}

	if err := database.DB.Table(TableSettings).Select(
		"password_min_length, password_max_length, password_need_number, password_need_letter, password_need_case, password_need_special",
	).First(&setting).Error; err != nil {
		Logger.Errorf("从数据库加载密码规则失败: %v", err)
		return
	}

	Logger.Infof("从数据库加载的密码规则: 长度(%d-%d), 需要数字:%v, 需要字母:%v, 需要大小写:%v, 需要特殊字符:%v",
		setting.PasswordMinLength, setting.PasswordMaxLength,
		setting.PasswordNeedNumber, setting.PasswordNeedLetter,
		setting.PasswordNeedCase, setting.PasswordNeedSpecial)

	UpdatePasswordRules(
		setting.PasswordMinLength,
		setting.PasswordMaxLength,
		setting.PasswordNeedNumber,
		setting.PasswordNeedLetter,
		setting.PasswordNeedCase,
		setting.PasswordNeedSpecial,
	)

	Logger.Info("密码规则加载成功并已更新")
}

// generatePasswordErrorMessage 生成密码验证错误消息
func generatePasswordErrorMessage() string {
	var requirements []string

	// 添加长度要求
	requirements = append(requirements, "长度"+strconv.Itoa(cachedPasswordRules.MinLength)+"-"+
		strconv.Itoa(cachedPasswordRules.MaxLength)+"位")

	// 添加其他规则要求
	if cachedPasswordRules.NeedNumber {
		requirements = append(requirements, "包含数字")
	}
	if cachedPasswordRules.NeedLetter {
		requirements = append(requirements, "包含字母")
	}
	if cachedPasswordRules.NeedCase {
		requirements = append(requirements, "同时包含大小写字母")
	}
	if cachedPasswordRules.NeedSpecial {
		requirements = append(requirements, "包含特殊字符")
	}

	return "必须" + strings.Join(requirements, "、")
}

// TranslateError 翻译验证错误
func TranslateError(err error) string {
	if validationErrors, ok := err.(playgroundValidator.ValidationErrors); ok {
		var errMsgs []string
		for _, e := range validationErrors {
			// 获取字段名称
			field := e.Field()
			// 根据不同的验证标签生成错误消息
			switch e.Tag() {
			case "required":
				errMsgs = append(errMsgs, field+"不能为空")
			case "min":
				errMsgs = append(errMsgs, field+"长度不能小于"+e.Param())
			case "max":
				errMsgs = append(errMsgs, field+"长度不能大于"+e.Param())
			case "email":
				errMsgs = append(errMsgs, field+"必须是有效的邮箱地址")
			case "len":
				errMsgs = append(errMsgs, field+"长度必须是"+e.Param()+"位")
			case "password":
				// 查找具体的密码错误信息
				if specificError, exists := passwordErrors[field]; exists {
					errMsgs = append(errMsgs, field+specificError)
				} else {
					// 使用默认错误信息
					errMsgs = append(errMsgs, field+generatePasswordErrorMessage())
				}
			default:
				errMsgs = append(errMsgs, field+"格式不正确")
			}
		}
		return strings.Join(errMsgs, "; ")
	}
	return err.Error()
}
