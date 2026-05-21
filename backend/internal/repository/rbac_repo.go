package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/industrial-ai/platform/internal/model"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrRoleAssigned       = errors.New("role already assigned to user")
	ErrPermissionAssigned = errors.New("permission already assigned to role")
)

// RBACRepository handles RBAC data access
type RBACRepository struct {
	db *sql.DB
}

// NewRBACRepository creates a new RBAC repository
func NewRBACRepository(db *sql.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// CreateRole creates a new role
func (r *RBACRepository) CreateRole(ctx context.Context, role *model.Role) error {
	query := `
		INSERT INTO roles (name, description, tenant_id, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		role.Name, role.Description, role.TenantID, role.IsSystem,
		role.CreatedAt, role.UpdatedAt,
	).Scan(&role.ID)
}

// GetRoleByID retrieves a role by ID
func (r *RBACRepository) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles WHERE id = $1
	`
	role := &model.Role{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.TenantID,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *RBACRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles WHERE name = $1
	`
	role := &model.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.TenantID,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

// ListRoles retrieves all roles
func (r *RBACRepository) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 OR tenant_id = '' OR tenant_id IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.TenantID,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// UpdateRole updates a role
func (r *RBACRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	query := `
		UPDATE roles SET
			name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND is_system = false
	`
	result, err := r.db.ExecContext(ctx, query,
		role.Name, role.Description, role.UpdatedAt, role.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRoleNotFound
	}
	return nil
}

// DeleteRole deletes a role (only non-system roles)
// This operation is idempotent - deleting a non-existent role returns no error
func (r *RBACRepository) DeleteRole(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = false`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// AssignRoleToUser assigns a role to a user
func (r *RBACRepository) AssignRoleToUser(ctx context.Context, userID, roleID int, tenantID string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, tenant_id, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query, userID, roleID, tenantID, time.Now())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRoleAssigned
	}
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// GetUserRoles retrieves all roles for a user
func (r *RBACRepository) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.tenant_id, r.is_system, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.TenantID,
			&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// CreatePermission creates a new permission
func (r *RBACRepository) CreatePermission(ctx context.Context, perm *model.Permission) error {
	query := `
		INSERT INTO permissions (name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		perm.Name, perm.Resource, perm.Action, perm.Description, perm.CreatedAt,
	).Scan(&perm.ID)
}

// GetPermissionByID retrieves a permission by ID
func (r *RBACRepository) GetPermissionByID(ctx context.Context, id int) (*model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions WHERE id = $1
	`
	perm := &model.Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// ListPermissions retrieves all permissions
func (r *RBACRepository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions ORDER BY resource, action
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		if err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt,
		); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *RBACRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPermissionAssigned
	}
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *RBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}

// GetRolePermissions retrieves all permissions for a role
func (r *RBACRepository) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		if err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt,
		); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (r *RBACRepository) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3
	`
	var hasPermission bool
	err := r.db.QueryRowContext(ctx, query, userID, resource, action).Scan(&hasPermission)
	if err != nil {
		return false, err
	}
	return hasPermission, nil
}

// GetUserPermissions retrieves all permissions for a user through their roles
func (r *RBACRepository) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var perm model.Permission
		if err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt,
		); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// InitializeDefaultRBAC creates default roles and permissions
func (r *RBACRepository) InitializeDefaultRBAC(ctx context.Context) error {
	// Create default permissions
	for _, perm := range model.DefaultPermissions {
		_, err := r.GetPermissionByID(ctx, perm.ID)
		if err == ErrPermissionNotFound {
			permCopy := perm
			permCopy.CreatedAt = time.Now()
			if err := r.CreatePermission(ctx, &permCopy); err != nil {
				return err
			}
		}
	}

	// Create default roles
	for _, role := range model.DefaultRoles {
		_, err := r.GetRoleByID(ctx, role.ID)
		if err == ErrRoleNotFound {
			roleCopy := role
			roleCopy.CreatedAt = time.Now()
			roleCopy.UpdatedAt = time.Now()
			if err := r.CreateRole(ctx, &roleCopy); err != nil {
				return err
			}
		}
	}

	// Assign default permissions to admin role (role_id = 1)
	adminPermissions := []int{1, 3, 5, 7, 8, 9, 11, 13, 14}
	for _, permID := range adminPermissions {
		r.AssignPermissionToRole(ctx, 1, permID)
	}

	// Assign permissions to operator role (role_id = 2)
	operatorPermissions := []int{1, 2, 5, 6, 7, 9, 10, 11, 12}
	for _, permID := range operatorPermissions {
		r.AssignPermissionToRole(ctx, 2, permID)
	}

	// Assign permissions to viewer role (role_id = 3)
	viewerPermissions := []int{2, 4, 6, 7, 10, 12}
	for _, permID := range viewerPermissions {
		r.AssignPermissionToRole(ctx, 3, permID)
	}

	return nil
}
