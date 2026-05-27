package service

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/constants"
	"github.com/industrial-ai/platform/pkg/errors"
)

// UserService 用户服务 (用于 AuthHandler)
type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

// Authenticate 验证用户登录
// FIX-019: 添加 Context 超时设置
func (s *UserService) Authenticate(username, password string) (*model.User, error) {
	// FIX-019: 使用带超时的 context，防止数据库操作无限等待
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(constants.DefaultServiceTimeoutSec)*time.Second)
	defer cancel()

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.NewAuthFailedError()
	}

	// 使用 PasswordHash 或 Password 字段进行验证
	hash := user.PasswordHash
	if hash == "" {
		hash = user.Password
	}

	if !VerifyPassword(password, hash) {
		return nil, errors.NewAuthFailedError()
	}

	return user, nil
}

// GetByID 根据 ID 获取用户
// FIX-019: 添加 Context 超时设置
func (s *UserService) GetByID(id int) (*model.User, error) {
	// FIX-019: 使用带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(constants.DefaultServiceTimeoutSec)*time.Second)
	defer cancel()
	return s.userRepo.GetByID(ctx, id)
}

// GetTokenVersion 获取用户的 Token 版本号
func (s *UserService) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	return s.userRepo.GetTokenVersion(ctx, userID)
}

// UpdatePassword 更新用户密码
// FIX-019: 添加 Context 超时设置
func (s *UserService) UpdatePassword(id int, passwordHash string) error {
	// FIX-019: 使用带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(constants.DefaultServiceTimeoutSec)*time.Second)
	defer cancel()
	return s.userRepo.UpdatePassword(ctx, id, passwordHash)
}

// UpdateTokenVersion 递增用户的 Token 版本号
func (s *UserService) UpdateTokenVersion(ctx context.Context, userID int) error {
	return s.userRepo.UpdateTokenVersion(ctx, userID)
}
