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

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.NewInternalError(err.Error())
	}

	return user, token, nil
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
