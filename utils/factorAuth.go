package utils

import (
	"cmdb/pkg/middleware"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 常量定义
const (
	// OTP默认配置
	OTPDigits    = 6      // OTP验证码位数
	OTPPeriod    = 30     // OTP有效期(秒)
	OTPAlgorithm = "SHA1" // OTP使用的哈希算法

	// 应用信息
	IssuerName = "CMDB系统" // 发行方名称
)

// GenerateSecret 生成随机密钥
func GenerateSecret(length int) (string, error) {
	// 生成随机字节
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 转换为Base32编码
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secret := encoder.EncodeToString(b)
	// 确保Secret是大写的
	return strings.ToUpper(secret), nil
}

// GenerateOTP 生成OTP密钥和二维码URL
func GenerateOTP(username string) (secret string, qrCode string, err error) {
	middleware.Logger.Infof("[双因子验证] 为用户 %s 生成OTP密钥", username)

	// 生成16字节(128位)的随机密钥
	secret, err = GenerateSecret(16)
	if err != nil {
		middleware.Logger.Errorf("[双因子验证] 生成密钥失败: %v", err)
		return "", "", fmt.Errorf("生成密钥失败: %v", err)
	}

	// 生成TOTP URL (用于生成QR码)
	qrCode, err = GenerateTOTPURL(IssuerName, username, secret)
	if err != nil {
		middleware.Logger.Errorf("[双因子验证] 生成TOTP URL失败: %v", err)
		return "", "", err
	}

	middleware.Logger.Infof("[双因子验证] 为用户 %s 生成OTP密钥成功", username)
	return secret, qrCode, nil
}

// GenerateTOTPURL 生成TOTP URL
func GenerateTOTPURL(issuer, account, secret string) (string, error) {
	// URL编码issuer和account
	issuerEncoded := url.QueryEscape(issuer)
	accountEncoded := url.QueryEscape(account)

	// 构建otpauth URL
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("/%s:%s", issuerEncoded, accountEncoded),
	}

	// 添加参数
	q := u.Query()
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", OTPAlgorithm)
	q.Set("digits", fmt.Sprintf("%d", OTPDigits))
	q.Set("period", fmt.Sprintf("%d", OTPPeriod))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ValidateOTP 验证OTP验证码
func ValidateOTP(secret string, code string) bool {
	middleware.Logger.Infof("[双因子验证] 验证OTP码")

	// 解码Secret
	decoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	secretBytes, err := decoder.DecodeString(strings.ToUpper(secret))
	if err != nil {
		middleware.Logger.Errorf("[双因子验证] 解码密钥失败: %v", err)
		return false
	}

	// 获取当前时间戳
	now := time.Now().UTC().Unix()

	// 检查前后时间窗口的码是否匹配 (-1, 0, +1)
	// 这样可以容忍用户和服务器之间存在轻微的时间偏差
	for i := -1; i <= 1; i++ {
		// 计算时间窗口
		timeWindow := int64(math.Floor(float64(now)/float64(OTPPeriod))) + int64(i)

		// 生成验证码
		generatedCode := generateTOTP(secretBytes, timeWindow)

		// 比较验证码
		if generatedCode == code {
			middleware.Logger.Infof("[双因子验证] OTP验证成功")
			return true
		}
	}

	middleware.Logger.Infof("[双因子验证] OTP验证失败")
	return false
}

// generateTOTP 生成TOTP码
func generateTOTP(secret []byte, timeWindow int64) string {
	// 将时间窗口转换为字节数组
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeWindow))

	// 使用HMAC-SHA1计算哈希
	h := hmac.New(sha1.New, secret)
	h.Write(timeBytes)
	hash := h.Sum(nil)

	// 动态截断
	offset := hash[19] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	// 生成指定位数的验证码
	mod := uint32(math.Pow10(OTPDigits))
	code := truncatedHash % mod

	// 格式化为固定位数的字符串
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", OTPDigits), code)
}

// GenerateBackupCodes 生成备用恢复码
func GenerateBackupCodes(count int, length int) []string {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// 生成随机字节
		b := make([]byte, length/2) // 每个字节可以生成2个16进制字符
		rand.Read(b)

		// 转换为16进制字符串
		code := fmt.Sprintf("%x", b)

		// 如果长度超过了指定长度，截断
		if len(code) > length {
			code = code[:length]
		}

		// 格式化为更易读的形式，例如: XXXX-XXXX-XXXX-XXXX
		if length >= 16 {
			formattedCode := ""
			for j := 0; j < len(code); j += 4 {
				if j+4 <= len(code) {
					if j > 0 {
						formattedCode += "-"
					}
					formattedCode += code[j : j+4]
				}
			}
			codes[i] = strings.ToUpper(formattedCode)
		} else {
			codes[i] = strings.ToUpper(code)
		}
	}

	return codes
}

// LogFactorAuthActionHTTP 记录双因子验证操作日志
// 从HTTP请求中获取IP和用户代理信息
func LogFactorAuthActionHTTP(userID int64, action, status string, details map[string]interface{}, r *http.Request) {
	// 获取IP地址
	ip := getClientIP(r)

	// 获取User-Agent
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "unknown"
	}

	// 这里应该调用模型的创建方法存储日志
	// 由于这是工具类，不直接操作数据库，所以只记录日志
	detailsStr := ""
	if details != nil {
		detailsJSON, _ := json.Marshal(details)
		detailsStr = string(detailsJSON)
	}

	middleware.Logger.Infof("[双因子验证] 记录操作日志: 用户=%d, 操作=%s, 状态=%s, IP=%s, 详情=%s",
		userID, action, status, ip, detailsStr)
}

// getClientIP 获取客户端真实IP地址
func getClientIP(r *http.Request) string {
	// 尝试从各种头部获取真实IP
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			if ip != "" {
				return ip
			}
		}
	}

	// 如果没有代理头部，使用RemoteAddr
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if ip == "" {
		return "unknown"
	}
	return ip
}
