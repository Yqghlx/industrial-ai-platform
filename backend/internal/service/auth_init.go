package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// JWTServiceInterface defines the interface for JWT service
type JWTServiceInterface interface {
	GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error)
	GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error)
	GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error)
	ParseToken(tokenString string) (*Claims, error)
	RefreshAccessToken(refreshTokenString string) (*TokenPair, error)
	RevokeToken(tokenString string) error
	RevokeAllUserTokens(userID int) error
	ValidateToken(tokenString string) (*Claims, error)
	GetSecret() []byte
}

// ============================================
// JWT 配置常量
// =============================================
const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	TokenIssuer          = "industrial-ai-platform"
	MinSecretLength      = 32
)

// 全局 JWT 服务实例 (向后兼容)
var globalJWTService *JWTService

// InitJWT 初始化全局 JWT 服务 (向后兼容)
// Deprecated: Use NewJWTService instead
func InitJWT(secret string) error {
	service, err := NewJWTService(secret)
	if err != nil {
		return err
	}
	globalJWTService = service
	return nil
}

// IsJWTInitialized 检查 JWT 是否已初始化
func IsJWTInitialized() bool {
	return globalJWTService != nil
}

// SetJWTSecret 设置 JWT 密钥 (向后兼容)
func SetJWTSecret(secret string) error {
	if secret == "" {
		return &JWTInitError{Message: "JWT_SECRET is required"}
	}
	// SEC-HIGH-01: 密钥长度验证
	if len(secret) < MinSecretLength {
		return &JWTInitError{
			Message: fmt.Sprintf("JWT_SECRET must be at least %d characters, got %d", MinSecretLength, len(secret)),
		}
	}
	globalJWTService = &JWTService{secret: []byte(secret)}
	return nil
}

// SetRedisClient 设置 Redis 客户端 (向后兼容)
func SetRedisClient(client *redis.Client) {
	if globalJWTService != nil {
		globalJWTService.SetTokenBlacklist(NewHybridTokenBlacklist(client))
	}
}

// SetRedisClientWithFallback 设置 Redis 客户端，支持 fallback
func SetRedisClientWithFallback(client *redis.Client, useHybrid bool) {
	if globalJWTService != nil {
		if useHybrid {
			globalJWTService.SetTokenBlacklist(NewHybridTokenBlacklist(client))
		} else if client != nil {
			globalJWTService.SetTokenBlacklist(NewRedisTokenBlacklist(client))
		} else {
			globalJWTService.SetTokenBlacklist(NewMemoryTokenBlacklist())
		}
	}
}

// SetMemoryTokenBlacklist 设置纯内存 Token 黑名单
func SetMemoryTokenBlacklist() {
	if globalJWTService != nil {
		globalJWTService.SetTokenBlacklist(NewMemoryTokenBlacklist())
	}
}

// SetUserTokenStore 设置用户 Token 存储 (向后兼容)
func SetUserTokenStore(store UserTokenStoreInterface) {
	if globalJWTService != nil {
		globalJWTService.SetUserTokenStore(store)
	}
}

// GetJWTService 获取全局 JWT 服务实例
func GetJWTService() *JWTService {
	return globalJWTService
}

// GenerateAccessToken 全局函数 (向后兼容)
func GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	if globalJWTService == nil {
		return "", "", errors.New("JWT not initialized")
	}
	return globalJWTService.GenerateAccessToken(userID, username, role, tenantID, tokenVersion)
}

// GenerateRefreshToken 全局函数 (向后兼容)
func GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	if globalJWTService == nil {
		return "", "", errors.New("JWT not initialized")
	}
	return globalJWTService.GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
}

// GenerateTokenPair 全局函数 (向后兼容)
func GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized")
	}
	return globalJWTService.GenerateTokenPair(userID, username, role, tenantID, tokenVersion)
}

// ParseToken 全局函数 (向后兼容)
func ParseToken(tokenString string) (*Claims, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized")
	}
	return globalJWTService.ParseToken(tokenString)
}

// RefreshAccessToken 全局函数 (向后兼容)
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized")
	}
	return globalJWTService.RefreshAccessToken(refreshTokenString)
}

// RevokeToken 全局函数 (向后兼容)
func RevokeToken(tokenString string) error {
	if globalJWTService == nil {
		return errors.New("JWT not initialized")
	}
	return globalJWTService.RevokeToken(tokenString)
}

// RevokeAllUserTokens 全局函数 (向后兼容)
func RevokeAllUserTokens(userID int) error {
	if globalJWTService == nil {
		return errors.New("JWT not initialized")
	}
	return globalJWTService.RevokeAllUserTokens(userID)
}

// GenerateToken 旧 API (向后兼容)
func GenerateToken(userID int, username, role string) (string, error) {
	if globalJWTService == nil {
		return "", errors.New("JWT not initialized")
	}
	token, _, err := globalJWTService.GenerateAccessToken(userID, username, role, "", 0)
	return token, err
}