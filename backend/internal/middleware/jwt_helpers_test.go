package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Initialize JWT secret for tests
	SetJWTSecret("test-secret-key-for-jwt-testing-123456")
}

// ============================================
// NewJWTConfig Tests
// ============================================

func TestNewJWTConfig_Success(t *testing.T) {
	config, err := NewJWTConfig("my-secret-key-very-long-1234567890")
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.NotEmpty(t, config.GetSecret())
}

func TestNewJWTConfig_EmptySecret(t *testing.T) {
	config, err := NewJWTConfig("")
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Equal(t, "JWT secret is required", err.Error())
}

func TestNewJWTConfig_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid secret",
			secret:  "valid-secret-key-123456789012345",
			wantErr: false,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: true,
			errMsg:  "JWT secret is required",
		},
		{
			name:    "short secret",
			secret:  "short",
			wantErr: false, // Short secrets are allowed, just not empty
		},
		{
			name:    "long secret",
			secret:  "this-is-a-very-long-secret-key-for-testing-purposes-1234567890",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewJWTConfig(tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

// ============================================
// JWTConfig.GetSecret Tests
// ============================================

func TestJWTConfig_GetSecret(t *testing.T) {
	secret := "test-secret-for-get-secret-12345678"
	config, err := NewJWTConfig(secret)
	require.NoError(t, err)

	retrievedSecret := config.GetSecret()
	assert.Equal(t, []byte(secret), retrievedSecret)
}

func TestJWTConfig_GetSecret_ConcurrentAccess(t *testing.T) {
	config, err := NewJWTConfig("initial-secret-key-12345678901")
	require.NoError(t, err)

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			secret := config.GetSecret()
			assert.NotEmpty(t, secret)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================
// JWTConfig.SetSecret Tests
// ============================================

func TestJWTConfig_SetSecret(t *testing.T) {
	config, err := NewJWTConfig("original-secret-key-12345678901")
	require.NoError(t, err)

	// Set new secret
	newSecret := "new-secret-key-rotated-123456789012"
	config.SetSecret(newSecret)

	// Verify new secret
	assert.Equal(t, []byte(newSecret), config.GetSecret())
}

func TestJWTConfig_SetSecret_EmptyString(t *testing.T) {
	originalSecret := "original-secret-key-1234567890123"
	config, err := NewJWTConfig(originalSecret)
	require.NoError(t, err)

	// Try to set empty secret - should not change
	config.SetSecret("")

	// Original secret should remain unchanged
	assert.Equal(t, []byte(originalSecret), config.GetSecret())
}

func TestJWTConfig_SetSecret_ConcurrentAccess(t *testing.T) {
	config, err := NewJWTConfig("initial-secret-key-123456789012")
	require.NoError(t, err)

	done := make(chan bool)

	// Concurrent writes and reads
	for i := 0; i < 5; i++ {
		go func(i int) {
			config.SetSecret("concurrent-secret-" + string(rune('a'+i)))
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			_ = config.GetSecret()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestJWTConfig_SetSecret_MultipleRotations(t *testing.T) {
	config, err := NewJWTConfig("secret-v1-12345678901234567890")
	require.NoError(t, err)

	secrets := []string{
		"secret-v2-12345678901234567890",
		"secret-v3-12345678901234567890",
		"secret-v4-12345678901234567890",
	}

	for _, newSecret := range secrets {
		config.SetSecret(newSecret)
		assert.Equal(t, []byte(newSecret), config.GetSecret())
	}
}

// ============================================
// GenerateJWTWithConfig Tests
// ============================================

func TestGenerateJWTWithConfig_Success(t *testing.T) {
	config, err := NewJWTConfig("test-secret-for-config-12345678901")
	require.NoError(t, err)

	token, err := GenerateJWTWithConfig(1, "testuser", "admin", "tenant-1", config)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateJWTWithConfig_NilConfig(t *testing.T) {
	token, err := GenerateJWTWithConfig(1, "testuser", "admin", "tenant-1", nil)
	require.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "JWT config is nil", err.Error())
}

func TestGenerateJWTWithConfig_TableDriven(t *testing.T) {
	config, err := NewJWTConfig("table-driven-secret-123456789012")
	require.NoError(t, err)

	tests := []struct {
		name     string
		userID   int
		username string
		role     string
		tenantID string
		config   *JWTConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid user",
			userID:   1,
			username: "admin",
			role:     "admin",
			tenantID: "tenant-1",
			config:   config,
			wantErr:  false,
		},
		{
			name:     "nil config",
			userID:   1,
			username: "test",
			role:     "user",
			tenantID: "tenant-1",
			config:   nil,
			wantErr:  true,
			errMsg:   "JWT config is nil",
		},
		{
			name:     "empty username allowed",
			userID:   2,
			username: "",
			role:     "viewer",
			tenantID: "tenant-2",
			config:   config,
			wantErr:  false,
		},
		{
			name:     "empty tenant allowed",
			userID:   3,
			username: "user",
			role:     "user",
			tenantID: "",
			config:   config,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWTWithConfig(tt.userID, tt.username, tt.role, tt.tenantID, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestGenerateJWTWithConfig_TokenValidity(t *testing.T) {
	config, err := NewJWTConfig("validity-test-secret-1234567890123")
	require.NoError(t, err)

	token, err := GenerateJWTWithConfig(42, "validuser", "operator", "tenant-42", config)
	require.NoError(t, err)

	// Parse token and verify claims
	claims, err := ParseJWTWithConfig(token, config)
	require.NoError(t, err)

	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "validuser", claims.Username)
	assert.Equal(t, "operator", claims.Role)
	assert.Equal(t, "tenant-42", claims.TenantID)
	assert.Equal(t, "industrial-ai-platform", claims.Issuer)
}

func TestGenerateJWTWithConfig_ExpiryTime(t *testing.T) {
	config, err := NewJWTConfig("expiry-test-secret-12345678901234")
	require.NoError(t, err)

	beforeGen := time.Now()
	token, err := GenerateJWTWithConfig(1, "user", "user", "t1", config)
	require.NoError(t, err)

	claims, err := ParseJWTWithConfig(token, config)
	require.NoError(t, err)

	// Check expiry is approximately AccessTokenDuration from now
	expectedExpiry := beforeGen.Add(AccessTokenDuration)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

// ============================================
// ParseJWTWithConfig Tests
// ============================================

func TestParseJWTWithConfig_Success(t *testing.T) {
	config, err := NewJWTConfig("parse-test-secret-123456789012345")
	require.NoError(t, err)

	token, err := GenerateJWTWithConfig(1, "testuser", "admin", "tenant-1", config)
	require.NoError(t, err)

	claims, err := ParseJWTWithConfig(token, config)
	require.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "tenant-1", claims.TenantID)
}

func TestParseJWTWithConfig_NilConfig(t *testing.T) {
	claims, err := ParseJWTWithConfig("some-token", nil)
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, "JWT config is nil", err.Error())
}

func TestParseJWTWithConfig_InvalidToken(t *testing.T) {
	config, err := NewJWTConfig("invalid-token-secret-123456789012")
	require.NoError(t, err)

	claims, err := ParseJWTWithConfig("invalid-token-string", config)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTWithConfig_WrongSecret(t *testing.T) {
	config1, err := NewJWTConfig("secret-one-123456789012345678901")
	require.NoError(t, err)

	config2, err := NewJWTConfig("secret-two-different-1234567890")
	require.NoError(t, err)

	// Generate token with config1
	token, err := GenerateJWTWithConfig(1, "user", "admin", "t1", config1)
	require.NoError(t, err)

	// Try to parse with config2 - should fail
	claims, err := ParseJWTWithConfig(token, config2)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTWithConfig_EmptyToken(t *testing.T) {
	config, err := NewJWTConfig("empty-token-secret-12345678901234")
	require.NoError(t, err)

	claims, err := ParseJWTWithConfig("", config)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWTWithConfig_TableDriven(t *testing.T) {
	validConfig, err := NewJWTConfig("valid-config-secret-12345678901234")
	require.NoError(t, err)

	validToken, err := GenerateJWTWithConfig(1, "validuser", "admin", "tenant-1", validConfig)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		config  *JWTConfig
		wantErr bool
	}{
		{
			name:    "valid token and config",
			token:   validToken,
			config:  validConfig,
			wantErr: false,
		},
		{
			name:    "nil config",
			token:   validToken,
			config:  nil,
			wantErr: true,
		},
		{
			name:    "invalid token",
			token:   "not-a-valid-jwt-token",
			config:  validConfig,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			config:  validConfig,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseJWTWithConfig(tt.token, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

// ============================================
// InitJWTConfig Tests
// ============================================

func TestInitJWTConfig_Success(t *testing.T) {
	err := InitJWTConfig("test-init-secret-12345678901234567")
	require.NoError(t, err)

	// Verify global config is set
	globalConfig := GetGlobalJWTConfig()
	require.NotNil(t, globalConfig)
	assert.NotEmpty(t, globalConfig.GetSecret())
}

func TestInitJWTConfig_EmptySecret(t *testing.T) {
	// Save current global config
	originalConfig := GetGlobalJWTConfig()

	err := InitJWTConfig("")
	require.Error(t, err)
	assert.Equal(t, "JWT secret is required", err.Error())

	// Global config should not have changed
	assert.Equal(t, originalConfig, GetGlobalJWTConfig())
}

func TestInitJWTConfig_MultipleInitializations(t *testing.T) {
	// First initialization
	err := InitJWTConfig("first-secret-123456789012345678")
	require.NoError(t, err)

	config1 := GetGlobalJWTConfig()
	require.NotNil(t, config1)

	// Second initialization - should overwrite
	err = InitJWTConfig("second-secret-1234567890123456")
	require.NoError(t, err)

	config2 := GetGlobalJWTConfig()
	require.NotNil(t, config2)

	// The secret should have changed
	assert.NotEqual(t, config1.GetSecret(), config2.GetSecret())
	assert.Equal(t, []byte("second-secret-1234567890123456"), config2.GetSecret())
}

// ============================================
// GetGlobalJWTConfig Tests
// ============================================

func TestGetGlobalJWTConfig_AfterInit(t *testing.T) {
	err := InitJWTConfig("get-global-config-secret-12345678")
	require.NoError(t, err)

	config := GetGlobalJWTConfig()
	require.NotNil(t, config)
	assert.Equal(t, []byte("get-global-config-secret-12345678"), config.GetSecret())
}

func TestGetGlobalJWTConfig_ReturnsSameInstance(t *testing.T) {
	err := InitJWTConfig("same-instance-secret-12345678901")
	require.NoError(t, err)

	config1 := GetGlobalJWTConfig()
	config2 := GetGlobalJWTConfig()

	// Should return the same instance
	assert.Equal(t, config1, config2)
}

// ============================================
// GenerateRefreshToken Tests
// ============================================

func TestGenerateRefreshToken_Success(t *testing.T) {
	token, err := GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	// Should be 64 hex characters (32 bytes)
	assert.Len(t, token, 64)
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate multiple tokens and verify uniqueness
	for i := 0; i < 100; i++ {
		token, err := GenerateRefreshToken()
		require.NoError(t, err)

		// Each token should be unique
		assert.False(t, tokens[token], "Generated duplicate token")
		tokens[token] = true
	}

	// All 100 tokens should be unique
	assert.Len(t, tokens, 100)
}

func TestGenerateRefreshToken_Format(t *testing.T) {
	token, err := GenerateRefreshToken()
	require.NoError(t, err)

	// Should only contain hex characters
	for _, c := range token {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"Token should only contain hex characters")
	}
}

// ============================================
// GetJWTSecret Tests
// ============================================

func TestGetJWTSecret_AfterInit(t *testing.T) {
	testSecret := "get-jwt-secret-test-12345678901234"
	err := InitJWTConfig(testSecret)
	require.NoError(t, err)

	secret := GetJWTSecret()
	assert.Equal(t, []byte(testSecret), secret)
}

func TestGetJWTSecret_NilGlobalConfig(t *testing.T) {
	// Save original global config
	originalConfig := globalJWTConfig

	// Set to nil
	globalJWTConfig = nil

	secret := GetJWTSecret()
	assert.Nil(t, secret)

	// Restore original config
	globalJWTConfig = originalConfig
}

// ============================================
// Integration Tests for JWTConfig
// ============================================

func TestJWTConfig_FullWorkflow(t *testing.T) {
	// Create config
	config, err := NewJWTConfig("workflow-test-secret-1234567890")
	require.NoError(t, err)

	// Generate token
	token, err := GenerateJWTWithConfig(1, "workflowuser", "admin", "tenant-workflow", config)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse token
	claims, err := ParseJWTWithConfig(token, config)
	require.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "workflowuser", claims.Username)

	// Rotate secret
	newSecret := "rotated-workflow-secret-123456789"
	config.SetSecret(newSecret)
	assert.Equal(t, []byte(newSecret), config.GetSecret())

	// Old token should still be parseable with new secret? No - different secret
	// Actually, SetSecret changes the secret, so old token won't validate
	_, err = ParseJWTWithConfig(token, config)
	assert.Error(t, err, "Old token should not validate with new secret")

	// Generate new token with rotated secret
	newToken, err := GenerateJWTWithConfig(2, "newuser", "user", "tenant-new", config)
	require.NoError(t, err)

	// New token should validate
	newClaims, err := ParseJWTWithConfig(newToken, config)
	require.NoError(t, err)
	assert.Equal(t, 2, newClaims.UserID)
}

func TestJWTConfig_ConcurrentTokenOperations(t *testing.T) {
	config, err := NewJWTConfig("concurrent-test-secret-12345678901")
	require.NoError(t, err)

	done := make(chan bool)

	// Concurrent token generation
	for i := 0; i < 10; i++ {
		go func(id int) {
			token, err := GenerateJWTWithConfig(id, "user", "admin", "tenant", config)
			require.NoError(t, err)

			_, err = ParseJWTWithConfig(token, config)
			require.NoError(t, err)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================
// GenerateJWT Tests
// ============================================

func TestGenerateJWT_Success(t *testing.T) {
	token, err := GenerateJWT(1, "testuser", "admin", "tenant-1", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateJWT_WithCustomSecret(t *testing.T) {
	customSecret := []byte("custom-secret-key-very-long-12345678")
	token, err := GenerateJWT(2, "user2", "user", "tenant-2", customSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateJWT_EmptySecret(t *testing.T) {
	// When secret is nil and not initialized, should return error
	// But we initialized in init(), so this should work
	token, err := GenerateJWT(1, "test", "user", "t1", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateJWT_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		role     string
		tenantID string
		secret   []byte
		wantErr  bool
	}{
		{
			name:     "valid admin",
			userID:   1,
			username: "admin",
			role:     "admin",
			tenantID: "t1",
			secret:   nil,
			wantErr:  false,
		},
		{
			name:     "valid user",
			userID:   2,
			username: "user",
			role:     "user",
			tenantID: "t2",
			secret:   nil,
			wantErr:  false,
		},
		{
			name:     "with custom secret",
			userID:   3,
			username: "custom",
			role:     "viewer",
			tenantID: "",
			secret:   []byte("custom-secret-12345678901234567890"),
			wantErr:  false,
		},
		{
			name:     "empty username",
			userID:   4,
			username: "",
			role:     "user",
			tenantID: "t4",
			secret:   nil,
			wantErr:  false, // Empty username is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.username, tt.role, tt.tenantID, tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

// ============================================
// ParseJWT Tests
// ============================================

func TestParseJWT_Success(t *testing.T) {
	// First generate a token
	token, err := GenerateJWT(1, "testuser", "admin", "tenant-1", nil)
	require.NoError(t, err)

	// Then parse it
	claims, err := ParseJWT(token, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "tenant-1", claims.TenantID)
}

func TestParseJWT_WithCustomSecret(t *testing.T) {
	customSecret := []byte("custom-secret-key-very-long-12345678")
	token, err := GenerateJWT(2, "user2", "user", "tenant-2", customSecret)
	require.NoError(t, err)

	// Parse with same secret
	claims, err := ParseJWT(token, customSecret)
	require.NoError(t, err)
	assert.Equal(t, 2, claims.UserID)
	assert.Equal(t, "user2", claims.Username)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	claims, err := ParseJWT("invalid-token-string", nil)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_WrongSecret(t *testing.T) {
	token, err := GenerateJWT(1, "test", "user", "t1", nil)
	require.NoError(t, err)

	// Try to parse with different secret
	wrongSecret := []byte("wrong-secret-key-123456789012345")
	claims, err := ParseJWT(token, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_TableDriven(t *testing.T) {
	validToken, _ := GenerateJWT(1, "test", "admin", "t1", nil)
	customSecret := []byte("custom-secret-12345678901234567890")
	customToken, _ := GenerateJWT(2, "custom", "user", "t2", customSecret)

	tests := []struct {
		name     string
		token    string
		secret   []byte
		wantErr  bool
		wantUser string
	}{
		{
			name:     "valid token default secret",
			token:    validToken,
			secret:   nil,
			wantErr:  false,
			wantUser: "test",
		},
		{
			name:     "valid token custom secret",
			token:    customToken,
			secret:   customSecret,
			wantErr:  false,
			wantUser: "custom",
		},
		{
			name:    "invalid token format",
			token:   "not-a-token",
			secret:  nil,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  nil,
			wantErr: true,
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  []byte("wrong-secret-12345678901234567"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseJWT(tt.token, tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				if tt.wantUser != "" {
					assert.Equal(t, tt.wantUser, claims.Username)
				}
			}
		})
	}
}

// ============================================
// SetJWTSecret Tests
// ============================================

func TestSetJWTSecret(t *testing.T) {
	newSecret := "new-test-secret-12345678901234567"
	SetJWTSecret(newSecret)

	// Generate and parse with new secret should work
	token, err := GenerateJWT(100, "newuser", "admin", "t100", nil)
	require.NoError(t, err)

	claims, err := ParseJWT(token, nil)
	require.NoError(t, err)
	assert.Equal(t, 100, claims.UserID)

	// Reset to original for other tests
	SetJWTSecret("test-secret-key-for-jwt-testing-123456")
}

// ============================================
// Token Expiry Tests
// ============================================

func TestGenerateJWT_TokenExpiry(t *testing.T) {
	token, err := GenerateJWT(1, "test", "admin", "t1", nil)
	require.NoError(t, err)

	claims, err := ParseJWT(token, nil)
	require.NoError(t, err)

	// Check expiry is set correctly
	assert.NotNil(t, claims.ExpiresAt)
	expectedExpiry := time.Now().Add(AccessTokenDuration)
	// Allow 5 second tolerance
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestGenerateJWT_Issuer(t *testing.T) {
	token, err := GenerateJWT(1, "test", "admin", "t1", nil)
	require.NoError(t, err)

	claims, err := ParseJWT(token, nil)
	require.NoError(t, err)

	assert.Equal(t, "industrial-ai-platform", claims.Issuer)
}

// ============================================
// Round-trip Tests
// ============================================

func TestJWT_RoundTrip(t *testing.T) {
	tests := []struct {
		userID   int
		username string
		role     string
		tenantID string
	}{
		{1, "admin", "admin", "tenant-1"},
		{2, "user", "user", "tenant-2"},
		{3, "viewer", "viewer", ""},
		{999, "testuser", "tenant_admin", "tenant-999"},
	}

	for _, tt := range tests {
		token, err := GenerateJWT(tt.userID, tt.username, tt.role, tt.tenantID, nil)
		require.NoError(t, err)

		claims, err := ParseJWT(token, nil)
		require.NoError(t, err)

		assert.Equal(t, tt.userID, claims.UserID)
		assert.Equal(t, tt.username, claims.Username)
		assert.Equal(t, tt.role, claims.Role)
		assert.Equal(t, tt.tenantID, claims.TenantID)
	}
}

// ============================================
// parseJWTInternal Edge Case Tests
// ============================================

func TestParseJWT_WrongSigningMethod(t *testing.T) {
	// Create a token with a different signing method (not HMAC)
	// This should trigger the signature method validation error branch
	claims := jwt.MapClaims{
		"user_id":   1,
		"username":  "testuser",
		"role":      "admin",
		"tenant_id": "tenant-1",
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
		"iss":       "industrial-ai-platform",
	}

	// Use a non-HMAC signing method (this will fail parsing)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// This should fail because it's not HMAC
	result, err := ParseJWT(tokenString, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseJWT_MalformedToken(t *testing.T) {
	// Test various malformed token formats
	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"no dots", "notokennodelimiters"},
		{"one dot", "one.dotonly"},
		{"too many dots", "too.many.dots.in.token"},
		{"invalid base64", "invalid!base64.invalid!header.invalid!payload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseJWT(tt.token, nil)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestGenerateJWT_NotInitialized(t *testing.T) {
	// Save original global config
	originalConfig := globalJWTConfig
	globalJWTConfig = nil

	// Generate with nil secret and no global config should fail
	token, err := GenerateJWT(1, "test", "user", "t1", nil)
	require.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "JWT not initialized")

	// Restore original config
	globalJWTConfig = originalConfig
}

func TestParseJWT_NotInitialized(t *testing.T) {
	// Save original global config
	originalConfig := globalJWTConfig
	globalJWTConfig = nil

	// Parse with nil secret and no global config should fail
	claims, err := ParseJWT("some-valid-looking-token", nil)
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "JWT not initialized")

	// Restore original config
	globalJWTConfig = originalConfig
}

// ============================================
// Claims Validation Edge Cases
// ============================================

func TestJWTClaims_EdgeValues(t *testing.T) {
	config, err := NewJWTConfig("edge-values-secret-12345678901234")
	require.NoError(t, err)

	tests := []struct {
		name     string
		userID   int
		username string
		role     string
		tenantID string
	}{
		{"zero user ID", 0, "zero", "viewer", "t0"},
		{"negative user ID", -1, "negative", "user", "t-negative"},
		{"large user ID", 999999999, "largeid", "admin", "t-large"},
		{"special chars in username", 1, "user@domain.com", "user", "t1"},
		{"unicode username", 2, "用户名", "admin", "t-unicode"},
		{"long username", 3, "verylongusernamewithmanycharacters", "viewer", "t-long"},
		{"space in username", 4, "user name", "user", "t-space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWTWithConfig(tt.userID, tt.username, tt.role, tt.tenantID, config)
			require.NoError(t, err)

			claims, err := ParseJWTWithConfig(token, config)
			require.NoError(t, err)

			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.role, claims.Role)
			assert.Equal(t, tt.tenantID, claims.TenantID)
		})
	}
}
