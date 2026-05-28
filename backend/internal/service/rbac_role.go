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

// RBACService handles role-based access control business logic
type RBACService struct {
	roleRepo   *repository.RoleRepo
	permRepo   *repository.PermissionRepo
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepo
	rbacRepo   *repository.RBACRepository // Using existing RBACRepository for compatibility
}

// NewRBACService creates a new RBAC service
func NewRBACService(
	roleRepo *repository.RoleRepo,
	permRepo *repository.PermissionRepo,
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepo,
) *RBACService {
	return &RBACService{
		roleRepo:   roleRepo,
		permRepo:   permRepo,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// NewRBACServiceWithRBACRepo creates a new RBAC service with the existing RBACRepository
func NewRBACServiceWithRBACRepo(
	rbacRepo *repository.RBACRepository,
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepo,
) *RBACService {
	return &RBACService{
		rbacRepo:   rbacRepo,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// CreateRole creates a new role
// FIX-019: 添加 Context 超时设置
func (s *RBACService) CreateRole(ctx context.Context, tenantID, name, displayName, description string) (*model.Role, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Check if role already exists
	if s.roleRepo != nil {
		existing, err := s.roleRepo.GetByName(ctx, tenantID, name)
		if err == nil && existing != nil {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeConflict, "Role already exists", "")
		}
	} else if s.rbacRepo != nil {
		existing, err := s.rbacRepo.GetRoleByName(ctx, name)
		if err == nil && existing != nil {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeConflict, "Role already exists", "")
		}
	}

	role := &model.Role{
		Name:        name,
		Description: description,
		TenantID:    tenantID,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var err error
	if s.roleRepo != nil {
		err = s.roleRepo.Create(ctx, role)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.CreateRole(ctx, role)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
// FIX-019: 添加 Context 超时设置
func (s *RBACService) GetRole(ctx context.Context, id int) (*model.Role, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(ctx, id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// GetRoleWithPermissions retrieves a role with its permissions
// FIX-019: 添加 Context 超时设置
func (s *RBACService) GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var result *model.RoleResponse
	var err error

	if s.roleRepo != nil {
		result, err = s.roleRepo.GetByIDWithPermissions(ctx, id)
	} else if s.rbacRepo != nil {
		role, err := s.rbacRepo.GetRoleByID(ctx, id)
		if err != nil {
			return nil, err
		}
		permissions, err := s.rbacRepo.GetRolePermissions(ctx, id)
		if err != nil {
			return nil, err
		}
		result = &model.RoleResponse{
			Role:        *role,
			Permissions: permissions,
		}
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return nil, fmt.Errorf("failed to get role with permissions: %w", err)
	}
	return result, nil
}

// ListRoles retrieves all roles for a tenant
// FIX-019: 添加 Context 超时设置
func (s *RBACService) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var roles []model.Role
	var err error

	if s.roleRepo != nil {
		roles, err = s.roleRepo.ListByTenant(ctx, tenantID)
	} else if s.rbacRepo != nil {
		roles, err = s.rbacRepo.ListRoles(ctx, tenantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// UpdateRole updates a role
// FIX-019: 添加 Context 超时设置
func (s *RBACService) UpdateRole(ctx context.Context, id int, updates map[string]interface{}) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(ctx, id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		role.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		role.Description = description
	}
	role.UpdatedAt = time.Now()

	if s.roleRepo != nil {
		err = s.roleRepo.Update(ctx, role)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.UpdateRole(ctx, role)
	}

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// DeleteRole deletes a role by ID
// FIX-019: 添加 Context 超时设置
func (s *RBACService) DeleteRole(ctx context.Context, id int) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(ctx, id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return pkgerrors.NewAppError(pkgerrors.ErrCodeForbidden, "Cannot delete system role", "")
	}

	if s.roleRepo != nil {
		err = s.roleRepo.Delete(ctx, id)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.DeleteRole(ctx, id)
	}

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user
// FIX-019: 添加 Context 超时设置
func (s *RBACService) AssignRole(ctx context.Context, userID, roleID int, tenantID string) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// Verify role exists
	var err error
	if s.roleRepo != nil {
		_, err = s.roleRepo.GetByID(ctx, roleID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetRoleByID(ctx, roleID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	if s.roleRepo != nil {
		err = s.roleRepo.AssignRoleToUser(ctx, userID, roleID, tenantID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.AssignRoleToUser(ctx, userID, roleID, tenantID)
	}

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
// FIX-019: 添加 Context 超时设置
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var err error
	if s.roleRepo != nil {
		err = s.roleRepo.RemoveRoleFromUser(ctx, userID, roleID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.RemoveRoleFromUser(ctx, userID, roleID)
	}

	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

// GetUserRoles retrieves all roles for a user
// FIX-019: 添加 Context 超时设置
func (s *RBACService) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var roles []model.Role
	var err error

	if s.roleRepo != nil {
		roles, err = s.roleRepo.GetUserRoles(ctx, userID)
	} else if s.rbacRepo != nil {
		roles, err = s.rbacRepo.GetUserRoles(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GetUserPermissions retrieves all permissions for a user
// FIX-019: 添加 Context 超时设置
func (s *RBACService) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var permissions []model.Permission
	var err error

	if s.roleRepo != nil {
		permissions, err = s.roleRepo.GetUserPermissions(ctx, userID)
	} else if s.rbacRepo != nil {
		permissions, err = s.rbacRepo.GetUserPermissions(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
// FIX-019: 添加 Context 超时设置
func (s *RBACService) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	var hasPermission bool
	var err error

	if s.roleRepo != nil {
		hasPermission, err = s.roleRepo.CheckUserPermission(ctx, userID, resource, action)
	} else if s.rbacRepo != nil {
		hasPermission, err = s.rbacRepo.CheckPermission(ctx, userID, resource, action)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return hasPermission, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
// FIX-019: 添加 Context 超时设置
func (s *RBACService) HasAnyPermission(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// FIX-008: N+1 查询优化 - 批量获取所有用户权限
	allPermissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 构建权限集合用于快速查找
	permSet := make(map[string]bool)
	for _, perm := range allPermissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		permSet[key] = true
	}

	// 检查是否有任一权限
	for _, perm := range permissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		if permSet[key] {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
// FIX-019: 添加 Context 超时设置
func (s *RBACService) HasAllPermissions(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
	// FIX-019: 确保 context 有超时
	ctx, cancel := ensureContextTimeout(ctx)
	defer cancel()

	// FIX-008: N+1 查询优化 - 批量获取所有用户权限
	allPermissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 构建权限集合用于快速查找
	permSet := make(map[string]bool)
	for _, perm := range allPermissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		permSet[key] = true
	}

	// 检查是否拥有所有权限
	for _, perm := range permissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		if !permSet[key] {
			return false, nil
		}
	}
	return true, nil
}

// AssignPermissionToRole assigns a permission to a role
