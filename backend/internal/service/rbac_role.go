package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	pkgerrors "github.com/industrial-ai/platform/pkg/errors"
)

// RBACService 处理基于角色的访问控制业务逻辑
// 所有数据访问通过 RBACRepositoryInterface 完成，便于测试和替换
type RBACService struct {
	rbacRepo repository.RBACRepositoryInterface
}

// NewRBACService 创建 RBAC 服务
func NewRBACService(rbacRepo repository.RBACRepositoryInterface) *RBACService {
	return &RBACService{
		rbacRepo: rbacRepo,
	}
}

// CreateRole 创建新角色
func (s *RBACService) CreateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// 检查角色是否已存在
	existing, err := s.rbacRepo.GetRoleByName(ctx, role.Name)
	if err == nil && existing != nil {
		return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeConflict, "Role already exists", "")
	}

	role.IsSystem = false
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if err := s.rbacRepo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRoleByID 通过 ID 获取角色
func (s *RBACService) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	role, err := s.rbacRepo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// GetRole 是 GetRoleByID 的别名，保持向后兼容
func (s *RBACService) GetRole(ctx context.Context, id int) (*model.Role, error) {
	return s.GetRoleByID(ctx, id)
}

// GetRoleWithPermissions 获取角色及其权限
func (s *RBACService) GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	role, err := s.rbacRepo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	permissions, err := s.rbacRepo.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return &model.RoleResponse{
		Role:        *role,
		Permissions: permissions,
	}, nil
}

// ListRoles 获取所有角色
func (s *RBACService) ListRoles(ctx context.Context) ([]model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	roles, err := s.rbacRepo.ListRoles(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// ListRolesByTenant 获取指定租户的角色
func (s *RBACService) ListRolesByTenant(ctx context.Context, tenantID string) ([]model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	roles, err := s.rbacRepo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// UpdateRole 更新角色
func (s *RBACService) UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// 验证角色存在
	if _, err := s.rbacRepo.GetRoleByID(ctx, role.ID); err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role.UpdatedAt = time.Now()

	if err := s.rbacRepo.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(ctx context.Context, id int) error {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	role, err := s.rbacRepo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return pkgerrors.NewAppError(pkgerrors.ErrCodeForbidden, "Cannot delete system role", "")
	}

	if err := s.rbacRepo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignRoleToUser 分配角色给用户
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
	return s.AssignRole(ctx, userID, roleID, "")
}

// AssignRole 分配角色给用户（带租户上下文）
func (s *RBACService) AssignRole(ctx context.Context, userID, roleID int, tenantID string) error {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	if _, err := s.rbacRepo.GetRoleByID(ctx, roleID); err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	if err := s.rbacRepo.AssignRoleToUser(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRoleFromUser 移除用户的角色
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	if err := s.rbacRepo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

// ListUserRoles 获取用户的所有角色
func (s *RBACService) ListUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	return s.GetUserRoles(ctx, userID)
}

// GetUserRoles 获取用户的所有角色
func (s *RBACService) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	roles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *RBACService) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	permissions, err := s.rbacRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// CheckPermission 检查用户是否有指定权限
func (s *RBACService) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	hasPermission, err := s.rbacRepo.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return hasPermission, nil
}

// HasAnyPermission 检查用户是否有任一指定权限
func (s *RBACService) HasAnyPermission(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	allPermissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	permSet := make(map[string]bool, len(allPermissions))
	for _, perm := range allPermissions {
		permSet[fmt.Sprintf("%s:%s", perm.Resource, perm.Action)] = true
	}

	for _, perm := range permissions {
		if permSet[fmt.Sprintf("%s:%s", perm.Resource, perm.Action)] {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions 检查用户是否拥有所有指定权限
func (s *RBACService) HasAllPermissions(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	allPermissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	permSet := make(map[string]bool, len(allPermissions))
	for _, perm := range allPermissions {
		permSet[fmt.Sprintf("%s:%s", perm.Resource, perm.Action)] = true
	}

	for _, perm := range permissions {
		if !permSet[fmt.Sprintf("%s:%s", perm.Resource, perm.Action)] {
			return false, nil
		}
	}
	return true, nil
}
