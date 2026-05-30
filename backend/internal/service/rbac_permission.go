package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	pkgerrors "github.com/industrial-ai/platform/pkg/errors"
)

// AssignPermissionToRole 分配权限给角色
func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	// 验证角色存在
	if _, err := s.rbacRepo.GetRoleByID(ctx, roleID); err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Role not found", "")
		}
		return fmt.Errorf("failed to verify role: %w", err)
	}

	// 验证权限存在
	if _, err := s.rbacRepo.GetPermissionByID(ctx, permissionID); err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Permission not found", "")
		}
		return fmt.Errorf("failed to verify permission: %w", err)
	}

	return s.rbacRepo.AssignPermissionToRole(ctx, roleID, permissionID)
}

// RemovePermissionFromRole 移除角色的权限
func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	return s.rbacRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

// GetRolePermissions 获取角色的所有权限
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	return s.rbacRepo.GetRolePermissions(ctx, roleID)
}

// CreatePermission 创建新权限
func (s *RBACService) CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error) {
	permission := &model.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
	}

	if err := s.rbacRepo.CreatePermission(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	return permission, nil
}

// GetPermission 通过 ID 获取权限
func (s *RBACService) GetPermission(ctx context.Context, id int) (*model.Permission, error) {
	perm, err := s.rbacRepo.GetPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Permission not found", "")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return perm, nil
}

// ListPermissions 获取所有权限
func (s *RBACService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	permissions, err := s.rbacRepo.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

// ListPermissionsByResource 按资源过滤权限
func (s *RBACService) ListPermissionsByResource(ctx context.Context, resource string) ([]model.Permission, error) {
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

// DeletePermission 删除权限
func (s *RBACService) DeletePermission(ctx context.Context, id int) error {
	if err := s.rbacRepo.DeletePermission(ctx, id); err != nil {
		if errors.Is(err, repository.ErrPermissionNotFound) {
			return pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Permission not found", "")
		}
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// SeedDefaultRolesAndPermissions 为租户创建默认角色和权限
func (s *RBACService) SeedDefaultRolesAndPermissions(ctx context.Context, tenantID string) error {
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
		_, err := s.CreateRole(ctx, &model.Role{
			Name:        role.name,
			Description: role.description,
			TenantID:    tenantID,
		})
		if err != nil {
			continue // 角色已存在则跳过
		}
	}

	return s.SeedPermissionsOnly(ctx)
}

// SeedPermissionsOnly 创建默认权限
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
			continue // 权限已存在则跳过
		}
	}

	return nil
}

// GetUserPermissionStrings 获取用户权限字符串列表（格式: resource:action）
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

// HasSystemPermission 检查用户是否有系统级权限
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

// GetUserRoleNames 获取用户角色名称列表
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

// IsAdmin 检查用户是否拥有管理员角色
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
