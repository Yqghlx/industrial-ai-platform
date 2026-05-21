package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
)

// RBACRepositoryInterface defines the interface for RBAC repository
type RBACRepositoryInterface interface {
	CreateRole(ctx context.Context, role *model.Role) error
	GetRoleByID(ctx context.Context, id int) (*model.Role, error)
	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	ListRoles(ctx context.Context, tenantID string) ([]model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id int) error
	AssignRoleToUser(ctx context.Context, userID, roleID int, tenantID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	GetUserRoles(ctx context.Context, userID int) ([]model.Role, error)
	CreatePermission(ctx context.Context, perm *model.Permission) error
	GetPermissionByID(ctx context.Context, id int) (*model.Permission, error)
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error)
	CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error)
	InitializeDefaultRBAC(ctx context.Context) error
}
