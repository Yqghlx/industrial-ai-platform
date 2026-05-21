package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// JWT 配置常量
const (
	AccessTokenDuration  = 15 * time.Minute   // 短期访问 Token (15 分钟)
	RefreshTokenDuration = 7 * 24 * time.Hour // 刷新 Token (7 天)
	TokenIssuer          = "industrial-ai-platform"
	BlacklistPrefix      = "jwt_blacklist:" // Token 黑名单 Redis key 前缀
	MinSecretLength      = 32               // JWT 密钥最小长度
)

// JWTInitError JWT 初始化错误
type JWTInitError struct {
	Message string
}

func (e *JWTInitError) Error() string {
	return e.Message
}

// TokenBlacklistInterface Token 黑名单接口
type TokenBlacklistInterface interface {
	Add(ctx context.Context, tokenID string, duration time.Duration) error
	Exists(ctx context.Context, tokenID string) bool
	// AddUserRevocation 记录用户撤销时间戳（用于撤销该用户的所有 Token）
	AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error
	// GetUserRevocation 获取用户撤销时间戳（返回零值表示未被撤销）
	GetUserRevocation(ctx context.Context, userID int) (time.Time, error)
}

// UserTokenStoreInterface 用户 Token 版本存储接口
// 用于验证 Token 版本和更新版本号
type UserTokenStoreInterface interface {
	GetTokenVersion(ctx context.Context, userID int) (int, error)
	UpdateTokenVersion(ctx context.Context, userID int) error
}

// RedisTokenBlacklist Redis 实现的 Token 黑名单
type RedisTokenBlacklist struct {
	client *redis.Client
}

func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	return b.client.Set(ctx, BlacklistPrefix+tokenID, "1", duration).Err()
}

func (b *RedisTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	val, err := b.client.Exists(ctx, BlacklistPrefix+tokenID).Result()
	return err == nil && val > 0
}

// UserRevocationPrefix 用户撤销记录前缀
const UserRevocationPrefix = "user_revoke:"

// AddUserRevocation 记录用户撤销时间戳
func (b *RedisTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	key := fmt.Sprintf("%s%d", UserRevocationPrefix, userID)
	// 存储 ISO 格式的时间戳
	return b.client.Set(ctx, key, revokedAt.Format(time.RFC3339), duration).Err()
}

// GetUserRevocation 获取用户撤销时间戳
func (b *RedisTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	key := fmt.Sprintf("%s%d", UserRevocationPrefix, userID)
	val, err := b.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return time.Time{}, nil // 未找到记录，返回零值
		}
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

// MemoryTokenBlacklist 内存实现的 Token 黑名单
// FIX-041: Redis 不可用时的内存 fallback
// FIX-001: 添加 shutdown channel 解决 Goroutine 泄漏
type MemoryTokenBlacklist struct {
	entries         map[string]time.Time    // tokenID -> 过期时间
	userRevocations map[int]revocationEntry // userID -> 撤销记录
	mu              sync.RWMutex
	shutdown        chan struct{} // FIX-001: 用于停止清理 goroutine
}

// revocationEntry 用户撤销记录
type revocationEntry struct {
	revokedAt time.Time // 撤销时间
	expiresAt time.Time // 记录过期时间
}

// NewMemoryTokenBlacklist 创建内存 Token 黑名单
// FIX-001: 初始化 shutdown channel
func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	bl := &MemoryTokenBlacklist{
		entries:         make(map[string]time.Time),
		userRevocations: make(map[int]revocationEntry),
		shutdown:        make(chan struct{}), // FIX-001: 创建 shutdown channel
	}
	// 启动后台清理任务
	go bl.cleanupExpiredEntries()
	return bl
}

// Add 将 Token 加入黑名单
func (b *MemoryTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries[BlacklistPrefix+tokenID] = time.Now().Add(duration)
	return nil
}

// Exists 检查 Token 是否在黑名单中
func (b *MemoryTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	expiry, exists := b.entries[BlacklistPrefix+tokenID]
	if !exists {
		return false
	}

	// 检查是否已过期
	if time.Now().After(expiry) {
		return false // 已过期，视为不存在
	}

	return true
}

// AddUserRevocation 记录用户撤销时间戳
func (b *MemoryTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.userRevocations[userID] = revocationEntry{
		revokedAt: revokedAt,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

// GetUserRevocation 获取用户撤销时间戳
func (b *MemoryTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entry, exists := b.userRevocations[userID]
	if !exists {
		return time.Time{}, nil // 未找到记录，返回零值
	}

	// 检查是否已过期
	if time.Now().After(entry.expiresAt) {
		return time.Time{}, nil // 已过期，视为不存在
	}

	return entry.revokedAt, nil
}

// cleanupExpiredEntries 定期清理过期条目
// FIX-001: 添加 shutdown 监听，解决 Goroutine 泄漏
func (b *MemoryTokenBlacklist) cleanupExpiredEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			now := time.Now()
			// 清理过期的 Token 黑名单
			for tokenID, expiry := range b.entries {
				if now.After(expiry) {
					delete(b.entries, tokenID)
				}
			}
			// 清理过期的用户撤销记录
			for userID, entry := range b.userRevocations {
				if now.After(entry.expiresAt) {
					delete(b.userRevocations, userID)
				}
			}
			b.mu.Unlock()
		case <-b.shutdown:
			// FIX-001: 收到停止信号，退出 goroutine
			return
		}
	}
}

// Stop 停止清理 goroutine
// FIX-001: 新增 Stop 方法
func (b *MemoryTokenBlacklist) Stop() {
	close(b.shutdown)
}

// Size 返回当前黑名单大小
func (b *MemoryTokenBlacklist) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries)
}

// HybridTokenBlacklist 混合实现的 Token 黑名单
// FIX-041: 优先使用 Redis，fallback 到内存
// FIX-002: 添加 shutdown channel 解决 Goroutine 泄漏
type HybridTokenBlacklist struct {
	redisBlacklist  *RedisTokenBlacklist
	memoryBlacklist *MemoryTokenBlacklist
	useRedis        bool // 当前是否使用 Redis
	checkInterval   time.Duration
	shutdown        chan struct{} // FIX-002: 用于停止健康检查 goroutine
}

// NewHybridTokenBlacklist 创建混合 Token 黑名单
// FIX-041: 支持 Redis fallback
// FIX-002: 初始化 shutdown channel
func NewHybridTokenBlacklist(redisClient *redis.Client) *HybridTokenBlacklist {
	hb := &HybridTokenBlacklist{
		memoryBlacklist: NewMemoryTokenBlacklist(),
		checkInterval:   30 * time.Second,
		shutdown:        make(chan struct{}), // FIX-002: 创建 shutdown channel
	}

	if redisClient != nil {
		hb.redisBlacklist = NewRedisTokenBlacklist(redisClient)
		hb.useRedis = true
		// 启动 Redis 健康检查
		go hb.checkRedisHealth(redisClient)
	}

	return hb
}

// Add 将 Token 加入黑名单
func (b *HybridTokenBlacklist) Add(ctx context.Context, tokenID string, duration time.Duration) error {
	// 先尝试写入内存（保证 fallback 可用）
	b.memoryBlacklist.Add(ctx, tokenID, duration)

	// 如果 Redis 可用，同步写入 Redis
	if b.useRedis && b.redisBlacklist != nil {
		err := b.redisBlacklist.Add(ctx, tokenID, duration)
		if err != nil {
			// Redis 写入失败，记录日志但不返回错误（内存已写入）
			fmt.Printf("Warning: Redis blacklist write failed, using memory fallback: %v\n", err)
			b.useRedis = false
		}
	}

	return nil
}

// Exists 检查 Token 是否在黑名单中
func (b *HybridTokenBlacklist) Exists(ctx context.Context, tokenID string) bool {
	// 优先检查 Redis
	if b.useRedis && b.redisBlacklist != nil {
		if b.redisBlacklist.Exists(ctx, tokenID) {
			return true
		}
	}

	// Fallback 到内存检查
	return b.memoryBlacklist.Exists(ctx, tokenID)
}

// checkRedisHealth 定期检查 Redis 健康状态
// FIX-002: 添加 shutdown 监听，解决 Goroutine 泄漏
func (b *HybridTokenBlacklist) checkRedisHealth(client *redis.Client) {
	ticker := time.NewTicker(b.checkInterval)
	defer ticker.Stop()

	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			// 尝试 Ping Redis
			err := client.Ping(ctx).Err()
			if err != nil {
				if b.useRedis {
					fmt.Printf("Warning: Redis unavailable, switching to memory blacklist: %v\n", err)
					b.useRedis = false
				}
			} else {
				if !b.useRedis {
					fmt.Printf("Info: Redis recovered, switching back to Redis blacklist\n")
					b.useRedis = true
				}
			}
		case <-b.shutdown:
			// FIX-002: 收到停止信号，退出 goroutine
			return
		}
	}
}

// Stop 停止健康检查 goroutine
// FIX-002: 新增 Stop 方法
func (b *HybridTokenBlacklist) Stop() {
	close(b.shutdown)
	// 同时停止内存黑名单的清理 goroutine
	if b.memoryBlacklist != nil {
		b.memoryBlacklist.Stop()
	}
}

// IsUsingRedis 返回当前是否使用 Redis
func (b *HybridTokenBlacklist) IsUsingRedis() bool {
	return b.useRedis
}

// AddUserRevocation 记录用户撤销时间戳
func (b *HybridTokenBlacklist) AddUserRevocation(ctx context.Context, userID int, revokedAt time.Time, duration time.Duration) error {
	// 先写入内存（保证 fallback 可用）
	b.memoryBlacklist.AddUserRevocation(ctx, userID, revokedAt, duration)

	// 如果 Redis 可用，同步写入 Redis
	if b.useRedis && b.redisBlacklist != nil {
		err := b.redisBlacklist.AddUserRevocation(ctx, userID, revokedAt, duration)
		if err != nil {
			// Redis 写入失败，记录日志但不返回错误（内存已写入）
			fmt.Printf("Warning: Redis user revocation write failed, using memory fallback: %v\n", err)
		}
	}

	return nil
}

// GetUserRevocation 获取用户撤销时间戳
func (b *HybridTokenBlacklist) GetUserRevocation(ctx context.Context, userID int) (time.Time, error) {
	// 优先检查 Redis
	if b.useRedis && b.redisBlacklist != nil {
		revokedAt, err := b.redisBlacklist.GetUserRevocation(ctx, userID)
		if err == nil && !revokedAt.IsZero() {
			return revokedAt, nil
		}
	}

	// Fallback 到内存检查
	return b.memoryBlacklist.GetUserRevocation(ctx, userID)
}

// Claims represents JWT claims (增强版)
type Claims struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	TenantID     string `json:"tenant_id"`     // 多租户支持
	TokenType    string `json:"token_type"`    // "access" 或 "refresh"
	TokenID      string `json:"token_id"`      // Token 唯一标识 (用于黑名单)
	TokenVersion int    `json:"token_version"` // Token 版本号 (用于撤销)
	jwt.RegisteredClaims
}

// TokenPair 包含 Access Token 和 Refresh Token
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access Token 过期时间 (秒)
	TokenType    string `json:"token_type"` // "Bearer"
}

// JWTService JWT 服务结构体 - 封装所有 JWT 相关功能
// FIX-037: 移除全局变量，改用 Service 结构体封装
type JWTService struct {
	secret         []byte
	tokenBlacklist TokenBlacklistInterface
	userTokenStore UserTokenStoreInterface
}

// NewJWTService 创建新的 JWT 服务实例
// 必须传入有效的 secret，否则返回错误
func NewJWTService(secret string) (*JWTService, error) {
	if secret == "" {
		return nil, &JWTInitError{Message: "JWT_SECRET is required. Please set the JWT_SECRET environment variable."}
	}

	// 密钥强度验证
	if len(secret) < MinSecretLength {
		return nil, &JWTInitError{
			Message: fmt.Sprintf("JWT_SECRET must be at least %d characters for security. Current length: %d", MinSecretLength, len(secret)),
		}
	}

	return &JWTService{
		secret: []byte(secret),
	}, nil
}

// SetTokenBlacklist 设置 Token 黑名单
func (s *JWTService) SetTokenBlacklist(blacklist TokenBlacklistInterface) {
	s.tokenBlacklist = blacklist
}

// SetUserTokenStore 设置用户 Token 版本存储
func (s *JWTService) SetUserTokenStore(store UserTokenStoreInterface) {
	s.userTokenStore = store
}

// GetSecret 获取 JWT 密钥 (用于签名验证)
func (s *JWTService) GetSecret() []byte {
	return s.secret
}

// GenerateAccessToken generates a short-lived access token (15 minutes)
func (s *JWTService) GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	tokenID := generateTokenID()

	claims := Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		TenantID:     tenantID,
		TokenType:    "access",
		TokenID:      tokenID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	return tokenString, tokenID, err
}

// GenerateRefreshToken generates a long-lived refresh token (7 days)
func (s *JWTService) GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	tokenID := generateTokenID()

	claims := Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		TenantID:     tenantID,
		TokenType:    "refresh",
		TokenID:      tokenID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	return tokenString, tokenID, err
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error) {
	accessToken, _, err := s.GenerateAccessToken(userID, username, role, tenantID, tokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ParseToken parses and validates a JWT token
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		ctx := context.Background()

		// 1. 检查 Token 是否在黑名单中（单个 Token 撤销）
		if s.tokenBlacklist != nil && s.tokenBlacklist.Exists(ctx, claims.TokenID) {
			return nil, errors.New("token has been revoked")
		}

		// 2. 检查用户撤销记录（用户级别撤销）
		// 如果用户被撤销，且 Token 签发时间在撤销时间之前，则 Token 无效
		if s.tokenBlacklist != nil {
			revokedAt, err := s.tokenBlacklist.GetUserRevocation(ctx, claims.UserID)
			if err == nil && !revokedAt.IsZero() {
				// Token 的签发时间在撤销时间之前，则 Token 无效
				if claims.IssuedAt != nil && claims.IssuedAt.Before(revokedAt) {
					return nil, errors.New("token has been revoked (user-level revocation)")
				}
			}
		}

		// 3. 验证 Issuer
		if claims.Issuer != TokenIssuer {
			return nil, errors.New("invalid token issuer")
		}

		// 4. 验证 TokenVersion (如果设置了 userTokenStore)
		if s.userTokenStore != nil {
			currentVersion, err := s.userTokenStore.GetTokenVersion(ctx, claims.UserID)
			if err != nil {
				// 如果无法获取版本，记录日志但不阻止验证 (向后兼容)
				fmt.Printf("Warning: failed to get token version for user %d: %v\n", claims.UserID, err)
			} else if claims.TokenVersion != currentVersion {
				return nil, errors.New("token has been revoked (version mismatch)")
			}
		}

		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// RefreshAccessToken 使用 Refresh Token 生成新的 Access Token
func (s *JWTService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := s.ParseToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 验证是否是 refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	// 获取最新的 TokenVersion (如果设置了 userTokenStore)
	tokenVersion := claims.TokenVersion
	if s.userTokenStore != nil {
		currentVersion, err := s.userTokenStore.GetTokenVersion(context.Background(), claims.UserID)
		if err != nil {
			fmt.Printf("Warning: failed to get token version for refresh: %v\n", err)
		} else {
			tokenVersion = currentVersion
		}
	}

	// 生成新的 token pair
	return s.GenerateTokenPair(claims.UserID, claims.Username, claims.Role, claims.TenantID, tokenVersion)
}

// RevokeToken 将 Token 加入黑名单 (用于注销/修改密码)
func (s *JWTService) RevokeToken(tokenString string) error {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return err
	}

	if s.tokenBlacklist == nil {
		return errors.New("token blacklist not initialized")
	}

	// 计算剩余有效期
	remainingTime := time.Until(claims.ExpiresAt.Time)
	if remainingTime <= 0 {
		return nil // Token 已过期，无需加入黑名单
	}

	// 加入黑名单
	return s.tokenBlacklist.Add(context.Background(), claims.TokenID, remainingTime)
}

// RevokeAllUserTokens 撤销用户的所有 Token (修改密码时使用)
// 使用双重机制：1) Token 版本递增（数据库） 2) 用户撤销记录（Redis 黑名单）
func (s *JWTService) RevokeAllUserTokens(userID int) error {
	ctx := context.Background()

	// 1. 在黑名单中记录用户撤销时间戳（用于快速检查）
	// TTL 设置为 RefreshTokenDuration (7 天)，确保所有可能的 Token 都过期
	if s.tokenBlacklist != nil {
		now := time.Now()
		err := s.tokenBlacklist.AddUserRevocation(ctx, userID, now, RefreshTokenDuration)
		if err != nil {
			// 黑名单写入失败，记录日志但继续执行（版本机制仍有效）
			fmt.Printf("Warning: failed to add user revocation to blacklist for user %d: %v\n", userID, err)
		}
	}

	// 2. 递增用户的 TokenVersion（数据库）
	if s.userTokenStore == nil {
		return errors.New("user token store not initialized")
	}

	return s.userTokenStore.UpdateTokenVersion(ctx, userID)
}

// generateTokenID 生成唯一的 Token ID (使用 crypto/rand)
// FIX-036: 使用 crypto/rand 替代时间戳
func generateTokenID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 备用方案：时间戳
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}
	return hex.EncodeToString(bytes)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// === 向后兼容的全局变量和函数 ===
// 这些函数已被标记为 deprecated，建议使用 JWTService 结构体

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

// IsJWTInitialized 检查 JWT 是否已初始化 (向后兼容)
// Deprecated: Check if globalJWTService != nil instead
func IsJWTInitialized() bool {
	return globalJWTService != nil
}

// SetJWTSecret 设置 JWT 密钥 (向后兼容)
// Deprecated: Use NewJWTService or InitJWT instead
func SetJWTSecret(secret string) {
	if secret != "" {
		if len(secret) < MinSecretLength {
			fmt.Printf("WARNING: JWT_SECRET length (%d) is below recommended minimum (32). This may be insecure.\n", len(secret))
		}
		// 创建服务但不返回错误（向后兼容）
		globalJWTService = &JWTService{secret: []byte(secret)}
	}
}

// SetRedisClient 设置 Redis 客户端 (向后兼容)
// Deprecated: Use JWTService.SetTokenBlacklist instead
// FIX-041: 使用 HybridTokenBlacklist 支持 Redis fallback
func SetRedisClient(client *redis.Client) {
	if globalJWTService != nil {
		// 使用混合黑名单（支持 Redis fallback 到内存）
		globalJWTService.SetTokenBlacklist(NewHybridTokenBlacklist(client))
	}
}

// SetRedisClientWithFallback 设置 Redis 客户端，支持 fallback
// FIX-041: 新增函数，允许显式选择黑名单实现
func SetRedisClientWithFallback(client *redis.Client, useHybrid bool) {
	if globalJWTService != nil {
		if useHybrid {
			globalJWTService.SetTokenBlacklist(NewHybridTokenBlacklist(client))
		} else if client != nil {
			globalJWTService.SetTokenBlacklist(NewRedisTokenBlacklist(client))
		} else {
			// 没有 Redis，使用纯内存黑名单
			globalJWTService.SetTokenBlacklist(NewMemoryTokenBlacklist())
		}
	}
}

// SetMemoryTokenBlacklist 设置纯内存 Token 黑名单
// FIX-041: 新增函数，用于无 Redis 环境
func SetMemoryTokenBlacklist() {
	if globalJWTService != nil {
		globalJWTService.SetTokenBlacklist(NewMemoryTokenBlacklist())
	}
}

// SetUserTokenStore 设置用户 Token 存储 (向后兼容)
// Deprecated: Use JWTService.SetUserTokenStore instead
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
// Deprecated: Use JWTService.GenerateAccessToken instead
func GenerateAccessToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	if globalJWTService == nil {
		return "", "", errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.GenerateAccessToken(userID, username, role, tenantID, tokenVersion)
}

// GenerateRefreshToken 全局函数 (向后兼容)
// Deprecated: Use JWTService.GenerateRefreshToken instead
func GenerateRefreshToken(userID int, username, role, tenantID string, tokenVersion int) (string, string, error) {
	if globalJWTService == nil {
		return "", "", errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
}

// GenerateTokenPair 全局函数 (向后兼容)
// Deprecated: Use JWTService.GenerateTokenPair instead
func GenerateTokenPair(userID int, username, role, tenantID string, tokenVersion int) (*TokenPair, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.GenerateTokenPair(userID, username, role, tenantID, tokenVersion)
}

// ParseToken 全局函数 (向后兼容)
// Deprecated: Use JWTService.ParseToken instead
func ParseToken(tokenString string) (*Claims, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.ParseToken(tokenString)
}

// RefreshAccessToken 全局函数 (向后兼容)
// Deprecated: Use JWTService.RefreshAccessToken instead
func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	if globalJWTService == nil {
		return nil, errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.RefreshAccessToken(refreshTokenString)
}

// RevokeToken 全局函数 (向后兼容)
// Deprecated: Use JWTService.RevokeToken instead
func RevokeToken(tokenString string) error {
	if globalJWTService == nil {
		return errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.RevokeToken(tokenString)
}

// RevokeAllUserTokens 全局函数 (向后兼容)
// Deprecated: Use JWTService.RevokeAllUserTokens instead
func RevokeAllUserTokens(userID int) error {
	if globalJWTService == nil {
		return errors.New("JWT not initialized - call InitJWT first")
	}
	return globalJWTService.RevokeAllUserTokens(userID)
}

// GenerateToken 旧 API (向后兼容)
// Deprecated: Use JWTService.GenerateAccessToken instead
func GenerateToken(userID int, username, role string) (string, error) {
	if globalJWTService == nil {
		return "", errors.New("JWT not initialized - call InitJWT first")
	}
	token, _, err := globalJWTService.GenerateAccessToken(userID, username, role, "", 0)
	return token, err
}
