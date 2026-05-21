package service

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenVersionValidation 测试 Token 版本验证功能
func TestTokenVersionValidation(t *testing.T) {
	// 初始化 JWT
	err := InitJWT("test-secret-key-at-least-32-characters-long")
	require.NoError(t, err)

	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)

	// 设置 UserTokenStore
	SetUserTokenStore(userRepo)

	// 测试数据
	userID := 1
	username := "testuser"
	role := "user"
	tenantID := "tenant1"
	initialTokenVersion := 1

	t.Run("GenerateAndParseTokenWithVersion", func(t *testing.T) {
		// 生成 Token Pair
		tokenPair, err := GenerateTokenPair(userID, username, role, tenantID, initialTokenVersion)
		require.NoError(t, err)
		require.NotEmpty(t, tokenPair.AccessToken)
		require.NotEmpty(t, tokenPair.RefreshToken)

		// Mock 数据库查询 - 返回匹配的 token version
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(initialTokenVersion))

		// 解析 Access Token - 应该成功
		claims, err := ParseToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, initialTokenVersion, claims.TokenVersion)

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TokenVersionMismatch", func(t *testing.T) {
		// 生成 Token (version = 1)
		tokenPair, err := GenerateTokenPair(userID, username, role, tenantID, initialTokenVersion)
		require.NoError(t, err)

		// Mock 数据库查询 - 返回不同的 token version (version = 2)
		// 这意味着 token 已被撤销
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(initialTokenVersion + 1))

		// 解析 Token - 应该失败，因为版本不匹配
		claims, err := ParseToken(tokenPair.AccessToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token has been revoked")

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RevokeAllUserTokens", func(t *testing.T) {
		// Mock 数据库更新 - 递增 token version
		mock.ExpectExec("UPDATE users SET token_version").
			WithArgs(sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 撤销用户所有 Token
		err := RevokeAllUserTokens(userID)
		assert.NoError(t, err)

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestChangePasswordRevokesTokens 测试修改密码后旧 Token 失效
func TestChangePasswordRevokesTokens(t *testing.T) {
	// 初始化 JWT
	err := InitJWT("test-secret-key-at-least-32-characters-long")
	require.NoError(t, err)

	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	SetUserTokenStore(userRepo)

	// 初始数据
	userID := 1
	username := "testuser"
	initialVersion := 1
	newVersion := 2

	t.Run("OldTokenInvalidAfterPasswordChange", func(t *testing.T) {
		// 1. 用户登录，获取 Token (version = 1)
		tokenPair, err := GenerateTokenPair(userID, username, "user", "tenant1", initialVersion)
		require.NoError(t, err)

		// 2. 验证 Token 有效
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(initialVersion))

		claims, err := ParseToken(tokenPair.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// 3. 用户修改密码 - 撤销所有 Token
		mock.ExpectExec("UPDATE users SET token_version").
			WithArgs(sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = RevokeAllUserTokens(userID)
		assert.NoError(t, err)

		// 4. 旧 Token 应该失效 (version mismatch)
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(newVersion))

		claims, err = ParseToken(tokenPair.AccessToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token has been revoked")

		// 5. 用户重新登录，获取新 Token (version = 2)
		newTokenPair, err := GenerateTokenPair(userID, username, "user", "tenant1", newVersion)
		require.NoError(t, err)

		// 6. 新 Token 应该有效
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(newVersion))

		claims, err = ParseToken(newTokenPair.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, newVersion, claims.TokenVersion)

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestRefreshTokenWithVersion 测试 Refresh Token 版本验证
func TestRefreshTokenWithVersion(t *testing.T) {
	// 初始化 JWT
	err := InitJWT("test-secret-key-at-least-32-characters-long")
	require.NoError(t, err)

	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	SetUserTokenStore(userRepo)

	userID := 1
	username := "testuser"
	role := "user"
	tenantID := "tenant1"
	tokenVersion := 1

	t.Run("RefreshTokenSuccess", func(t *testing.T) {
		// 生成 Refresh Token
		// FIX-007: GenerateRefreshToken 返回 (tokenString, tokenID, err)
		refreshToken, _, err := GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
		require.NoError(t, err)

		// Mock 数据库查询 - 返回当前 token version
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(tokenVersion))

		// 使用 Refresh Token 刷新
		newTokenPair, err := RefreshAccessToken(refreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, newTokenPair)
		assert.NotEmpty(t, newTokenPair.AccessToken)
		assert.NotEmpty(t, newTokenPair.RefreshToken)

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RefreshTokenAfterRevoke", func(t *testing.T) {
		// 生成 Refresh Token (version = 1)
		// FIX-007: GenerateRefreshToken 返回 (tokenString, tokenID, err)
		refreshToken, _, err := GenerateRefreshToken(userID, username, role, tenantID, tokenVersion)
		require.NoError(t, err)

		// Mock 数据库查询 - 返回新的 token version (version = 2)
		// 意味着 Token 已被撤销
		mock.ExpectQuery("SELECT token_version FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"token_version"}).
				AddRow(tokenVersion + 1))

		// 使用 Refresh Token 刷新 - 应该失败
		newTokenPair, err := RefreshAccessToken(refreshToken)
		assert.Error(t, err)
		assert.Nil(t, newTokenPair)
		assert.Contains(t, err.Error(), "token has been revoked")

		// 验证所有期望都被满足
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestTokenExpiration 测试 Token 过期时间
func TestTokenExpiration(t *testing.T) {
	// 初始化 JWT
	err := InitJWT("test-secret-key-at-least-32-characters-long")
	require.NoError(t, err)

	userID := 1
	username := "testuser"
	tokenVersion := 1

	t.Run("AccessTokenExpiration", func(t *testing.T) {
		// 生成 Access Token
		accessToken, _, err := GenerateAccessToken(userID, username, "user", "tenant1", tokenVersion)
		require.NoError(t, err)

		// 解析 Token
		claims, err := ParseToken(accessToken)
		require.NoError(t, err)

		// 验证过期时间 (15 分钟)
		expiresAt := claims.ExpiresAt.Time
		expectedExpiry := time.Now().Add(AccessTokenDuration)

		// 允许 1 秒误差
		diff := expiresAt.Sub(expectedExpiry)
		assert.True(t, diff < time.Second && diff > -time.Second,
			"Access token should expire in approximately 15 minutes")
	})

	t.Run("RefreshTokenExpiration", func(t *testing.T) {
		// 生成 Refresh Token
		refreshToken, _, err := GenerateRefreshToken(userID, username, "user", "tenant1", tokenVersion)
		require.NoError(t, err)

		// 解析 Token
		claims, err := ParseToken(refreshToken)
		require.NoError(t, err)

		// 验证过期时间 (7 天)
		expiresAt := claims.ExpiresAt.Time
		expectedExpiry := time.Now().Add(RefreshTokenDuration)

		// 允许 1 秒误差
		diff := expiresAt.Sub(expectedExpiry)
		assert.True(t, diff < time.Second && diff > -time.Second,
			"Refresh token should expire in approximately 7 days")
	})
}
