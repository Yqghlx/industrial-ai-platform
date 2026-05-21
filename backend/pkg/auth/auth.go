// Package auth provides authentication and authorization utilities
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a random token string
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ParseClaims extracts claims from a JWT token
// FIX-007: 添加算法验证防止算法混淆攻击
func ParseClaims(tokenString string, secretKey []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// FIX-007: 验证算法类型 - 只允许 HMAC 算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// Reject "none" algorithm and other non-HMAC algorithms
			return nil, ErrInvalidToken
		}
		// Specifically reject "none" algorithm (belt and suspenders)
		if token.Method.Alg() == "none" {
			return nil, ErrInvalidToken
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// FIX-007: 验证必要的 claims
		if claims.UserID == 0 || claims.Username == "" {
			return nil, ErrInvalidToken
		}
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(claims *Claims) bool {
	return claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now())
}
