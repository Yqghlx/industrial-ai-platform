package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrRoleAlreadyExists      = errors.New("role already exists")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
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
func (s *RBACService) CreateRole(ctx context.Context, tenantID, name, displayName, description string) (*model.Role, error) {
	// Check if role already exists
	if s.roleRepo != nil {
		existing, err := s.roleRepo.GetByName(tenantID, name)
		if err == nil && existing != nil {
			return nil, ErrRoleAlreadyExists
		}
	} else if s.rbacRepo != nil {
		existing, err := s.rbacRepo.GetRoleByName(ctx, name)
		if err == nil && existing != nil {
			return nil, ErrRoleAlreadyExists
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
		err = s.roleRepo.Create(role)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.CreateRole(ctx, role)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (s *RBACService) GetRole(ctx context.Context, id int) (*model.Role, error) {
	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// GetRoleWithPermissions retrieves a role with its permissions
func (s *RBACService) GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	var result *model.RoleResponse
	var err error

	if s.roleRepo != nil {
		result, err = s.roleRepo.GetByIDWithPermissions(id)
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
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role with permissions: %w", err)
	}
	return result, nil
}

// ListRoles retrieves all roles for a tenant
func (s *RBACService) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	var roles []model.Role
	var err error

	if s.roleRepo != nil {
		roles, err = s.roleRepo.ListByTenant(tenantID)
	} else if s.rbacRepo != nil {
		roles, err = s.rbacRepo.ListRoles(ctx, tenantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// UpdateRole updates a role
func (s *RBACService) UpdateRole(ctx context.Context, id int, updates map[string]interface{}) error {
	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return ErrRoleNotFound
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
		err = s.roleRepo.Update(role)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.UpdateRole(ctx, role)
	}

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// DeleteRole deletes a role by ID
func (s *RBACService) DeleteRole(ctx context.Context, id int) error {
	var role *model.Role
	var err error

	if s.roleRepo != nil {
		role, err = s.roleRepo.GetByID(id)
	} else if s.rbacRepo != nil {
		role, err = s.rbacRepo.GetRoleByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return ErrCannotDeleteSystemRole
	}

	if s.roleRepo != nil {
		err = s.roleRepo.Delete(id)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.DeleteRole(ctx, id)
	}

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user
func (s *RBACService) AssignRole(ctx context.Context, userID, roleID int, tenantID string) error {
	// Verify role exists
	var err error
	if s.roleRepo != nil {
		_, err = s.roleRepo.GetByID(roleID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetRoleByID(ctx, roleID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	if s.roleRepo != nil {
		err = s.roleRepo.AssignRoleToUser(userID, roleID, tenantID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.AssignRoleToUser(ctx, userID, roleID, tenantID)
	}

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	var err error
	if s.roleRepo != nil {
		err = s.roleRepo.RemoveRoleFromUser(userID, roleID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.RemoveRoleFromUser(ctx, userID, roleID)
	}

	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	var roles []model.Role
	var err error

	if s.roleRepo != nil {
		roles, err = s.roleRepo.GetUserRoles(userID)
	} else if s.rbacRepo != nil {
		roles, err = s.rbacRepo.GetUserRoles(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	var permissions []model.Permission
	var err error

	if s.roleRepo != nil {
		permissions, err = s.roleRepo.GetUserPermissions(userID)
	} else if s.rbacRepo != nil {
		permissions, err = s.rbacRepo.GetUserPermissions(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (s *RBACService) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	var hasPermission bool
	var err error

	if s.roleRepo != nil {
		hasPermission, err = s.roleRepo.CheckUserPermission(userID, resource, action)
	} else if s.rbacRepo != nil {
		hasPermission, err = s.rbacRepo.CheckPermission(ctx, userID, resource, action)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return hasPermission, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *RBACService) HasAnyPermission(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
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
func (s *RBACService) HasAllPermissions(ctx context.Context, userID int, permissions []struct {
	Resource string
	Action   string
}) (bool, error) {
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
func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	// Verify role exists
	var err error
	if s.roleRepo != nil {
		_, err = s.roleRepo.GetByID(roleID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetRoleByID(ctx, roleID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	// Verify permission exists
	if s.permRepo != nil {
		_, err = s.permRepo.GetByID(permissionID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetPermissionByID(ctx, permissionID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to verify permission: %w", err)
	}

	if s.roleRepo != nil {
		err = s.roleRepo.AssignPermissionToRole(roleID, permissionID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.AssignPermissionToRole(ctx, roleID, permissionID)
	}

	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	var err error
	if s.roleRepo != nil {
		err = s.roleRepo.RemovePermissionFromRole(roleID, permissionID)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
	}

	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	var permissions []model.Permission
	var err error

	if s.roleRepo != nil {
		permissions, err = s.roleRepo.GetRolePermissions(roleID)
	} else if s.rbacRepo != nil {
		permissions, err = s.rbacRepo.GetRolePermissions(ctx, roleID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return permissions, nil
}

// CreatePermission creates a new permission
func (s *RBACService) CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error) {
	// Check if permission already exists
	var err error
	if s.permRepo != nil {
		existing, err := s.permRepo.GetByResourceAction(resource, action)
		if err == nil && existing != nil {
			return nil, errors.New("permission already exists for this resource and action")
		}
	}

	perm := &model.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if s.permRepo != nil {
		err = s.permRepo.Create(perm)
	} else if s.rbacRepo != nil {
		err = s.rbacRepo.CreatePermission(ctx, perm)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return perm, nil
}

// GetPermission retrieves a permission by ID
func (s *RBACService) GetPermission(ctx context.Context, id int) (*model.Permission, error) {
	var perm *model.Permission
	var err error

	if s.permRepo != nil {
		perm, err = s.permRepo.GetByID(id)
	} else if s.rbacRepo != nil {
		perm, err = s.rbacRepo.GetPermissionByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return perm, nil
}

// ListPermissions retrieves all permissions
func (s *RBACService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	var err error

	if s.permRepo != nil {
		permissions, err = s.permRepo.List()
	} else if s.rbacRepo != nil {
		permissions, err = s.rbacRepo.ListPermissions(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

// ListPermissionsByResource retrieves all permissions for a resource
func (s *RBACService) ListPermissionsByResource(ctx context.Context, resource string) ([]model.Permission, error) {
	var permissions []model.Permission
	var err error

	if s.permRepo != nil {
		permissions, err = s.permRepo.ListByResource(resource)
	} else {
		// Fallback: filter from all permissions
		allPerms, err := s.ListPermissions(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range allPerms {
			if p.Resource == resource {
				permissions = append(permissions, p)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by resource: %w", err)
	}
	return permissions, nil
}

// DeletePermission deletes a permission by ID
func (s *RBACService) DeletePermission(ctx context.Context, id int) error {
	var err error
	if s.permRepo != nil {
		err = s.permRepo.Delete(id)
	}

	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// SeedDefaultRolesAndPermissions seeds the database with default roles and permissions
func (s *RBACService) SeedDefaultRolesAndPermissions(ctx context.Context, tenantID string) error {
	// Use existing RBACRepository's InitializeDefaultRBAC if available
	if s.rbacRepo != nil {
		return s.rbacRepo.InitializeDefaultRBAC(ctx)
	}

	// Seed default permissions using extended list
	defaultPermissions := model.ExtendedDefaultPermissions()
	for i := range defaultPermissions {
		perm := &defaultPermissions[i]
		if s.permRepo != nil {
			if err := s.permRepo.CreateIfNotExists(perm); err != nil {
				return fmt.Errorf("failed to seed permission %s: %w", perm.Name, err)
			}
		}
	}

	// Create default roles based on model.DefaultRoles
	for _, roleDef := range model.DefaultRoles {
		if s.roleRepo != nil {
			// Check if role exists
			existing, err := s.roleRepo.GetByName(tenantID, roleDef.Name)
			if err == nil && existing != nil {
				continue // Role already exists
			}

			role := &model.Role{
				Name:        roleDef.Name,
				Description: roleDef.Description,
				TenantID:    tenantID,
				IsSystem:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := s.roleRepo.Create(role); err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleDef.Name, err)
			}
		}
	}

	return nil
}

// SeedPermissionsOnly seeds only permissions without creating roles
func (s *RBACService) SeedPermissionsOnly(ctx context.Context) error {
	defaultPermissions := model.ExtendedDefaultPermissions()
	for i := range defaultPermissions {
		perm := &defaultPermissions[i]
		if s.permRepo != nil {
			if err := s.permRepo.CreateIfNotExists(perm); err != nil {
				return fmt.Errorf("failed to seed permission %s: %w", perm.Name, err)
			}
		} else if s.rbacRepo != nil {
			if err := s.rbacRepo.CreatePermission(ctx, perm); err != nil {
				return fmt.Errorf("failed to seed permission %s: %w", perm.Name, err)
			}
		}
	}
	return nil
}

// GetUserPermissionStrings returns user permissions as string array for JWT token
func (s *RBACService) GetUserPermissionStrings(ctx context.Context, userID int) ([]string, error) {
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(permissions))
	for i, perm := range permissions {
		result[i] = fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
	}
	return result, nil
}

// GetUserRoleNames returns user role names
func (s *RBACService) GetUserRoleNames(ctx context.Context, userID int) ([]string, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = role.Name
	}
	return result, nil
}

// IsAdmin checks if a user has admin role
func (s *RBACService) IsAdmin(ctx context.Context, userID int) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == model.RoleAdmin {
			return true, nil
		}
	}
	return false, nil
}

// HasSystemPermission checks if a user has system admin permission
func (s *RBACService) HasSystemPermission(ctx context.Context, userID int) (bool, error) {
	return s.CheckPermission(ctx, userID, model.ResourceSystem, model.ActionManage)
}
