package service

import (
	"context"
	"errors"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// AuthService handles authentication
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Verify password
	if !VerifyPassword(password, user.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	// Check if username exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, "", errors.New("username already exists")
	}

	// Check if email exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, "", errors.New("email already exists")
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
