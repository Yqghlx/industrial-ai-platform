package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Initialize JWT secret for tests
	SetJWTSecret("test-secret-key-for-jwt-testing-123456")
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
