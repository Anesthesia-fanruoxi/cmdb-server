package types

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// Claims 自定义的JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenExpiration token过期时间
const TokenExpiration = 24 * time.Hour
