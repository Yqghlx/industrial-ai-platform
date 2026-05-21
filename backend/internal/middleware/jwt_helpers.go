package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT Token 过期时间常量 - 与 auth_helpers.go 保持一致
const (
	AccessTokenDuration  = 15 * time.Minute   // AccessToken: 15 分钟
	RefreshTokenDuration = 7 * 24 * time.Hour // RefreshToken: 7 天
)

// 不允许默认密钥 - 必须通过配置设置
var jwtSecret []byte
var jwtInitialized bool

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID int, username, role, tenantID string, secret []byte) (string, error) {
	if len(secret) == 0 {
		if !jwtInitialized {
			return "", fmt.Errorf("JWT not initialized - call SetJWTSecret first")
		}
		secret = jwtSecret
	}

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)), // AccessToken: 15 分钟
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "industrial-ai-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseJWT validates and parses a JWT token
func ParseJWT(tokenString string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		if !jwtInitialized {
			return nil, fmt.Errorf("JWT not initialized")
		}
		secret = jwtSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// GenerateRefreshToken creates a refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetJWTSecret returns the configured JWT secret
// Returns nil if not initialized (caller should handle error)
func GetJWTSecret() []byte {
	return jwtSecret
}

// SetJWTSecret sets the JWT secret (must be called before any JWT operations)
func SetJWTSecret(secret string) {
	if secret != "" {
		if len(secret) < 32 {
			fmt.Printf("WARNING: JWT_SECRET length (%d) is below recommended minimum (32)\n", len(secret))
		}
		jwtSecret = []byte(secret)
		jwtInitialized = true
	}
}
