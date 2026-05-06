package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

// DeviceChallengeRequest 获取挑战码请求
type DeviceChallengeRequest struct {
	UserName  string `json:"user_name" validate:"required"`
	MachineID string `json:"machine_id" validate:"required"`
}

// DeviceChallengeResponse 获取挑战码响应
type DeviceChallengeResponse struct {
	Challenge string `json:"challenge"`
	ExpiresAt int64  `json:"expires_at"`
}

// ChallengeInfo 挑战码信息（存储在Redis中）
type ChallengeInfo struct {
	UserName  string `json:"user_name"`
	MachineID string `json:"machine_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// DeviceChallenge 获取挑战码接口
func DeviceChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求
	var req DeviceChallengeRequest
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

	// 频率限制检查（同一设备每分钟最多5次）
	rateKey := middleware.GetDeviceChallengeRateKey(req.MachineID)
	count, err := database.RDB.Incr(ctx, rateKey).Result()
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "系统错误")
		return
	}
	if count == 1 {
		database.RDB.Expire(ctx, rateKey, 60*time.Second)
	}
	if count > 5 {
		middleware.ResponseError(w, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
		return
	}

	// 检查设备是否已绑定
	var device system.UserDevice
	err = database.DB.Where("user_name = ? AND machine_id = ? AND status = 1",
		req.UserName, req.MachineID).First(&device).Error
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "设备未绑定")
		return
	}

	// 生成挑战码（32位十六进制，即16字节）
	challengeBytes := make([]byte, 16)
	if _, err := rand.Read(challengeBytes); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成挑战码失败")
		return
	}
	challenge := hex.EncodeToString(challengeBytes)
	expiresAt := time.Now().Unix() + 30 // 30秒后过期

	// 存储挑战码到Redis
	challengeInfo := ChallengeInfo{
		UserName:  req.UserName,
		MachineID: req.MachineID,
		ExpiresAt: expiresAt,
	}
	infoJSON, _ := json.Marshal(challengeInfo)
	challengeKey := middleware.GetDeviceChallengeKey(challenge)
	if err := database.RDB.Set(ctx, challengeKey, string(infoJSON), 30*time.Second).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "存储挑战码失败")
		return
	}

	middleware.ResponseSuccess(w, "success", DeviceChallengeResponse{
		Challenge: challenge,
		ExpiresAt: expiresAt,
	})
}
