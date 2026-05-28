package repository

import (
	"context"

	"database/sql"
	"errors"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

// Additional error definitions for role_repo
var ErrRoleAlreadyExists = errors.New("role already exists")

// RoleRepo handles role data access
type RoleRepo struct {
	db database.QueryExecutor
}

// NewRoleRepo creates a new role repository
func NewRoleRepo(db database.QueryExecutor) *RoleRepo {
	return &RoleRepo{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *RoleRepo) WithTx(tx database.TransactionInterface) *RoleRepo {
	return &RoleRepo{db: tx}
}

// Create inserts a new role
func (r *RoleRepo) Create(ctx context.Context, role *model.Role) error {
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	query := `
		INSERT INTO roles (name, display_name, description, tenant_id, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		role.Name,
		role.Description,
		role.TenantID,
		role.IsSystem,
		role.CreatedAt,
		role.UpdatedAt,
	).Scan(&role.ID)
}

// GetByID retrieves a role by ID
func (r *RoleRepo) GetByID(ctx context.Context, id int) (*model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1
	`
	role := &model.Role{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.TenantID,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetByName retrieves a role by name within a tenant
func (r *RoleRepo) GetByName(ctx context.Context, tenantID, name string) (*model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2
	`
	role := &model.Role{}
	err := r.db.QueryRow(ctx, query, tenantID, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.TenantID,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

// ListByTenant retrieves all roles for a tenant
func (r *RoleRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.Role, error) {
	query := `
		SELECT id, name, description, tenant_id, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 OR tenant_id = ''
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		role := model.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.TenantID,
			&role.IsSystem,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	// Check for errors during rows iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

// Update updates a role
func (r *RoleRepo) Update(ctx context.Context, role *model.Role) error {
	role.UpdatedAt = time.Now()
	query := `
		UPDATE roles
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND is_system = false
	`
	result, err := r.db.Exec(ctx, query,
		role.ID,
		role.Name,
		role.Description,
		role.UpdatedAt,
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

// Delete removes a role by ID
func (r *RoleRepo) Delete(ctx context.Context, id int) error {
	// Check if it's a system role
	role, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	// Delete role permissions first
	_, err = r.db.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1", id)
	if err != nil {
		return err
	}

	// Delete user role assignments
	_, err = r.db.Exec(ctx, "DELETE FROM user_roles WHERE role_id = $1", id)
	if err != nil {
		return err
	}

	// Delete the role
	result, err := r.db.Exec(ctx, "DELETE FROM roles WHERE id = $1", id)
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

// AssignRoleToUser assigns a role to a user
func (r *RoleRepo) AssignRoleToUser(ctx context.Context, userID, roleID int, tenantID string) error {
	// Check if already assigned
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		userID, roleID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already assigned
	}

	query := `
		INSERT INTO user_roles (user_id, role_id, tenant_id, assigned_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.Exec(ctx, query, userID, roleID, tenantID, time.Now())
	return err
}

// RemoveRoleFromUser removes a role from a user
func (r *RoleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	result, err := r.db.Exec(ctx,
		"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
		userID, roleID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("role assignment not found")
	}
	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *RoleRepo) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.tenant_id, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		role := model.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.TenantID,
			&role.IsSystem,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	// Check for errors during rows iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *RoleRepo) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		perm := model.Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	// Check for errors during rows iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *RoleRepo) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, roleID, permissionID)
	return err
}

// RemovePermissionFromRole removes a permission from a role
func (r *RoleRepo) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	result, err := r.db.Exec(ctx,
		"DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2",
		roleID, permissionID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("permission assignment not found")
	}
	return nil
}

// GetUserPermissions retrieves all permissions for a user through their roles
func (r *RoleRepo) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		perm := model.Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	// Check for errors during rows iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

// CheckUserPermission checks if a user has a specific permission
func (r *RoleRepo) CheckUserPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3
		)
	`
	var hasPermission bool
	err := r.db.QueryRow(ctx, query, userID, resource, action).Scan(&hasPermission)
	if err != nil {
		return false, err
	}
	return hasPermission, nil
}

// GetByIDWithPermissions retrieves a role with its permissions
func (r *RoleRepo) GetByIDWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	role, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	permissions, err := r.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.RoleResponse{
		Role:        *role,
		Permissions: permissions,
	}, nil
}
