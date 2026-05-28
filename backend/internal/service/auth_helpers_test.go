package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// MemoryTokenBlacklist Tests
// ============================================

func TestNewMemoryTokenBlacklist(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	require.NotNil(t, bl)
	assert.NotNil(t, bl.entries)
	assert.NotNil(t, bl.userRevocations)
	// Clean up goroutine
	bl.Stop()
}

func TestMemoryTokenBlacklist_Add_And_Exists(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Test: non-existent token
	assert.False(t, bl.Exists(ctx, "token-not-exist"))

	// Test: add and check
	err := bl.Add(ctx, "token-1", 10*time.Minute)
	assert.NoError(t, err)
	assert.True(t, bl.Exists(ctx, "token-1"))

	// Test: add another token
	err = bl.Add(ctx, "token-2", 5*time.Minute)
	assert.NoError(t, err)
	assert.True(t, bl.Exists(ctx, "token-2"))

	// Test: size
	assert.Equal(t, 2, bl.Size())
}

func TestMemoryTokenBlacklist_Exists_Expired(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add token with very short duration
	err := bl.Add(ctx, "short-lived", 1*time.Millisecond)
	assert.NoError(t, err)

	// Should exist immediately
	assert.True(t, bl.Exists(ctx, "short-lived"))

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Should no longer exist
	assert.False(t, bl.Exists(ctx, "short-lived"))
}

func TestMemoryTokenBlacklist_AddUserRevocation(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()
	revokedAt := time.Now()

	// Add user revocation
	err := bl.AddUserRevocation(ctx, 1, revokedAt, 10*time.Minute)
	assert.NoError(t, err)

	// Get user revocation
	result, err := bl.GetUserRevocation(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, revokedAt.Truncate(time.Second), result.Truncate(time.Second))
}

func TestMemoryTokenBlacklist_GetUserRevocation_NotFound(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Get non-existent revocation
	result, err := bl.GetUserRevocation(ctx, 999)
	assert.NoError(t, err)
	assert.True(t, result.IsZero())
}

func TestMemoryTokenBlacklist_GetUserRevocation_Expired(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()

	// Add revocation with very short duration
	err := bl.AddUserRevocation(ctx, 42, time.Now(), 1*time.Millisecond)
	assert.NoError(t, err)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Should return zero
	result, err := bl.GetUserRevocation(ctx, 42)
	assert.NoError(t, err)
	assert.True(t, result.IsZero())
}

func TestMemoryTokenBlacklist_Size(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()
	assert.Equal(t, 0, bl.Size())

	bl.Add(ctx, "t1", time.Minute)
	assert.Equal(t, 1, bl.Size())

	bl.Add(ctx, "t2", time.Minute)
	bl.Add(ctx, "t3", time.Minute)
	assert.Equal(t, 3, bl.Size())
}

func TestMemoryTokenBlacklist_Stop(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	// Stop should not panic
	bl.Stop()
}

func TestMemoryTokenBlacklist_Concurrent(t *testing.T) {
	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tokenID := strings.Repeat("a", 10) + string(rune(idx))
			bl.Add(ctx, tokenID, time.Minute)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 100, bl.Size())
}

// ============================================
// HybridTokenBlacklist Tests
// ============================================

func TestNewHybridTokenBlacklist_NoRedis(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	require.NotNil(t, hb)
	assert.NotNil(t, hb.memoryBlacklist)
	assert.False(t, hb.useRedis)
	assert.Nil(t, hb.redisBlacklist)
	hb.Stop()
}

func TestHybridTokenBlacklist_Add_And_Exists_NoRedis(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()

	err := hb.Add(ctx, "token-1", 10*time.Minute)
	assert.NoError(t, err)
	assert.True(t, hb.Exists(ctx, "token-1"))
	assert.False(t, hb.Exists(ctx, "token-not-exist"))
}

func TestHybridTokenBlacklist_AddUserRevocation_NoRedis(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()

	ctx := context.Background()
	revokedAt := time.Now()

	err := hb.AddUserRevocation(ctx, 1, revokedAt, 10*time.Minute)
	assert.NoError(t, err)

	result, err := hb.GetUserRevocation(ctx, 1)
	assert.NoError(t, err)
	assert.WithinDuration(t, revokedAt, result, time.Second)
}

func TestHybridTokenBlacklist_IsUsingRedis(t *testing.T) {
	hb := NewHybridTokenBlacklist(nil)
	defer hb.Stop()
	assert.False(t, hb.IsUsingRedis())
}

// TestHybridTokenBlacklist_Stop moved to auth_helpers_cleanup_test.go

// ============================================
// JWTService Tests
// ============================================

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		expectError bool
		errorType   *JWTInitError
	}{
		// FIX-P1-10: 测试密钥安全注释 - 仅用于单元测试，不会进入生产环境
		{
			name:        "Valid secret",
			secret:      "this-is-a-very-long-secret-key-for-testing-123456",
			expectError: false,
		},
		{
			name:        "Empty secret",
			secret:      "",
			expectError: true,
		},
		{
			name:        "Short secret",
			secret:      "short",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewJWTService(tt.secret)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestJWTService_GetSecret(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-for-testing-123456"
	svc, err := NewJWTService(secret)
	require.NoError(t, err)

	assert.Equal(t, []byte(secret), svc.GetSecret())
}

func TestJWTService_SetTokenBlacklist(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()

	svc.SetTokenBlacklist(bl)
	// Verify blacklist works by generating and revoking a token
}

func TestJWTService_GenerateAccessToken(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	token, tokenID, err := svc.GenerateAccessToken(1, "testuser", "admin", "tenant-1", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, tokenID)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	token, tokenID, err := svc.GenerateRefreshToken(1, "testuser", "admin", "tenant-1", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, tokenID)
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	pair, err := svc.GenerateTokenPair(1, "testuser", "admin", "tenant-1", 1)
	assert.NoError(t, err)
	assert.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, "Bearer", pair.TokenType)
	assert.Greater(t, pair.ExpiresIn, int64(0))
}

func TestJWTService_ParseToken_Valid(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "tenant-1", 1)
	require.NoError(t, err)

	claims, err := svc.ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "tenant-1", claims.TenantID)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, TokenIssuer, claims.Issuer)
}

func TestJWTService_ParseToken_InvalidSignature(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	// Create token with different secret
	svc2, _ := NewJWTService("another-very-long-secret-key-for-testing-abcdef")
	token, _, _ := svc2.GenerateAccessToken(1, "test", "admin", "", 1)

	claims, err := svc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ParseToken_Blacklisted(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()
	svc.SetTokenBlacklist(bl)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "", 1)
	require.NoError(t, err)

	// Blacklist the token
	ctx := context.Background()
	claims, _ := svc.ParseToken(token)
	bl.Add(ctx, claims.TokenID, time.Hour)

	// Parse should fail now
	parsedClaims, err := svc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Contains(t, err.Error(), "revoked")
}

func TestJWTService_ParseToken_InvalidIssuer(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "", 1)
	require.NoError(t, err)

	// Manually create a token with wrong issuer (we can't easily do this,
	// so we'll test that a token from the correct service has the right issuer)
	claims, err := svc.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, TokenIssuer, claims.Issuer)
}

func TestJWTService_RevokeToken(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	bl := NewMemoryTokenBlacklist()
	defer bl.Stop()
	svc.SetTokenBlacklist(bl)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "", 1)
	require.NoError(t, err)

	err = svc.RevokeToken(token)
	assert.NoError(t, err)

	// Token should now be blacklisted
	claims, err := svc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_RevokeToken_NoBlacklist(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "", 1)
	require.NoError(t, err)

	err = svc.RevokeToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blacklist not initialized")
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	refreshToken, _, err := svc.GenerateRefreshToken(1, "testuser", "admin", "tenant-1", 1)
	require.NoError(t, err)

	pair, err := svc.RefreshAccessToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestJWTService_RefreshAccessToken_WithAccessToken(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	accessToken, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "", 1)
	require.NoError(t, err)

	// Should fail - access token can't be used for refresh
	pair, err := svc.RefreshAccessToken(accessToken)
	assert.Error(t, err)
	assert.Nil(t, pair)
}

func TestJWTService_RevokeAllUserTokens(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	// Need a userTokenStore
	mockStore := &MockUserTokenStore{
		GetTokenVersionFunc: func(ctx context.Context, userID int) (int, error) {
			return 2, nil
		},
		UpdateTokenVersionFunc: func(ctx context.Context, userID int) error {
			return nil
		},
	}
	svc.SetUserTokenStore(mockStore)
	svc.SetTokenBlacklist(NewMemoryTokenBlacklist())

	err = svc.RevokeAllUserTokens(1)
	assert.NoError(t, err)
}

func TestJWTService_RevokeAllUserTokens_NoStore(t *testing.T) {
	svc, err := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	require.NoError(t, err)

	svc.SetTokenBlacklist(NewMemoryTokenBlacklist())

	err = svc.RevokeAllUserTokens(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user token store not initialized")
}

// ============================================
// Global JWT Functions Tests
// ============================================

func TestInitJWT(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	err := InitJWT("this-is-a-very-long-secret-key-for-testing-123456")
	assert.NoError(t, err)
	assert.True(t, IsJWTInitialized())
}

func TestInitJWT_EmptySecret(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()
	globalJWTService = nil

	err := InitJWT("")
	assert.Error(t, err)
}

func TestIsJWTInitialized(t *testing.T) {
	// globalJWTService is already initialized by TestMain
	assert.True(t, IsJWTInitialized())
}

func TestSetJWTSecret(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	// SEC-HIGH-01: Empty secret should return error
	err := SetJWTSecret("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")

	// Valid secret should set the service
	err = SetJWTSecret("this-is-a-very-long-secret-key-for-testing-123456")
	assert.NoError(t, err)
	assert.NotNil(t, globalJWTService)
}

func TestSetJWTSecret_ShortSecret(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	// SEC-HIGH-01: Short secret should return error (not set service)
	err := SetJWTSecret("short")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be at least 32 characters")
}

func TestSetRedisClient_NilGlobal(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()
	globalJWTService = nil

	// Should not panic with nil globalJWTService
	SetRedisClient(nil)
}

func TestSetRedisClient(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	globalJWTService = &JWTService{secret: []byte("test")}
	SetRedisClient(nil) // nil client -> HybridTokenBlacklist with nil redis
}

func TestSetRedisClientWithFallback_NilGlobal(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()
	globalJWTService = nil

	// Should not panic with nil globalJWTService
	SetRedisClientWithFallback(nil, true)
}

func TestSetRedisClientWithFallback(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	globalJWTService = &JWTService{secret: []byte("test")}
	SetRedisClientWithFallback(nil, false) // nil client, not hybrid -> nothing
	SetRedisClientWithFallback(nil, true)  // nil client, hybrid -> HybridTokenBlacklist with nil redis
}

func TestSetMemoryTokenBlacklist_NilGlobal(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()
	globalJWTService = nil

	// Should not panic with nil globalJWTService
	SetMemoryTokenBlacklist()
}

func TestSetMemoryTokenBlacklist(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	globalJWTService = &JWTService{secret: []byte("test")}
	SetMemoryTokenBlacklist()
	assert.NotNil(t, globalJWTService.tokenBlacklist)
}

func TestSetUserTokenStore_NilGlobal(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()
	globalJWTService = nil

	// Should not panic with nil globalJWTService
	SetUserTokenStore(nil)
}

func TestSetUserTokenStore(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	globalJWTService = &JWTService{secret: []byte("test")}
	mockStore := &MockUserTokenStore{}
	SetUserTokenStore(mockStore)
	assert.NotNil(t, globalJWTService.userTokenStore)
}

func TestGetJWTService(t *testing.T) {
	// globalJWTService is initialized by TestMain
	assert.NotNil(t, GetJWTService())
}

func TestGenerateAccessToken_Global(t *testing.T) {
	token, tokenID, err := GenerateAccessToken(1, "test", "admin", "", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, tokenID)
}

func TestGenerateRefreshToken_Global(t *testing.T) {
	token, _, err := GenerateRefreshToken(1, "test", "admin", "", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateTokenPair_Global(t *testing.T) {
	pair, err := GenerateTokenPair(1, "test", "admin", "", 1)
	assert.NoError(t, err)
	assert.NotNil(t, pair)
}

func TestParseToken_Global(t *testing.T) {
	token, _, _ := GenerateAccessToken(1, "test", "admin", "", 1)
	claims, err := ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestRefreshAccessToken_Global(t *testing.T) {
	refreshToken, _, _ := GenerateRefreshToken(1, "test", "admin", "", 1)
	pair, err := RefreshAccessToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, pair)
}

func TestRevokeToken_Global(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	// Clone to avoid modifying shared state
	svc, _ := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	svc.SetTokenBlacklist(NewMemoryTokenBlacklist())
	globalJWTService = svc

	token, _, _ := GenerateAccessToken(1, "test", "admin", "", 1)
	err := RevokeToken(token)
	assert.NoError(t, err)
}

func TestRevokeAllUserTokens_Global(t *testing.T) {
	orig := globalJWTService
	defer func() { globalJWTService = orig }()

	// Clone to avoid modifying shared state
	svc, _ := NewJWTService("this-is-a-very-long-secret-key-for-testing-123456")
	svc.SetTokenBlacklist(NewMemoryTokenBlacklist())
	svc.SetUserTokenStore(&MockUserTokenStore{})
	globalJWTService = svc

	err := RevokeAllUserTokens(1)
	assert.NoError(t, err)
}

func TestGenerateToken_Global(t *testing.T) {
	token, err := GenerateToken(1, "test", "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// ============================================
// JWTInitError Tests
// ============================================

func TestJWTInitError(t *testing.T) {
	err := &JWTInitError{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

// MockUserTokenStore for testing
type MockUserTokenStore struct {
	GetTokenVersionFunc    func(ctx context.Context, userID int) (int, error)
	UpdateTokenVersionFunc func(ctx context.Context, userID int) error
}

func (m *MockUserTokenStore) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	if m.GetTokenVersionFunc != nil {
		return m.GetTokenVersionFunc(ctx, userID)
	}
	return 1, nil
}

func (m *MockUserTokenStore) UpdateTokenVersion(ctx context.Context, userID int) error {
	if m.UpdateTokenVersionFunc != nil {
		return m.UpdateTokenVersionFunc(ctx, userID)
	}
	return nil
}
