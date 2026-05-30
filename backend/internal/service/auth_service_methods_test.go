package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ctxMatcher 返回一个匹配任意 context.Context 的 mock.Matcher
// 因为 ensureContextTimeout 会将 context.Background() 包装成 timerCtx，
// 所以不能直接用 ctx 对象做参数匹配
func ctxMatcher() any {
	return mock.Anything
}

// ============================================
// RefreshToken 方法测试
// ============================================

// TestRefreshToken_Success 测试使用有效的 refresh token 刷新令牌
func TestRefreshToken_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 使用全局 JWT 服务生成真实的 refresh token（由 TestMain 初始化）
	refreshToken, _, err := GenerateRefreshToken(1, "testuser", "admin", "", 0)
	require.NoError(t, err)

	tokenPair, err := authService.RefreshToken(ctx, refreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Greater(t, tokenPair.ExpiresIn, int64(0))
}

// TestRefreshToken_InvalidToken 测试使用无效的 token 刷新令牌
func TestRefreshToken_InvalidToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	tokenPair, err := authService.RefreshToken(ctx, "invalid-refresh-token")

	assert.Error(t, err)
	assert.Nil(t, tokenPair)
	assert.Contains(t, err.Error(), "Authentication failed")
}

// TestRefreshToken_WithAccessToken 测试使用 access token 代替 refresh token（应失败）
func TestRefreshToken_WithAccessToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 生成 access token 而非 refresh token
	accessToken, _, err := GenerateAccessToken(1, "testuser", "admin", "", 0)
	require.NoError(t, err)

	tokenPair, err := authService.RefreshToken(ctx, accessToken)

	assert.Error(t, err)
	assert.Nil(t, tokenPair)
}

// TestRefreshToken_ExpiredToken 测试使用过期的 refresh token
func TestRefreshToken_ExpiredToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 生成一个过期的 refresh token（通过直接构造 JWT）
	claims := Claims{
		UserID:    1,
		Username:  "testuser",
		Role:      "admin",
		TokenType: "refresh",
		TokenID:   "expired-refresh-id",
		RegisteredClaims: getExpiredRegisteredClaims(),
	}
	expiredToken := signTestToken(claims)

	tokenPair, err := authService.RefreshToken(ctx, expiredToken)

	assert.Error(t, err)
	assert.Nil(t, tokenPair)
}

// ============================================
// ChangePassword 方法测试
// ============================================

// TestChangePassword_Success 测试成功修改密码
func TestChangePassword_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 准备密码哈希
	oldPassword := "OldPass123!@#"
	hashedOldPassword, err := HashPassword(oldPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Password: hashedOldPassword,
		Role:     "user",
	}

	// Mock: GetByID 返回用户（ctxMatcher 因为 ensureContextTimeout 会改变 context 类型）
	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)
	// Mock: UpdatePassword 成功
	mockRepo.On("UpdatePassword", ctxMatcher(), 1, mock.AnythingOfType("string")).Return(nil)
	// Mock: UpdateTokenVersion 成功（密码修改后撤销旧 token）
	mockRepo.On("UpdateTokenVersion", ctxMatcher(), 1).Return(nil)

	newPassword := "NewPass456$%^"
	err = authService.ChangePassword(ctx, 1, oldPassword, newPassword)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_UserNotFound 测试修改密码时用户不存在
func TestChangePassword_UserNotFound(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// Mock: GetByID 返回错误
	mockRepo.On("GetByID", ctxMatcher(), 999).Return(nil, fmt.Errorf("user not found"))

	err := authService.ChangePassword(ctx, 999, "OldPass123!@#", "NewPass456$%^")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_WrongOldPassword 测试修改密码时旧密码错误
func TestChangePassword_WrongOldPassword(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	correctPassword := "CorrectPass123!@#"
	hashedPassword, err := HashPassword(correctPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	// Mock: GetByID 返回用户
	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)

	err = authService.ChangePassword(ctx, 1, "WrongPass456$%^", "NewPass789&*()")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication failed")
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_NewPasswordTooWeak 测试新密码不符合复杂度要求
func TestChangePassword_NewPasswordTooWeak(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	oldPassword := "OldPass123!@#"
	hashedPassword, err := HashPassword(oldPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)

	// 新密码太简单（缺少大写字母、特殊字符等）
	err = authService.ChangePassword(ctx, 1, oldPassword, "weak")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_UsesPasswordHashField 测试当 PasswordHash 字段有值时使用 PasswordHash
func TestChangePassword_UsesPasswordHashField(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	oldPassword := "OldPass123!@#"
	// 使用 PasswordHash 字段而非 Password 字段
	hashedPassword, err := HashPassword(oldPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Password:     "", // Password 为空，应使用 PasswordHash
		Role:         "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)
	mockRepo.On("UpdatePassword", ctxMatcher(), 1, mock.AnythingOfType("string")).Return(nil)
	mockRepo.On("UpdateTokenVersion", ctxMatcher(), 1).Return(nil)

	err = authService.ChangePassword(ctx, 1, oldPassword, "NewPass456$%^")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_UpdatePasswordFails 测试更新密码数据库操作失败
func TestChangePassword_UpdatePasswordFails(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	oldPassword := "OldPass123!@#"
	hashedPassword, err := HashPassword(oldPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)
	mockRepo.On("UpdatePassword", ctxMatcher(), 1, mock.AnythingOfType("string")).Return(fmt.Errorf("database error"))

	err = authService.ChangePassword(ctx, 1, oldPassword, "NewPass456$%^")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
	mockRepo.AssertExpectations(t)
}

// TestChangePassword_TokenVersionUpdateFails 测试修改密码后 token 版本更新失败（应仍然成功）
func TestChangePassword_TokenVersionUpdateFails(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	oldPassword := "OldPass123!@#"
	hashedPassword, err := HashPassword(oldPassword)
	require.NoError(t, err)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(user, nil)
	mockRepo.On("UpdatePassword", ctxMatcher(), 1, mock.AnythingOfType("string")).Return(nil)
	// Token 版本更新失败，但密码修改仍然应该成功
	mockRepo.On("UpdateTokenVersion", ctxMatcher(), 1).Return(fmt.Errorf("token version update failed"))

	err = authService.ChangePassword(ctx, 1, oldPassword, "NewPass456$%^")

	// 密码修改应该成功（token 版本失败只记录警告）
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// ============================================
// ValidateToken 方法测试
// ============================================

// TestValidateToken_Success 测试验证有效的 access token
func TestValidateToken_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 生成有效的 token
	token, err := GenerateToken(1, "testuser", "admin")
	require.NoError(t, err)

	claims, err := authService.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

// TestValidateToken_InvalidToken 测试验证无效的 token
func TestValidateToken_InvalidToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	claims, err := authService.ValidateToken(ctx, "invalid-token-string")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "Authentication failed")
}

// TestValidateToken_ExpiredToken 测试验证过期的 token
func TestValidateToken_ExpiredToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 构造过期的 token
	claims := Claims{
		UserID:    1,
		Username:  "testuser",
		Role:      "admin",
		TokenType: "access",
		TokenID:   "expired-token-id",
		RegisteredClaims: getExpiredRegisteredClaims(),
	}
	expiredToken := signTestToken(claims)

	result, err := authService.ValidateToken(ctx, expiredToken)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestValidateToken_TamperedToken 测试验证被篡改的 token
func TestValidateToken_TamperedToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	token, err := GenerateToken(1, "testuser", "admin")
	require.NoError(t, err)

	// 篡改 token 内容
	tamperedToken := token + "tampered"

	claims, err := authService.ValidateToken(ctx, tamperedToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateToken_EmptyToken 测试验证空 token
func TestValidateToken_EmptyToken(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	claims, err := authService.ValidateToken(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateToken_WithTokenVersion 测试验证带版本号的 token
func TestValidateToken_WithTokenVersion(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	// 使用带版本号的 token 生成
	token, err := GenerateTokenWithVersion(1, "testuser", "admin", "", 3)
	require.NoError(t, err)

	claims, err := authService.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

// ============================================
// ListUsers 方法测试
// ============================================

// TestListUsers_Success 测试成功列出用户
func TestListUsers_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	expectedUsers := []model.User{
		{ID: 1, Username: "admin", Email: "admin@test.com", Role: "admin"},
		{ID: 2, Username: "user1", Email: "user1@test.com", Role: "user"},
	}

	mockRepo.On("List", ctxMatcher(), 1, 10).Return(expectedUsers, 2, nil)

	users, total, err := authService.ListUsers(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, users, 2)
	assert.Equal(t, "admin", users[0].Username)
	assert.Equal(t, "user1", users[1].Username)
	mockRepo.AssertExpectations(t)
}

// TestListUsers_EmptyList 测试列出用户但列表为空
func TestListUsers_EmptyList(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctxMatcher(), 1, 10).Return([]model.User{}, 0, nil)

	users, total, err := authService.ListUsers(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, users)
	mockRepo.AssertExpectations(t)
}

// TestListUsers_DatabaseError 测试列出用户时数据库错误
func TestListUsers_DatabaseError(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctxMatcher(), 1, 10).Return([]model.User(nil), 0, fmt.Errorf("database error"))

	users, total, err := authService.ListUsers(ctx, 1, 10)

	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, users)
	mockRepo.AssertExpectations(t)
}

// TestListUsers_SecondPage 测试列出第二页用户
func TestListUsers_SecondPage(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	expectedUsers := []model.User{
		{ID: 11, Username: "user11", Email: "user11@test.com", Role: "user"},
	}

	mockRepo.On("List", ctxMatcher(), 2, 10).Return(expectedUsers, 15, nil)

	users, total, err := authService.ListUsers(ctx, 2, 10)

	assert.NoError(t, err)
	assert.Equal(t, 15, total)
	assert.Len(t, users, 1)
	mockRepo.AssertExpectations(t)
}

// ============================================
// DeleteUser 方法测试
// ============================================

// TestDeleteUser_Success 普通用户删除成功
func TestDeleteUser_Success(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	user := &model.User{
		ID:       2,
		Username: "normaluser",
		Role:     "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 2).Return(user, nil)
	mockRepo.On("Delete", ctxMatcher(), 2).Return(nil)

	err := authService.DeleteUser(ctx, 2)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_UserNotFound 测试删除不存在的用户
func TestDeleteUser_UserNotFound(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctxMatcher(), 999).Return(nil, fmt.Errorf("user not found"))

	err := authService.DeleteUser(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_LastAdmin 测试删除最后一个管理员（应失败）
func TestDeleteUser_LastAdmin(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	adminUser := &model.User{
		ID:       1,
		Username: "admin",
		Role:     "admin",
	}

	// 返回管理员用户
	mockRepo.On("GetByID", ctxMatcher(), 1).Return(adminUser, nil)
	// 返回用户列表中只有一个管理员
	mockRepo.On("List", ctxMatcher(), 1, 100).Return([]model.User{
		{ID: 1, Username: "admin", Role: "admin"},
		{ID: 2, Username: "user1", Role: "user"},
	}, 2, nil)

	err := authService.DeleteUser(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete the last admin")
	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_AdminWithOtherAdmins 测试删除非最后一个管理员（应成功）
func TestDeleteUser_AdminWithOtherAdmins(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	adminUser := &model.User{
		ID:       1,
		Username: "admin1",
		Role:     "admin",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(adminUser, nil)
	// 返回多个管理员
	mockRepo.On("List", ctxMatcher(), 1, 100).Return([]model.User{
		{ID: 1, Username: "admin1", Role: "admin"},
		{ID: 2, Username: "admin2", Role: "admin"},
		{ID: 3, Username: "user1", Role: "user"},
	}, 3, nil)
	mockRepo.On("Delete", ctxMatcher(), 1).Return(nil)

	err := authService.DeleteUser(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_DeleteFails 测试删除操作数据库失败
func TestDeleteUser_DeleteFails(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	user := &model.User{
		ID:       2,
		Username: "normaluser",
		Role:     "user",
	}

	mockRepo.On("GetByID", ctxMatcher(), 2).Return(user, nil)
	mockRepo.On("Delete", ctxMatcher(), 2).Return(fmt.Errorf("database constraint error"))

	err := authService.DeleteUser(ctx, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_LastAdmin_ListError 测试管理员删除时 List 出错（应允许删除）
func TestDeleteUser_LastAdmin_ListError(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo)
	ctx := context.Background()

	adminUser := &model.User{
		ID:       1,
		Username: "admin",
		Role:     "admin",
	}

	mockRepo.On("GetByID", ctxMatcher(), 1).Return(adminUser, nil)
	// List 返回错误，此时不会进行管理员计数检查，允许删除
	mockRepo.On("List", ctxMatcher(), 1, 100).Return([]model.User(nil), 0, fmt.Errorf("list error"))
	mockRepo.On("Delete", ctxMatcher(), 1).Return(nil)

	err := authService.DeleteUser(ctx, 1)

	// List 出错时，管理员保护逻辑被跳过，删除成功
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// ============================================
// 辅助函数
// ============================================

// getExpiredRegisteredClaims 返回一组已过期的 JWT RegisteredClaims
func getExpiredRegisteredClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-4 * time.Hour)),
		Issuer:    TokenIssuer,
		Subject:   "user:1",
		ID:        "expired-token-id",
	}
}

// signTestToken 使用全局 JWT 密钥对 Claims 进行签名，返回 token 字符串
func signTestToken(claims Claims) string {
	secret := GetJWTService().GetSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		panic("signTestToken: failed to sign token: " + err.Error())
	}
	return tokenString
}
