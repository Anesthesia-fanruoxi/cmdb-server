package setting

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"time"
)

// GetBasicSetting 获取基本设置
func GetBasicSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var setting system.Setting
	if err := database.DB.Table(middleware.TableSettings).First(&setting).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取系统设置失败")
		return
	}

	middleware.ResponseSuccess(w, "获取成功", setting)
}

// GetPasswordRules 获取密码规则（不需要鉴权）
func GetPasswdSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var setting system.Setting
	if err := database.DB.Table(middleware.TableSettings).First(&setting).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取密码规则失败")
		return
	}

	// 只返回密码规则相关字段
	passwordRules := map[string]interface{}{
		"password_min_length":   setting.PasswordMinLength,
		"password_max_length":   setting.PasswordMaxLength,
		"password_need_number":  setting.PasswordNeedNumber,
		"password_need_letter":  setting.PasswordNeedLetter,
		"password_need_case":    setting.PasswordNeedCase,
		"password_need_special": setting.PasswordNeedSpecial,
	}

	middleware.ResponseSuccess(w, "获取成功", passwordRules)
}

// UpdateBasicSetting 更新基本设置
func UpdateBasicSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求体
	var setting system.Setting
	if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(setting); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 开启事务
	tx := database.DB.Begin()

	// 构建更新数据，使用 map 确保包含所有字段，包括 false 值
	updates := map[string]interface{}{
		"system_name":           setting.SystemName,
		"system_short_name":     setting.SystemShortName,
		"system_url":            setting.SystemUrl,
		"fs_webhook_url":        setting.FeiShuWebhookUrl,
		"password_min_length":   setting.PasswordMinLength,
		"password_max_length":   setting.PasswordMaxLength,
		"password_need_number":  setting.PasswordNeedNumber,
		"password_need_letter":  setting.PasswordNeedLetter,
		"password_need_case":    setting.PasswordNeedCase,
		"password_need_special": setting.PasswordNeedSpecial,
		"login_protection":      setting.LoginProtection,
		"login_fail_count":      setting.LoginFailCount,
		"login_lock_time":       setting.LoginLockTime,
		"default_password":      setting.DefaultPassword,
		"token_expire_time":     setting.TokenExpireTime,
		"updated_at":            time.Now(),
	}

	// 更新设置
	if err := tx.Table(middleware.TableSettings).Where("id = ?", setting.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "更新系统设置失败")
		return
	}

	// 查询更新后的完整设置
	var updatedSetting system.Setting
	if err := tx.Table(middleware.TableSettings).First(&updatedSetting).Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "获取更新后的设置失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		middleware.ResponseError(w, http.StatusInternalServerError, "提交事务失败")
		return
	}

	// 更新密码规则缓存
	middleware.UpdatePasswordRules(
		setting.PasswordMinLength,
		setting.PasswordMaxLength,
		setting.PasswordNeedNumber,
		setting.PasswordNeedLetter,
		setting.PasswordNeedCase,
		setting.PasswordNeedSpecial,
	)

	middleware.ResponseSuccess(w, "更新成功", updatedSetting)
}
