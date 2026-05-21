package service

import (
	"context"
	"errors"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// UserService 用户服务 (用于 AuthHandler)
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Authenticate 验证用户登录
func (s *UserService) Authenticate(username, password string) (*model.User, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 使用 PasswordHash 或 Password 字段进行验证
	hash := user.PasswordHash
	if hash == "" {
		hash = user.Password
	}

	if !VerifyPassword(password, hash) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id int) (*model.User, error) {
	ctx := context.Background()
	return s.userRepo.GetByID(ctx, id)
}

// GetTokenVersion 获取用户的 Token 版本号
func (s *UserService) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	return s.userRepo.GetTokenVersion(ctx, userID)
}

// UpdatePassword 更新用户密码
func (s *UserService) UpdatePassword(id int, passwordHash string) error {
	ctx := context.Background()
	return s.userRepo.UpdatePassword(ctx, id, passwordHash)
}

// UpdateTokenVersion 递增用户的 Token 版本号
func (s *UserService) UpdateTokenVersion(ctx context.Context, userID int) error {
	return s.userRepo.UpdateTokenVersion(ctx, userID)
}
