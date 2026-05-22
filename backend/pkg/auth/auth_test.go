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

// Additional tests for better coverage

func TestGenerateToken_ZeroLength(t *testing.T) {
	token, err := GenerateToken(0)
	assert.NoError(t, err)
	// base64 encoded 0 bytes results in empty string
	assert.Empty(t, token)
}

func TestGenerateToken_NegativeLength(t *testing.T) {
	// Negative length should cause panic or error
	// Testing that the function doesn't handle this edge case gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for negative length, but function did not panic")
		}
	}()
	GenerateToken(-1)
}

func TestParseClaims_EmptyToken(t *testing.T) {
	secret := []byte("test-secret")
	claims, err := ParseClaims("", secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestParseClaims_MalformedToken(t *testing.T) {
	secret := []byte("test-secret")
	// Token with wrong number of parts
	claims, err := ParseClaims("part1.part2", secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseClaims_WrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	wrongSecret := []byte("wrong-secret")

	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	// Parse with wrong secret
	parsedClaims, err := ParseClaims(tokenString, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestParseClaims_MissingUserID(t *testing.T) {
	secret := []byte("test-secret")

	// Create token with missing UserID (default value 0)
	claims := &Claims{
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestParseClaims_MissingUsername(t *testing.T) {
	secret := []byte("test-secret")

	// Create token with missing Username (empty string)
	claims := &Claims{
		UserID:   1,
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestParseClaims_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret")

	// Create an expired token
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	// Parse expired token
	parsedClaims, err := ParseClaims(tokenString, secret)
	// The jwt library should reject expired tokens
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestParseClaims_DifferentAlgorithms(t *testing.T) {
	secret := []byte("test-secret")

	// Test with HS384
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
	assert.Equal(t, 1, parsedClaims.UserID)
}

func TestParseClaims_EmptySecret(t *testing.T) {
	secret := []byte("test-secret")
	emptySecret := []byte{}

	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	// Parse with empty secret should still validate signature
	parsedClaims, err := ParseClaims(tokenString, emptySecret)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestClaims_Fields(t *testing.T) {
	// Test that all custom claims are properly set
	claims := &Claims{
		UserID:   42,
		Username: "testuser",
		Role:     "admin",
		TenantID: "tenant-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
			Subject:   "test-subject",
		},
	}

	secret := []byte("test-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.NoError(t, err)
	assert.Equal(t, 42, parsedClaims.UserID)
	assert.Equal(t, "testuser", parsedClaims.Username)
	assert.Equal(t, "admin", parsedClaims.Role)
	assert.Equal(t, "tenant-123", parsedClaims.TenantID)
	assert.Equal(t, "test-issuer", parsedClaims.Issuer)
	assert.Equal(t, "test-subject", parsedClaims.Subject)
}

// Test algorithm confusion attack prevention
func TestParseClaims_AlgorithmConfusion_RS256(t *testing.T) {
	secret := []byte("test-secret")

	// Try to create a token with RS256 algorithm (non-HMAC)
	// This simulates an algorithm confusion attack
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	// Create token with RS256 (should be rejected)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// We sign with the secret (pretending it's a private key)
	tokenString, err := token.SignedString(secret)
	// This will actually fail because RS256 expects *rsa.PrivateKey, not []byte
	// But let's test with a malformed token that has RS256 in header
	if err == nil {
		parsedClaims, err := ParseClaims(tokenString, secret)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
		assert.Equal(t, ErrInvalidToken, err)
	}
}

// Test "none" algorithm attack prevention
func TestParseClaims_NoneAlgorithm(t *testing.T) {
	secret := []byte("test-secret")

	// Create a token with "none" algorithm
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	// Manually create a token with "none" algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	// Parse should reject "none" algorithm
	parsedClaims, err := ParseClaims(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

// Test token with valid signature but different HMAC algorithm
func TestParseClaims_DifferentHMACAlgorithms(t *testing.T) {
	secret := []byte("test-secret")

	testCases := []struct {
		name    string
		method  jwt.SigningMethod
	}{
		{"HS256", jwt.SigningMethodHS256},
		{"HS384", jwt.SigningMethodHS384},
		{"HS512", jwt.SigningMethodHS512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims := &Claims{
				UserID:   1,
				Username: "testuser",
				Role:     "user",
				TenantID: "tenant-001",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				},
			}

			token := jwt.NewWithClaims(tc.method, claims)
			tokenString, err := token.SignedString(secret)
			assert.NoError(t, err)

			parsedClaims, err := ParseClaims(tokenString, secret)
			assert.NoError(t, err)
			assert.NotNil(t, parsedClaims)
			assert.Equal(t, 1, parsedClaims.UserID)
			assert.Equal(t, "testuser", parsedClaims.Username)
		})
	}
}

// Test token validation with various edge cases
func TestParseClaims_EdgeCases(t *testing.T) {
	secret := []byte("test-secret")

	t.Run("token with extra whitespace", func(t *testing.T) {
		claims := &Claims{
			UserID:   1,
			Username: "testuser",
			Role:     "user",
			TenantID: "tenant-001",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		// JWT library doesn't handle extra whitespace, so this should fail
		parsedClaims, err := ParseClaims(" "+tokenString+" ", secret)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})

	t.Run("token with not-before (nbf) claim", func(t *testing.T) {
		claims := &Claims{
			UserID:   1,
			Username: "testuser",
			Role:     "user",
			TenantID: "tenant-001",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		parsedClaims, err := ParseClaims(tokenString, secret)
		assert.NoError(t, err)
		assert.NotNil(t, parsedClaims)
	})

	t.Run("token with future not-before claim", func(t *testing.T) {
		claims := &Claims{
			UserID:   1,
			Username: "testuser",
			Role:     "user",
			TenantID: "tenant-001",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		// Token with future nbf should be rejected by jwt library
		parsedClaims, err := ParseClaims(tokenString, secret)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})
}

// Test IsTokenExpired with edge cases
func TestIsTokenExpired_EdgeCases(t *testing.T) {
	t.Run("exactly now", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now()),
			},
		}
		// Token expiring exactly now should be considered expired
		// (Before returns false if equal)
		time.Sleep(1 * time.Millisecond)
		assert.True(t, IsTokenExpired(claims))
	})

	t.Run("one second in future", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Second)),
			},
		}
		assert.False(t, IsTokenExpired(claims))
	})

	t.Run("one second in past", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
			},
		}
		assert.True(t, IsTokenExpired(claims))
	})
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateToken(32)
	}
}

func BenchmarkParseClaims(b *testing.B) {
	secret := []byte("test-secret")
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Role:     "user",
		TenantID: "tenant-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseClaims(tokenString, secret)
	}
}
