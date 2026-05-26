package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

// ErrPermissionExists for permission_repo
var ErrPermissionExists = errors.New("permission already exists")

// PermissionRepo handles permission data access
type PermissionRepo struct {
	db database.QueryExecutor
}

// NewPermissionRepo creates a new permission repository
func NewPermissionRepo(db database.QueryExecutor) *PermissionRepo {
	return &PermissionRepo{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *PermissionRepo) WithTx(tx database.TransactionInterface) *PermissionRepo {
	return &PermissionRepo{db: tx}
}

// Create inserts a new permission
func (r *PermissionRepo) Create(perm *model.Permission) error {
	perm.CreatedAt = time.Now()

	query := `
		INSERT INTO permissions (name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(context.Background(), query,
		perm.Name,
		perm.Resource,
		perm.Action,
		perm.Description,
		perm.CreatedAt,
	).Scan(&perm.ID)
}

// GetByID retrieves a permission by ID
func (r *PermissionRepo) GetByID(id int) (*model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE id = $1
	`
	perm := &model.Permission{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// GetByResourceAction retrieves a permission by resource and action
func (r *PermissionRepo) GetByResourceAction(resource, action string) (*model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE resource = $1 AND action = $2
	`
	perm := &model.Permission{}
	err := r.db.QueryRow(context.Background(), query, resource, action).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// GetByName retrieves a permission by name
func (r *PermissionRepo) GetByName(name string) (*model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE name = $1
	`
	perm := &model.Permission{}
	err := r.db.QueryRow(context.Background(), query, name).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// List retrieves all permissions
func (r *PermissionRepo) List() ([]model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		ORDER BY resource, action
	`
	rows, err := r.db.Query(context.Background(), query)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return permissions, nil
}

// ListByResource retrieves all permissions for a specific resource
func (r *PermissionRepo) ListByResource(resource string) ([]model.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE resource = $1
		ORDER BY action
	`
	rows, err := r.db.Query(context.Background(), query, resource)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return permissions, nil
}

// Delete removes a permission by ID
func (r *PermissionRepo) Delete(id int) error {
	// First remove all role assignments
	_, err := r.db.Exec(context.Background(), "DELETE FROM role_permissions WHERE permission_id = $1", id)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(context.Background(), "DELETE FROM permissions WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return nil
}

// CreateIfNotExists creates a permission if it doesn't already exist
func (r *PermissionRepo) CreateIfNotExists(perm *model.Permission) error {
	// Check if permission already exists
	existing, err := r.GetByResourceAction(perm.Resource, perm.Action)
	if err == nil {
		// Permission exists, update the input with existing data
		*perm = *existing
		return nil
	}
	if err != ErrPermissionNotFound {
		return err
	}
	// Permission doesn't exist, create it
	return r.Create(perm)
}

// GetByIDs retrieves multiple permissions by their IDs
func (r *PermissionRepo) GetByIDs(ids []int) ([]model.Permission, error) {
	if len(ids) == 0 {
		return []model.Permission{}, nil
	}

	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE id = ANY($1)
		ORDER BY resource, action
	`
	rows, err := r.db.Query(context.Background(), query, ids)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return permissions, nil
}

// Count returns the total number of permissions
func (r *PermissionRepo) Count() (int, error) {
	query := `SELECT COUNT(*) FROM permissions`
	var count int
	err := r.db.QueryRow(context.Background(), query).Scan(&count)
	return count, err
}

// ListByRoleID retrieves all permissions for a specific role
func (r *PermissionRepo) ListByRoleID(roleID int) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.Query(context.Background(), query, roleID)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return permissions, nil
}
