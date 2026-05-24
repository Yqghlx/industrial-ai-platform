package service

import (
	"context"
	"fmt"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/errors"
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
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
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
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
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
func (s *AuthService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
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
func (s *AuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
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
		fmt.Printf("Warning: failed to update token version: %v\n", err)
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
