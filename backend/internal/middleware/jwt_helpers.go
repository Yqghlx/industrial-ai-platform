package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT Token 过期时间常量 - 与 auth_helpers.go 保持一致
const (
	AccessTokenDuration  = 15 * time.Minute   // AccessToken: 15 分钟
	RefreshTokenDuration = 7 * 24 * time.Hour // RefreshToken: 7 天
)

// JWTConfig 封装 JWT 配置和状态
// BE-P3-04: JWT 配置结构体文档
//
// JWTConfig 提供安全的 JWT 密钥管理，支持:
//   - 密钥存储与读取
//   - 并发安全的密钥访问 (使用 RWMutex)
//   - 密钥轮换功能
//
// 使用示例:
//
//	config, err := NewJWTConfig("your-secret-key")
//	if err != nil {
//	    return err
//	}
//	token, err := GenerateJWTWithConfig(1, "user", "admin", "tenant-1", config)
type JWTConfig struct {
	secret []byte
	mu     sync.RWMutex
}

// NewJWTConfig 创建新的 JWT 配置实例
func NewJWTConfig(secret string) (*JWTConfig, error) {
	if secret == "" {
		return nil, errors.New("JWT secret is required")
	}
	// SEC-LOW-02: 不打印密钥长度信息，避免泄露安全配置
	// 密钥长度验证应在配置层面处理，而非通过日志输出
	return &JWTConfig{
		secret: []byte(secret),
	}, nil
}

// GetSecret 获取 JWT 密钥
func (c *JWTConfig) GetSecret() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secret
}

// SetSecret 设置 JWT 密钥（用于轮换密钥）
func (c *JWTConfig) SetSecret(secret string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if secret != "" {
		c.secret = []byte(secret)
	}
}

// globalJWTConfig 全局默认配置（向后兼容）
// Deprecated: Use NewJWTConfig instead
var globalJWTConfig *JWTConfig

// Claims JWT Token 中包含的用户身份信息
// BE-P3-04: Claims 结构体文档
//
// Claims 包含以下字段:
//   - UserID: 用户唯一标识符
//   - Username: 用户名
//   - Role: 用户角色 (admin/operator/viewer)
//   - TenantID: 租户标识符
//   - RegisteredClaims: 标准 JWT claims (过期时间、签发时间等)
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// GenerateJWTWithConfig 使用指定配置创建 JWT token
func GenerateJWTWithConfig(userID int, username, role, tenantID string, config *JWTConfig) (string, error) {
	if config == nil {
		return "", errors.New("JWT config is nil")
	}
	secret := config.GetSecret()
	return generateJWTInternal(userID, username, role, tenantID, secret)
}

// GenerateJWT creates a new JWT token for a user (向后兼容)
// Deprecated: Use GenerateJWTWithConfig instead
func GenerateJWT(userID int, username, role, tenantID string, secret []byte) (string, error) {
	if len(secret) == 0 {
		if globalJWTConfig == nil {
			return "", errors.New("JWT not initialized - call InitJWTConfig first")
		}
		secret = globalJWTConfig.GetSecret()
	}
	return generateJWTInternal(userID, username, role, tenantID, secret)
}

// generateJWTInternal 内部实现
func generateJWTInternal(userID int, username, role, tenantID string, secret []byte) (string, error) {
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

// ParseJWTWithConfig 使用指定配置解析 JWT token
func ParseJWTWithConfig(tokenString string, config *JWTConfig) (*Claims, error) {
	if config == nil {
		return nil, errors.New("JWT config is nil")
	}
	secret := config.GetSecret()
	return parseJWTInternal(tokenString, secret)
}

// ParseJWT validates and parses a JWT token (向后兼容)
// Deprecated: Use ParseJWTWithConfig instead
func ParseJWT(tokenString string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		if globalJWTConfig == nil {
			return nil, errors.New("JWT not initialized")
		}
		secret = globalJWTConfig.GetSecret()
	}
	return parseJWTInternal(tokenString, secret)
}

// parseJWTInternal 内部实现
func parseJWTInternal(tokenString string, secret []byte) (*Claims, error) {
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

// GetJWTSecret returns the configured JWT secret (向后兼容)
// Deprecated: Use JWTConfig.GetSecret instead
func GetJWTSecret() []byte {
	if globalJWTConfig == nil {
		return nil
	}
	return globalJWTConfig.GetSecret()
}

// SetJWTSecret sets the JWT secret (向后兼容)
// Deprecated: Use NewJWTConfig instead
func SetJWTSecret(secret string) {
	if secret != "" {
		globalJWTConfig, _ = NewJWTConfig(secret)
	}
}

// InitJWTConfig 初始化全局 JWT 配置 (向后兼容)
// Deprecated: Use NewJWTConfig instead
func InitJWTConfig(secret string) error {
	config, err := NewJWTConfig(secret)
	if err != nil {
		return err
	}
	globalJWTConfig = config
	return nil
}

// GetGlobalJWTConfig 获取全局 JWT 配置
func GetGlobalJWTConfig() *JWTConfig {
	return globalJWTConfig
}
