package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/errors"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// AuthService handles authentication
type AuthService struct {
	userRepo repository.UserRepositoryInterface
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Login authenticates a user
// FIX-019: 添加 Context 超时设置
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errors.NewAuthFailedError()
	}

	// Verify password
	if !VerifyPassword(password, user.Password) {
		return nil, "", errors.NewAuthFailedError()
	}

	// Get token version for the user
	tokenVersion := 0
	if tv, err := s.userRepo.GetTokenVersion(ctx, user.ID); err == nil {
		tokenVersion = tv
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.NewInternalError(err.Error())
	}

	// Set user token version for verification
	if tokenVersion > 0 {
		// Use new token generation with version if available
		token, err = GenerateTokenWithVersion(user.ID, user.Username, user.Role, "", tokenVersion)
		if err != nil {
			return nil, "", errors.NewInternalError(err.Error())
		}
	}

	return user, token, nil
}

// GenerateTokenWithVersion generates token with version support
func GenerateTokenWithVersion(userID int, username, role, tenantID string, tokenVersion int) (string, error) {
	if !IsJWTInitialized() {
		return "", fmt.Errorf("JWT not initialized")
	}
	token, _, err := GenerateAccessToken(userID, username, role, tenantID, tokenVersion)
	return token, err
}

// Register creates a new user
// FIX-019: 添加 Context 超时设置
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Check if username exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, "", errors.NewAppError(errors.ErrCodeConflict, "Username already exists", req.Username)
	}

	// Check if email exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, "", errors.NewAppError(errors.ErrCodeConflict, "Email already exists", req.Email)
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", errors.NewInternalError(err.Error())
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", errors.NewDatabaseError(err.Error())
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.NewInternalError(err.Error())
	}

	return user, token, nil
}

// GetUserByID retrieves a user by ID
// FIX-019: 添加 Context 超时设置
func (s *AuthService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewUserNotFoundError(fmt.Sprintf("%d", id))
	}
	return user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if !IsJWTInitialized() {
		return nil, errors.NewInternalError("JWT not initialized")
	}

	tokenPair, err := RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, errors.NewAuthFailedError()
	}

	return tokenPair, nil
}

// ChangePassword changes a user's password
// FIX-019: 添加 Context 超时设置
func (s *AuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewUserNotFoundError(fmt.Sprintf("%d", userID))
	}

	// Verify old password
	hash := user.PasswordHash
	if hash == "" {
		hash = user.Password
	}
	if !VerifyPassword(oldPassword, hash) {
		return errors.NewAuthFailedError()
	}

	// Validate new password
	if err := model.ValidatePassword(newPassword); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// Hash new password
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return errors.NewInternalError(err.Error())
	}

	// Update password in database
	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return errors.NewDatabaseError(err.Error())
	}

	// Revoke all user tokens to force re-login
	if err := s.userRepo.UpdateTokenVersion(ctx, userID); err != nil {
		// Log warning but don't fail the password change
		logger.L().Warn("failed to update token version", zap.Error(err))
	}

	return nil
}

// ValidateToken validates a token and returns the claims
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	if !IsJWTInitialized() {
		return nil, errors.NewInternalError("JWT not initialized")
	}

	claims, err := ParseToken(token)
	if err != nil {
		return nil, errors.NewAuthFailedError()
	}

	return claims, nil
}

// ListUsers 列出所有用户
// FIX-019: 添加 Context 超时设置
func (s *AuthService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()
	return s.userRepo.List(ctx, page, pageSize)
}

// DeleteUser 删除用户
// SEC-HIGH-03: 新增删除用户方法
// FIX-019: 添加 Context 超时设置
func (s *AuthService) DeleteUser(ctx context.Context, userID int) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// 先检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewUserNotFoundError(fmt.Sprintf("%d", userID))
	}

	// 检查是否是最后一个管理员
	if user.Role == "admin" {
		users, _, err := s.userRepo.List(ctx, 1, 100)
		if err == nil {
			adminCount := 0
			for _, u := range users {
				if u.Role == "admin" {
					adminCount++
				}
			}
			if adminCount <= 1 {
				return errors.NewAppError(errors.ErrCodeForbidden, "cannot delete the last admin user", "")
			}
		}
	}

	// 执行删除
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return errors.NewDatabaseError(err.Error())
	}

	return nil
}

// generateRandomPassword 生成加密安全的随机密码
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 密码安全性不可妥协，crypto/rand 失败时直接终止
		logger.L().Fatal("生成安全随机密码失败", zap.Error(err))
	}
	return hex.EncodeToString(bytes)[:length]
}

// EnsureDefaultAdmin 确保默认管理员存在，不存在则创建
func (s *AuthService) EnsureDefaultAdmin(ctx context.Context, password string) error {
	_, err := s.userRepo.GetByUsername(ctx, "admin")
	if err == nil {
		return nil
	}

	if password == "" {
		password = generateRandomPassword(16)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		logger.L().Error("管理员密码哈希失败", zap.Error(err))
		return err
	}

	admin := &model.User{
		Username: "admin",
		Password: passwordHash,
		Email:    "admin@industrial.ai",
		Role:     "admin",
	}

	if err := s.userRepo.Create(ctx, admin); err != nil {
		logger.L().Error("创建默认管理员失败", zap.Error(err))
		return err
	}

	logger.L().Info("已创建默认管理员用户")
	return nil
}
