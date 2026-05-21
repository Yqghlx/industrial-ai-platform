package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(32)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 44) // base64 encoded 32 bytes = 44 chars
}

func TestGenerateToken_DifferentEachTime(t *testing.T) {
	token1, err := GenerateToken(16)
	assert.NoError(t, err)
	token2, err := GenerateToken(16)
	assert.NoError(t, err)
	assert.NotEqual(t, token1, token2)
}

func TestParseClaims_InvalidToken(t *testing.T) {
	secret := []byte("test-secret")
	claims, err := ParseClaims("invalid-token", secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestParseClaims_ValidToken(t *testing.T) {
	secret := []byte("test-secret")

	// Create a valid token
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "admin",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	// Parse the token
	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
	assert.Equal(t, 1, parsedClaims.UserID)
	assert.Equal(t, "testuser", parsedClaims.Username)
	assert.Equal(t, "admin", parsedClaims.Role)
	assert.Equal(t, "tenant-001", parsedClaims.TenantID)
}

func TestIsTokenExpired_NotExpired(t *testing.T) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	assert.False(t, IsTokenExpired(claims))
}

func TestIsTokenExpired_Expired(t *testing.T) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
		},
	}
	assert.True(t, IsTokenExpired(claims))
}

func TestIsTokenExpired_NoExpiry(t *testing.T) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	assert.False(t, IsTokenExpired(claims))
}
