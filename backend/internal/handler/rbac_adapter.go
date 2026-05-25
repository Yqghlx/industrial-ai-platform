package handler

import (
	"context"
	"errors"

	"github.com/industrial-ai/platform/internal/model"
)

// ErrNotImplemented is returned when a method is not implemented in the adapter
var ErrNotImplemented = errors.New("method not implemented in RBAC adapter")

// rbacServiceAdapter adapts service.RBACServiceInterface to handler.RBACServiceInterface
// This adapter bridges the interface gap between the service layer and handler layer
type rbacServiceAdapter struct {
	svc rbacServiceInternal
}

// rbacServiceInternal is an alias for the service layer interface
// We use a local interface to avoid import cycles
type rbacServiceInternal interface {
	CreateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	DeleteRole(ctx context.Context, id int) error
	GetRoleByID(ctx context.Context, id int) (*model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	ListUserRoles(ctx context.Context, userID int) ([]model.Role, error)
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permID int) error
}

// NewRBACServiceAdapter creates an adapter that wraps service.RBACServiceInterface
// to implement handler.RBACServiceInterface
func NewRBACServiceAdapter(svc rbacServiceInternal) RBACServiceInterface {
	return &rbacServiceAdapter{svc: svc}
}

func (a *rbacServiceAdapter) CreateRole(ctx context.Context, tenantID, name, displayName, description string) (*model.Role, error) {
	// Note: displayName is not supported in model.Role, use name instead
	role := &model.Role{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}
	return a.svc.CreateRole(ctx, role)
}

func (a *rbacServiceAdapter) GetRole(ctx context.Context, id int) (*model.Role, error) {
	return a.svc.GetRoleByID(ctx, id)
}

func (a *rbacServiceAdapter) GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	role, err := a.svc.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert to RoleResponse - RoleResponse contains Role and Permissions
	return &model.RoleResponse{
		Role:        *role,
		Permissions: []model.Permission{}, // Permissions not available via this adapter
	}, nil
}

func (a *rbacServiceAdapter) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	// Service layer ListRoles doesn't filter by tenantID, filter in memory if needed
	roles, err := a.svc.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by tenantID if specified
	if tenantID == "" {
		return roles, nil
	}

	filtered := make([]model.Role, 0)
	for _, role := range roles {
		if role.TenantID == tenantID {
			filtered = append(filtered, role)
		}
	}
	return filtered, nil
}

func (a *rbacServiceAdapter) UpdateRole(ctx context.Context, id int, updates map[string]interface{}) error {
	// Get existing role first
	role, err := a.svc.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		role.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		role.Description = description
	}

	_, err = a.svc.UpdateRole(ctx, role)
	return err
}

func (a *rbacServiceAdapter) DeleteRole(ctx context.Context, id int) error {
	return a.svc.DeleteRole(ctx, id)
}

func (a *rbacServiceAdapter) AssignRole(ctx context.Context, userID, roleID int, tenantID string) error {
	// Note: tenantID is ignored as service layer doesn't support it
	return a.svc.AssignRoleToUser(ctx, userID, roleID)
}

func (a *rbacServiceAdapter) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	return a.svc.RemoveRoleFromUser(ctx, userID, roleID)
}

func (a *rbacServiceAdapter) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	return a.svc.ListUserRoles(ctx, userID)
}

func (a *rbacServiceAdapter) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	// Service layer doesn't have GetUserPermissions method
	// Return empty list as placeholder
	return []model.Permission{}, nil
}

func (a *rbacServiceAdapter) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	// Service layer doesn't have CheckPermission method
	// Return false with nil error as placeholder
	return false, nil
}

func (a *rbacServiceAdapter) CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error) {
	// Service layer doesn't have CreatePermission method
	// Return nil with error indicating not implemented
	return nil, ErrNotImplemented
}

func (a *rbacServiceAdapter) GetPermission(ctx context.Context, id int) (*model.Permission, error) {
	// Service layer doesn't have GetPermission method
	// Return nil with error indicating not implemented
	return nil, ErrNotImplemented
}

func (a *rbacServiceAdapter) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	return a.svc.ListPermissions(ctx)
}

func (a *rbacServiceAdapter) DeletePermission(ctx context.Context, id int) error {
	// Service layer doesn't have DeletePermission method
	return ErrNotImplemented
}

func (a *rbacServiceAdapter) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	return a.svc.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (a *rbacServiceAdapter) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	return a.svc.RemovePermissionFromRole(ctx, roleID, permissionID)
}

func (a *rbacServiceAdapter) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	// Service layer doesn't have GetRolePermissions method
	return []model.Permission{}, nil
}