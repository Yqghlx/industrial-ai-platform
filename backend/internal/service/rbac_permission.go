package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	pkgerrors "github.com/industrial-ai/platform/pkg/errors"
)

// AssignPermissionToRole assigns a permission to a role
func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	var err error
	if s.roleRepo != nil {
		_, err = s.roleRepo.GetByID(roleID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetRoleByID(ctx, roleID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	if s.permRepo != nil {
		_, err = s.permRepo.GetByID(permissionID)
	} else if s.rbacRepo != nil {
		_, err = s.rbacRepo.GetPermissionByID(ctx, permissionID)
	}

	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Permission not found", "")
		}
		return fmt.Errorf("failed to verify permission: %w", err)
	}

	if s.rbacRepo != nil {
		return s.rbacRepo.AssignPermissionToRole(ctx, roleID, permissionID)
	}

	return pkgerrors.NewAppError(pkgerrors.ErrCodeService, "RBAC repository not initialized", "")
}

// RemovePermissionFromRole removes a permission from a role
func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	if s.rbacRepo != nil {
		return s.rbacRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
	}
	return pkgerrors.NewAppError(pkgerrors.ErrCodeService, "RBAC repository not initialized", "")
}

// GetRolePermissions retrieves all permissions for a role
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	if s.rbacRepo != nil {
		return s.rbacRepo.GetRolePermissions(ctx, roleID)
	}
	return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeService, "RBAC repository not initialized", "")
}

// CreatePermission creates a new permission
func (s *RBACService) CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error) {
	permission := &model.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
	}

	if s.permRepo != nil {
		err := s.permRepo.Create(permission)
		if err != nil {
			return nil, fmt.Errorf("failed to create permission: %w", err)
		}
		return permission, nil
	}

	if s.rbacRepo != nil {
		err := s.rbacRepo.CreatePermission(ctx, permission)
		if err != nil {
			return nil, fmt.Errorf("failed to create permission: %w", err)
		}
		return permission, nil
	}

	return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeService, "Permission repository not initialized", "")
}

// GetPermission retrieves a permission by ID
func (s *RBACService) GetPermission(ctx context.Context, id int) (*model.Permission, error) {
	if s.permRepo != nil {
		return s.permRepo.GetByID(id)
	}
	if s.rbacRepo != nil {
		return s.rbacRepo.GetPermissionByID(ctx, id)
	}
	return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeService, "Permission repository not initialized", "")
}

// ListPermissions lists all permissions
func (s *RBACService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	if s.permRepo != nil {
		return s.permRepo.List()
	}
	if s.rbacRepo != nil {
		return s.rbacRepo.ListPermissions(ctx)
	}
	return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeService, "Permission repository not initialized", "")
}

// ListPermissionsByResource lists permissions by resource
func (s *RBACService) ListPermissionsByResource(ctx context.Context, resource string) ([]model.Permission, error) {
	if s.permRepo != nil {
		return s.permRepo.ListByResource(resource)
	}
	// Fallback: filter from all permissions
	permissions, err := s.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	var result []model.Permission
	for _, perm := range permissions {
		if perm.Resource == resource {
			result = append(result, perm)
		}
	}
	return result, nil
}

// DeletePermission deletes a permission
func (s *RBACService) DeletePermission(ctx context.Context, id int) error {
	if s.permRepo != nil {
		return s.permRepo.Delete(id)
	}
	// RBACRepository doesn't have DeletePermission
	return pkgerrors.NewAppError(pkgerrors.ErrCodeService, "Permission repository not initialized", "")
}

// SeedDefaultRolesAndPermissions seeds default roles and permissions for a tenant
func (s *RBACService) SeedDefaultRolesAndPermissions(ctx context.Context, tenantID string) error {
	if s.rbacRepo == nil {
		return pkgerrors.NewAppError(pkgerrors.ErrCodeService, "RBAC repository not initialized", "")
	}

	// Create default roles
	defaultRoles := []struct {
		name        string
		displayName string
		description string
	}{
		{"admin", "Administrator", "Full system access"},
		{"operator", "Operator", "Device monitoring and control"},
		{"viewer", "Viewer", "Read-only access"},
	}

	for _, role := range defaultRoles {
		_, err := s.CreateRole(ctx, tenantID, role.name, role.displayName, role.description)
		if err != nil {
			// Ignore if role already exists
			continue
		}
	}

	// Seed permissions
	return s.SeedPermissionsOnly(ctx)
}

// SeedPermissionsOnly seeds default permissions
func (s *RBACService) SeedPermissionsOnly(ctx context.Context) error {
	defaultPermissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"device_read", "devices", "read", "View devices"},
		{"device_write", "devices", "write", "Create/update devices"},
		{"device_delete", "devices", "delete", "Delete devices"},
		{"alert_read", "alerts", "read", "View alerts"},
		{"alert_write", "alerts", "write", "Create/update alerts"},
		{"alert_delete", "alerts", "delete", "Delete alerts"},
		{"rule_read", "rules", "read", "View alert rules"},
		{"rule_write", "rules", "write", "Create/update alert rules"},
		{"rule_delete", "rules", "delete", "Delete alert rules"},
		{"report_read", "reports", "read", "View reports"},
		{"report_write", "reports", "write", "Generate reports"},
		{"user_read", "users", "read", "View users"},
		{"user_write", "users", "write", "Create/update users"},
		{"role_read", "roles", "read", "View roles"},
		{"role_write", "roles", "write", "Create/update roles"},
	}

	for _, perm := range defaultPermissions {
		_, err := s.CreatePermission(ctx, perm.name, perm.resource, perm.action, perm.description)
		if err != nil {
			continue // Ignore if permission already exists
		}
	}

	return nil
}

// GetUserPermissionStrings retrieves user permissions as strings
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

// HasSystemPermission checks if user has any system-level permission
func (s *RBACService) HasSystemPermission(ctx context.Context, userID int) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Resource == "system" {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoleNames retrieves user role names
func (s *RBACService) GetUserRoleNames(ctx context.Context, userID int) ([]string, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
	}
	return names, nil
}

// IsAdmin checks if user has admin role
func (s *RBACService) IsAdmin(ctx context.Context, userID int) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		if role.Name == "admin" || role.Name == "administrator" {
			return true, nil
		}
	}
	return false, nil
}