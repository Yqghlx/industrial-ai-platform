package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RoleRepo Tests

func TestRoleRepo_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{
		Name:        "operator",
		Description: "Operator role",
		TenantID:    "tenant-001",
		IsSystem:    false,
	}

	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs(role.Name, role.Description, role.TenantID, role.IsSystem, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(context.Background(), role)
	assert.NoError(t, err)
	assert.Equal(t, 1, role.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepo_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{
		Name:        "operator",
		Description: "Operator role",
		TenantID:    "tenant-001",
		IsSystem:    false,
	}

	mock.ExpectQuery(`INSERT INTO roles`).
		WillReturnError(errors.New("duplicate key value"))

	err = repo.Create(context.Background(), role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")
}

func TestRoleRepo_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Administrator role", "tenant-001", false, now, now)

	mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	role, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "admin", role.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepo_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	role, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
}

func TestRoleRepo_GetByName_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Administrator role", "tenant-001", false, now, now)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id = .* AND name = .*`).
		WithArgs("tenant-001", "admin").
		WillReturnRows(rows)

	role, err := repo.GetByName(context.Background(), "tenant-001", "admin")
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "admin", role.Name)
}

func TestRoleRepo_GetByName_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id = .* AND name = .*`).
		WithArgs("tenant-001", "nonexistent").
		WillReturnError(sql.ErrNoRows)

	role, err := repo.GetByName(context.Background(), "tenant-001", "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
}

func TestRoleRepo_ListByTenant_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, now, now).
		AddRow(2, "operator", "Operator role", "tenant-001", false, now, now)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id = .* OR tenant_id = '' ORDER BY created_at DESC`).
		WithArgs("tenant-001").
		WillReturnRows(rows)

	roles, err := repo.ListByTenant(context.Background(), "tenant-001")
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Name)
}

func TestRoleRepo_ListByTenant_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("empty-tenant").
		WillReturnRows(rows)

	roles, err := repo.ListByTenant(context.Background(), "empty-tenant")
	assert.NoError(t, err)
	assert.Len(t, roles, 0)
}

func TestRoleRepo_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{
		ID:          1,
		Name:        "updated_admin",
		Description: "Updated description",
	}

	mock.ExpectExec(`UPDATE roles SET name = .* description = .* updated_at = .* WHERE id = .* AND is_system = false`).
		WithArgs(role.ID, role.Name, role.Description, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), role)
	assert.NoError(t, err)
}

func TestRoleRepo_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{
		ID:          999,
		Name:        "updated",
		Description: "Updated",
	}

	mock.ExpectExec(`UPDATE roles SET`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), role)
	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
}

func TestRoleRepo_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// First, GetByID call
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "custom_role", "Custom role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	// Delete role_permissions
	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete user_roles
	mock.ExpectExec(`DELETE FROM user_roles WHERE role_id = .*`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete the role
	mock.ExpectExec(`DELETE FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestRoleRepo_Delete_SystemRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// First, GetByID returns a system role
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "", true, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	err = repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete system role")
}

func TestRoleRepo_AssignRoleToUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Check if already assigned - returns false
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WithArgs(1, 2).
		WillReturnRows(rows)

	// Insert new assignment
	mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs(1, 2, "tenant-001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AssignRoleToUser(context.Background(), 1, 2, "tenant-001")
	assert.NoError(t, err)
}

func TestRoleRepo_AssignRoleToUser_AlreadyAssigned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Check if already assigned - returns true
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WithArgs(1, 2).
		WillReturnRows(rows)

	err = repo.AssignRoleToUser(context.Background(), 1, 2, "tenant-001")
	assert.NoError(t, err) // Already assigned, no error
}

func TestRoleRepo_RemoveRoleFromUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RemoveRoleFromUser(context.Background(), 1, 2)
	assert.NoError(t, err)
}

func TestRoleRepo_RemoveRoleFromUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WithArgs(1, 999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.RemoveRoleFromUser(context.Background(), 1, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role assignment not found")
}

func TestRoleRepo_GetUserRoles_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, now, now).
		AddRow(2, "operator", "Operator role", "tenant-001", false, now, now)

	mock.ExpectQuery(`SELECT r\.id, r\.name, r\.description, r\.tenant_id, r\.is_system, r\.created_at, r\.updated_at FROM roles r INNER JOIN user_roles`).
		WithArgs(1).
		WillReturnRows(rows)

	roles, err := repo.GetUserRoles(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRoleRepo_GetRolePermissions_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "device:read", "device", "read", "Read devices", time.Now()).
		AddRow(2, "device:write", "device", "write", "Write devices", time.Now())

	mock.ExpectQuery(`SELECT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at FROM permissions p INNER JOIN role_permissions`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetRolePermissions(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestRoleRepo_AssignPermissionToRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`INSERT INTO role_permissions`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AssignPermissionToRole(context.Background(), 1, 2)
	assert.NoError(t, err)
}

func TestRoleRepo_RemovePermissionFromRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RemovePermissionFromRole(context.Background(), 1, 2)
	assert.NoError(t, err)
}

func TestRoleRepo_RemovePermissionFromRole_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
		WithArgs(1, 999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.RemovePermissionFromRole(context.Background(), 1, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission assignment not found")
}

func TestRoleRepo_GetUserPermissions_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "device:read", "device", "read", "Read devices", time.Now()).
		AddRow(2, "device:write", "device", "write", "Write devices", time.Now())

	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at FROM permissions p INNER JOIN`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetUserPermissions(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestRoleRepo_CheckUserPermission_HasPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\( SELECT 1 FROM permissions p INNER JOIN`).
		WithArgs(1, "device", "read").
		WillReturnRows(rows)

	hasPermission, err := repo.CheckUserPermission(context.Background(), 1, "device", "read")
	assert.NoError(t, err)
	assert.True(t, hasPermission)
}

func TestRoleRepo_CheckUserPermission_NoPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\( SELECT 1 FROM permissions`).
		WithArgs(1, "device", "delete").
		WillReturnRows(rows)

	hasPermission, err := repo.CheckUserPermission(context.Background(), 1, "device", "delete")
	assert.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestRoleRepo_GetByIDWithPermissions_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	// GetByID
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	// GetRolePermissions
	permRows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "device:read", "device", "read", "Read", time.Now())
	mock.ExpectQuery(`SELECT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at FROM permissions`).
		WithArgs(1).
		WillReturnRows(permRows)

	result, err := repo.GetByIDWithPermissions(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.Role.Name)
	assert.Len(t, result.Permissions, 1)
}

// Additional tests to improve coverage

func TestRoleRepo_WithTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Begin transaction
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	// WithTx should return a new repo with the transaction
	txRepo := repo.WithTx(&database.TxWrapper{Tx: tx})
	assert.NotNil(t, txRepo)
	assert.NotSame(t, repo, txRepo)
}

func TestRoleRepo_GetByID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Return a generic error (not sql.ErrNoRows)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnError(errors.New("database connection error"))

	role, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection error")
	assert.Nil(t, role)
}

func TestRoleRepo_GetByName_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Return a generic error (not sql.ErrNoRows)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id = .* AND name = .*`).
		WithArgs("tenant-001", "admin").
		WillReturnError(errors.New("database connection error"))

	role, err := repo.GetByName(context.Background(), "tenant-001", "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection error")
	assert.Nil(t, role)
}

func TestRoleRepo_ListByTenant_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("tenant-001").
		WillReturnError(errors.New("database error"))

	roles, err := repo.ListByTenant(context.Background(), "tenant-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, roles)
}

func TestRoleRepo_ListByTenant_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Return malformed data
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow("invalid", "admin", "Admin role", "tenant-001", false, time.Now(), time.Now())
	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("tenant-001").
		WillReturnRows(rows)

	roles, err := repo.ListByTenant(context.Background(), "tenant-001")
	assert.Error(t, err)
	assert.Nil(t, roles)
}

func TestRoleRepo_ListByTenant_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	// Create rows that will trigger an error during iteration
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, time.Now(), time.Now())
	rows.RowError(0, errors.New("rows iteration error"))
	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("tenant-001").
		WillReturnRows(rows)

	roles, err := repo.ListByTenant(context.Background(), "tenant-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows iteration error")
	assert.Nil(t, roles)
}

func TestRoleRepo_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{ID: 1, Name: "updated", Description: "Updated"}

	mock.ExpectExec(`UPDATE roles SET`).
		WillReturnError(errors.New("database error"))

	err = repo.Update(context.Background(), role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestRoleRepo_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	role := &model.Role{ID: 1, Name: "updated", Description: "Updated"}

	// Return a result that will error when RowsAffected() is called
	mock.ExpectExec(`UPDATE roles SET`).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("result error")))

	err = repo.Update(context.Background(), role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result error")
}

func TestRoleRepo_Delete_GetByIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestRoleRepo_Delete_DeleteRolePermissionsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "custom_role", "Custom role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WillReturnError(errors.New("delete error"))

	err = repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestRoleRepo_Delete_DeleteUserRolesError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "custom_role", "Custom role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM user_roles WHERE role_id = .*`).
		WillReturnError(errors.New("delete user_roles error"))

	err = repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete user_roles error")
}

func TestRoleRepo_Delete_DeleteRoleError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "custom_role", "Custom role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM user_roles WHERE role_id = .*`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM roles WHERE id = .*`).
		WillReturnError(errors.New("delete role error"))

	err = repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete role error")
}

func TestRoleRepo_AssignRoleToUser_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_roles`).
		WillReturnError(errors.New("query error"))

	err = repo.AssignRoleToUser(context.Background(), 1, 2, "tenant-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
}

func TestRoleRepo_AssignRoleToUser_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_roles`).
		WithArgs(1, 2).
		WillReturnRows(rows)

	mock.ExpectExec(`INSERT INTO user_roles`).
		WillReturnError(errors.New("insert error"))

	err = repo.AssignRoleToUser(context.Background(), 1, 2, "tenant-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
}

func TestRoleRepo_RemoveRoleFromUser_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WillReturnError(errors.New("delete error"))

	err = repo.RemoveRoleFromUser(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestRoleRepo_RemoveRoleFromUser_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .* AND role_id = .*`).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("result error")))

	err = repo.RemoveRoleFromUser(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestRoleRepo_GetUserRoles_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT r\.id, r\.name.*FROM roles r INNER JOIN user_roles`).
		WillReturnError(errors.New("query error"))

	roles, err := repo.GetUserRoles(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
	assert.Nil(t, roles)
}

func TestRoleRepo_GetUserRoles_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow("invalid", "admin", "Admin role", "tenant-001", false, time.Now(), time.Now())
	mock.ExpectQuery(`SELECT r\.id, r\.name.*FROM roles r INNER JOIN user_roles`).
		WithArgs(1).
		WillReturnRows(rows)

	roles, err := repo.GetUserRoles(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, roles)
}

func TestRoleRepo_GetUserRoles_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, time.Now(), time.Now())
	rows.RowError(0, errors.New("rows error"))
	mock.ExpectQuery(`SELECT r\.id, r\.name.*FROM roles r INNER JOIN user_roles`).
		WithArgs(1).
		WillReturnRows(rows)

	roles, err := repo.GetUserRoles(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error")
	assert.Nil(t, roles)
}

func TestRoleRepo_GetRolePermissions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT p\.id, p\.name.*FROM permissions p INNER JOIN role_permissions`).
		WillReturnError(errors.New("query error"))

	perms, err := repo.GetRolePermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
	assert.Nil(t, perms)
}

func TestRoleRepo_GetRolePermissions_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow("invalid", "device:read", "device", "read", "Read", time.Now())
	mock.ExpectQuery(`SELECT p\.id, p\.name.*FROM permissions p INNER JOIN role_permissions`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetRolePermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, perms)
}

func TestRoleRepo_GetRolePermissions_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "device:read", "device", "read", "Read", time.Now())
	rows.RowError(0, errors.New("rows error"))
	mock.ExpectQuery(`SELECT p\.id, p\.name.*FROM permissions p INNER JOIN role_permissions`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetRolePermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error")
	assert.Nil(t, perms)
}

func TestRoleRepo_AssignPermissionToRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`INSERT INTO role_permissions`).
		WillReturnError(errors.New("insert error"))

	err = repo.AssignPermissionToRole(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
}

func TestRoleRepo_RemovePermissionFromRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
		WillReturnError(errors.New("delete error"))

	err = repo.RemovePermissionFromRole(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestRoleRepo_RemovePermissionFromRole_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("result error")))

	err = repo.RemovePermissionFromRole(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestRoleRepo_GetUserPermissions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT DISTINCT p\.id.*FROM permissions p INNER JOIN`).
		WillReturnError(errors.New("query error"))

	perms, err := repo.GetUserPermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
	assert.Nil(t, perms)
}

func TestRoleRepo_GetUserPermissions_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow("invalid", "device:read", "device", "read", "Read", time.Now())
	mock.ExpectQuery(`SELECT DISTINCT p\.id.*FROM permissions p INNER JOIN`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetUserPermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, perms)
}

func TestRoleRepo_GetUserPermissions_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "device:read", "device", "read", "Read", time.Now())
	rows.RowError(0, errors.New("rows error"))
	mock.ExpectQuery(`SELECT DISTINCT p\.id.*FROM permissions p INNER JOIN`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := repo.GetUserPermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rows error")
	assert.Nil(t, perms)
}

func TestRoleRepo_CheckUserPermission_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT EXISTS\( SELECT 1 FROM permissions`).
		WillReturnError(errors.New("query error"))

	hasPermission, err := repo.CheckUserPermission(context.Background(), 1, "device", "read")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
	assert.False(t, hasPermission)
}

func TestRoleRepo_GetByIDWithPermissions_GetByIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByIDWithPermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, result)
}

func TestRoleRepo_GetByIDWithPermissions_GetRolePermissionsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepo(database.NewDBWrapper(db))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "tenant-001", false, now, now)
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectQuery(`SELECT p\.id.*FROM permissions p INNER JOIN role_permissions`).
		WillReturnError(errors.New("permissions error"))

	result, err := repo.GetByIDWithPermissions(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permissions error")
	assert.Nil(t, result)
}
