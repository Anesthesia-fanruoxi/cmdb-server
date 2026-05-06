package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DeviceLoginRequest 设备登录请求
type DeviceLoginRequest struct {
	UserName  string `json:"user_name" validate:"required"`
	MachineID string `json:"machine_id" validate:"required"`
	Challenge string `json:"challenge" validate:"required"`
	Timestamp int64  `json:"timestamp" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	Version   string `json:"version" validate:"required"`
}

// DeviceLoginResponse 设备登录响应
type DeviceLoginResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// DeviceLogin 设备登录接口
func DeviceLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req DeviceLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	ctx := context.Background()

	// 1. 验证挑战码
	challengeKey := middleware.GetDeviceChallengeKey(req.Challenge)
	challengeJSON, err := database.RDB.Get(ctx, challengeKey).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "挑战码无效或已过期")
		return
	}

	var challengeInfo ChallengeInfo
	if err := json.Unmarshal([]byte(challengeJSON), &challengeInfo); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析挑战码失败")
		return
	}

	// 2. 验证挑战码对应的用户和设备
	if challengeInfo.UserName != req.UserName || challengeInfo.MachineID != req.MachineID {
		middleware.ResponseError(w, http.StatusBadRequest, "挑战码不匹配")
		return
	}

	// 3. 验证时间戳（允许±60秒误差）
	currentTime := time.Now().Unix()
	if abs(currentTime-req.Timestamp) > 60 {
		middleware.ResponseError(w, http.StatusBadRequest, "时间戳超出有效范围")
		return
	}

	// 4. 验证用户绑定的设备
	var device system.UserDevice
	// 先查询用户是否有绑定的设备
	var userDevice system.UserDevice
	err = database.DB.Where("user_name = ?", req.UserName).First(&userDevice).Error
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "用户未绑定任何设备")
		return
	}

	// 验证用户是否使用其绑定的设备（一个用户只能绑定一个设备）
	if userDevice.MachineID != req.MachineID {
		middleware.ResponseError(w, http.StatusBadRequest, "请使用已绑定的设备登录")
		return
	}

	// 验证设备状态
	if userDevice.Status != 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "设备已被禁用")
		return
	}
	device = userDevice

	// 4.2 验证版本号（暂时取消版本验证）
	// if "v0.0.1" != req.Version {
	// 	middleware.ResponseError(w, http.StatusBadRequest, "客户端版本过低，请升级后重新绑定设备")
	// 	return
	// }

	// 5. 验证签名
	expectedSignature := generateSignature(req.Challenge, req.Timestamp, device.DeviceSecret)
	if !hmac.Equal([]byte(req.Signature), []byte(expectedSignature)) {
		// 记录失败次数
		failKey := middleware.GetDeviceLoginFailKey(req.MachineID)
		failCount, _ := database.RDB.Incr(ctx, failKey).Result()
		database.RDB.Expire(ctx, failKey, 5*time.Minute)

		if failCount >= 5 {
			// 禁用设备
			database.DB.Model(&system.UserDevice{}).
				Where("machine_id = ?", req.MachineID).
				Update("status", 0)
			middleware.ResponseError(w, http.StatusBadRequest, "签名验证失败次数过多，设备已被禁用")
			return
		}
		middleware.ResponseError(w, http.StatusBadRequest, "签名验证失败")
		return
	}

	// 6. 删除已使用的挑战码（一次性）
	database.RDB.Del(ctx, challengeKey)

	// 7. 清除失败计数
	failKey := middleware.GetDeviceLoginFailKey(req.MachineID)
	database.RDB.Del(ctx, failKey)

	// 8. 更新最后登录时间
	now := time.Now()
	database.DB.Model(&device).Update("last_login_at", now)

	// 9. 获取用户信息，生成Token
	var user system.User
	if err := database.DB.Where("user_name = ?", req.UserName).First(&user).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// 检查用户状态
	if !user.IsEnabled {
		middleware.ResponseError(w, http.StatusForbidden, "用户已被禁用")
		return
	}

	// 获取角色权限
	permissions, err := utils.GetRolePermissions(uint(user.RoleID))
	if err != nil {
		if err := utils.RefreshRolePermissions(uint(user.RoleID)); err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "刷新权限缓存失败")
			return
		}
		permissions, _ = utils.GetRolePermissions(uint(user.RoleID))
	}
	permissionStr := strings.Join(permissions, ",")

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.UserName, user.RoleID, user.DeptID, permissionStr, user.IsDefaultPass)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成token失败")
		return
	}

	// 存储token
	if err := utils.StoreToken(user.ID, token); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "存储token失败")
		return
	}

	// 记录登录日志
	utils.RecordLoginLog(r, user.ID, user.UserName, true, "设备自动登录成功")

	middleware.Logger.Infof("[设备登录] 用户 %s 通过设备自动登录成功", req.UserName)

	middleware.ResponseSuccess(w, "登录成功", DeviceLoginResponse{
		Token:    token,
		UserID:   fmt.Sprintf("%d", user.ID),
		UserName: user.UserName,
	})
}

// generateSignature 生成HMAC-SHA256签名
func generateSignature(challenge string, timestamp int64, deviceSecret string) string {
	message := fmt.Sprintf("%s:%d", challenge, timestamp)
	h := hmac.New(sha256.New, []byte(deviceSecret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// abs 返回绝对值
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
