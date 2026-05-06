package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

// DeviceBindRequest 绑定设备请求
type DeviceBindRequest struct {
	MachineID string `json:"machine_id" validate:"required"`
	TOTPCode  string `json:"totp_code" validate:"required"`
	Version   string `json:"version" validate:"required"`
}

// DeviceBindResponse 绑定设备响应
type DeviceBindResponse struct {
	DeviceSecret string `json:"device_secret"`
}

// DeviceBind 绑定设备接口
func DeviceBind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "未登录")
		return
	}

	// 解析请求
	var req DeviceBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 查询用户获取OTP密钥
	var user system.User
	if err := database.DB.Where("user_name = ?", userInfo.UserName).First(&user).Error; err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "用户不存在")
		return
	}

	// 验证用户是否启用了双因子认证
	if !user.OtpEnabled {
		middleware.ResponseError(w, http.StatusBadRequest, "请先启用双因子认证")
		return
	}

	// 验证TOTP
	if !utils.ValidateOTP(user.OtpSecret, req.TOTPCode) {
		middleware.ResponseError(w, http.StatusBadRequest, "TOTP验证码错误")
		return
	}

	// 生成设备密钥（64位十六进制，即32字节）
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成设备密钥失败")
		return
	}
	deviceSecret := hex.EncodeToString(secretBytes)

	// 设置凭据有效期为一个月
	expireAt := time.Now().AddDate(0, 1, 0)

	// 按 user_name 唯一，一个用户只能绑定一个设备，存在则更新，不存在则插入
	device := system.UserDevice{
		UserID:       user.ID,
		UserName:     userInfo.UserName,
		MachineID:    req.MachineID,
		DeviceSecret: deviceSecret,
		Version:      req.Version,
		ExpireAt:     &expireAt,
		Status:       1,
	}
	if err := database.DB.Where("user_name = ?", userInfo.UserName).
		Assign(device).FirstOrCreate(&device).Error; err != nil {
		middleware.Logger.Errorf("[设备绑定] 绑定失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "绑定设备失败")
		return
	}

	middleware.Logger.Infof("[设备绑定] 用户 %s 绑定设备成功，MachineID: %s", userInfo.UserName, req.MachineID)

	middleware.ResponseSuccess(w, "绑定成功", DeviceBindResponse{
		DeviceSecret: deviceSecret,
	})
}
