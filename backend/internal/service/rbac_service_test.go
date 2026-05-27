package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/industrial-ai/platform/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
)

// Test time value for mock rows
var testTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

// ============================================
// RBACService with RBACRepository (sqlmock)
// ============================================

func newTestRBACServiceWithRBACRepo(t *testing.T) (*RBACService, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	rbacRepo := repository.NewRBACRepository(database.NewDBWrapper(db))
	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	tenantRepo := repository.NewTenantRepo(database.NewDBWrapper(db))
	svc := NewRBACServiceWithRBACRepo(rbacRepo, userRepo, tenantRepo)
	t.Cleanup(func() {
		mock.ExpectationsWereMet()
		db.Close()
	})
	return svc, mock, db
}

func newTestRBACService(t *testing.T) (*RBACService, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	roleRepo := repository.NewRoleRepo(database.NewDBWrapper(db))
	permRepo := repository.NewPermissionRepo(database.NewDBWrapper(db))
	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	tenantRepo := repository.NewTenantRepo(database.NewDBWrapper(db))
	svc := NewRBACService(roleRepo, permRepo, userRepo, tenantRepo)
	t.Cleanup(func() {
		mock.ExpectationsWereMet()
		db.Close()
	})
	return svc, mock, db
}

func TestNewRBACService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewRBACService(
		repository.NewRoleRepo(database.NewDBWrapper(db)),
		repository.NewPermissionRepo(database.NewDBWrapper(db)),
		repository.NewUserRepository(database.NewDBWrapper(db)),
		repository.NewTenantRepo(database.NewDBWrapper(db)),
	)
	assert.NotNil(t, svc)
}

func TestNewRBACServiceWithRBACRepo(t *testing.T) {
	svc, _, _ := newTestRBACServiceWithRBACRepo(t)
	assert.NotNil(t, svc)
}

// ============================================
// CreateRole Tests
// ============================================

func TestRBACService_CreateRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// GetByName - not found
	mock.ExpectQuery(`SELECT .* FROM roles WHERE name = .*`).
		WithArgs("custom-role").
		WillReturnError(sql.ErrNoRows)

	// CreateRole
	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs("custom-role", "A custom role", "", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	role, err := svc.CreateRole(ctx, "", "custom-role", "", "A custom role")
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "custom-role", role.Name)
}

func TestRBACService_CreateRole_AlreadyExists(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "", true, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE name = .*`).
		WithArgs("admin").
		WillReturnRows(rows)

	role, err := svc.CreateRole(ctx, "", "admin", "", "Admin role")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Role already exists")
	assert.Nil(t, role)
}

func TestRBACService_CreateRole_DBError(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE name = .*`).
		WithArgs("new-role").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO roles`).
		WillReturnError(errors.New("db error"))

	role, err := svc.CreateRole(ctx, "", "new-role", "", "desc")
	assert.Error(t, err)
	assert.Nil(t, role)
}

// ============================================
// GetRole Tests
// ============================================

func TestRBACService_GetRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin role", "", true, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	role, err := svc.GetRole(ctx, 1)
	assert.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "admin", role.Name)
}

func TestRBACService_GetRole_NotFound(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(999).
		WillReturnError(repository.ErrRoleNotFound)

	role, err := svc.GetRole(ctx, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Role not found")
	assert.Nil(t, role)
}

// ============================================
// ListRoles Tests
// ============================================

func TestRBACService_ListRoles_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "t1", true, testTime, testTime).
		AddRow(2, "viewer", "Viewer", "t1", false, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles`).
		WithArgs("t1").
		WillReturnRows(rows)

	roles, err := svc.ListRoles(ctx, "t1")
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRBACService_ListRoles_DBError(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles`).
		WithArgs("t1").
		WillReturnError(errors.New("db error"))

	roles, err := svc.ListRoles(ctx, "t1")
	assert.Error(t, err)
	assert.Nil(t, roles)
}

// ============================================
// UpdateRole Tests
// ============================================

func TestRBACService_UpdateRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "old-name", "old desc", "", false, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectExec(`UPDATE roles SET`).
		WithArgs("new-name", "new desc", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateRole(ctx, 1, map[string]interface{}{
		"name":        "new-name",
		"description": "new desc",
	})
	assert.NoError(t, err)
}

func TestRBACService_UpdateRole_NotFound(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(999).
		WillReturnError(repository.ErrRoleNotFound)

	err := svc.UpdateRole(ctx, 999, map[string]interface{}{"name": "x"})
	assert.Contains(t, err.Error(), "Role not found")
}

// ============================================
// DeleteRole Tests
// ============================================

func TestRBACService_DeleteRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(5, "custom", "Custom role", "", false, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(5).
		WillReturnRows(rows)

	mock.ExpectExec(`DELETE FROM roles WHERE id = .*`).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteRole(ctx, 5)
	assert.NoError(t, err)
}

func TestRBACService_DeleteRole_SystemRole(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "", true, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	err := svc.DeleteRole(ctx, 1)
	assert.Contains(t, err.Error(), "Cannot delete system role")
}

func TestRBACService_DeleteRole_NotFound(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(999).
		WillReturnError(repository.ErrRoleNotFound)

	err := svc.DeleteRole(ctx, 999)
	assert.Contains(t, err.Error(), "Role not found")
}

// ============================================
// AssignRole Tests
// ============================================

func TestRBACService_AssignRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// Verify role exists
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "t1", true, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs(10, 1, "t1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.AssignRole(ctx, 10, 1, "t1")
	assert.NoError(t, err)
}

func TestRBACService_AssignRole_RoleNotFound(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(999).
		WillReturnError(repository.ErrRoleNotFound)

	err := svc.AssignRole(ctx, 10, 999, "t1")
	assert.Contains(t, err.Error(), "Role not found")
}

// ============================================
// RemoveRoleFromUser Tests
// ============================================

func TestRBACService_RemoveRoleFromUser_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .*`).
		WithArgs(10, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RemoveRoleFromUser(ctx, 10, 1)
	assert.NoError(t, err)
}

func TestRBACService_RemoveRoleFromUser_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = .*`).
		WillReturnError(errors.New("db error"))

	err := svc.RemoveRoleFromUser(ctx, 10, 1)
	assert.Error(t, err)
}

// ============================================
// GetUserRoles Tests
// ============================================

func TestRBACService_GetUserRoles_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "t1", true, testTime, testTime).
		AddRow(2, "viewer", "Viewer", "t1", false, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles .* JOIN user_roles`).
		WithArgs(10).
		WillReturnRows(rows)

	roles, err := svc.GetUserRoles(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRBACService_GetUserRoles_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles .* JOIN user_roles`).
		WillReturnError(errors.New("db error"))

	roles, err := svc.GetUserRoles(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, roles)
}

// ============================================
// GetUserPermissions Tests
// ============================================

func TestRBACService_GetUserPermissions_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime).
		AddRow(2, "devices.write", "devices", "write", "Write devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	perms, err := svc.GetUserPermissions(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestRBACService_GetUserPermissions_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WillReturnError(errors.New("db error"))

	perms, err := svc.GetUserPermissions(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, perms)
}

// ============================================
// CheckPermission Tests
// ============================================

func TestRBACService_CheckPermission_True(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(10, "devices", "read").
		WillReturnRows(sqlmock.NewRows([]string{"has_permission"}).AddRow(true))

	hasPerm, err := svc.CheckPermission(ctx, 10, "devices", "read")
	assert.NoError(t, err)
	assert.True(t, hasPerm)
}

func TestRBACService_CheckPermission_False(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(10, "devices", "delete").
		WillReturnRows(sqlmock.NewRows([]string{"has_permission"}).AddRow(false))

	hasPerm, err := svc.CheckPermission(ctx, 10, "devices", "delete")
	assert.NoError(t, err)
	assert.False(t, hasPerm)
}

// ============================================
// HasAnyPermission Tests
// ============================================

func TestRBACService_HasAnyPermission_True(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	perms := []struct {
		Resource string
		Action   string
	}{
		{Resource: "devices", Action: "read"},
		{Resource: "devices", Action: "write"},
	}

	hasAny, err := svc.HasAnyPermission(ctx, 10, perms)
	assert.NoError(t, err)
	assert.True(t, hasAny)
}

func TestRBACService_HasAnyPermission_False(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	perms := []struct {
		Resource string
		Action   string
	}{
		{Resource: "users", Action: "delete"},
		{Resource: "users", Action: "write"},
	}

	hasAny, err := svc.HasAnyPermission(ctx, 10, perms)
	assert.NoError(t, err)
	assert.False(t, hasAny)
}

// ============================================
// HasAllPermissions Tests
// ============================================

func TestRBACService_HasAllPermissions_True(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime).
		AddRow(2, "devices.write", "devices", "write", "Write devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	perms := []struct {
		Resource string
		Action   string
	}{
		{Resource: "devices", Action: "read"},
		{Resource: "devices", Action: "write"},
	}

	hasAll, err := svc.HasAllPermissions(ctx, 10, perms)
	assert.NoError(t, err)
	assert.True(t, hasAll)
}

func TestRBACService_HasAllPermissions_False(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	perms := []struct {
		Resource string
		Action   string
	}{
		{Resource: "devices", Action: "read"},
		{Resource: "devices", Action: "write"},
	}

	hasAll, err := svc.HasAllPermissions(ctx, 10, perms)
	assert.NoError(t, err)
	assert.False(t, hasAll)
}

// ============================================
// AssignPermissionToRole Tests
// ============================================

func TestRBACService_AssignPermissionToRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// Verify role exists
	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
			AddRow(1, "admin", "Admin", "", true, testTime, testTime))

	// Verify permission exists
	mock.ExpectQuery(`SELECT .* FROM permissions WHERE id = .*`).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(10, "devices.read", "devices", "read", "Read devices", testTime))

	// Assign
	mock.ExpectExec(`INSERT INTO role_permissions`).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.AssignPermissionToRole(ctx, 1, 10)
	assert.NoError(t, err)
}

func TestRBACService_AssignPermissionToRole_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO role_permissions`).
		WillReturnError(errors.New("db error"))

	err := svc.AssignPermissionToRole(ctx, 1, 10)
	assert.Error(t, err)
}

// ============================================
// RemovePermissionFromRole Tests
// ============================================

func TestRBACService_RemovePermissionFromRole_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RemovePermissionFromRole(ctx, 1, 10)
	assert.NoError(t, err)
}

func TestRBACService_RemovePermissionFromRole_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .*`).
		WillReturnError(errors.New("db error"))

	err := svc.RemovePermissionFromRole(ctx, 1, 10)
	assert.Error(t, err)
}

// ============================================
// CreatePermission Tests
// ============================================

func TestRBACService_CreatePermission_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// CreatePermission with rbacRepo calls rbacRepo.CreatePermission which uses QueryRow
	mock.ExpectQuery(`INSERT INTO permissions`).
		WithArgs("devices.read", "devices", "read", "Read devices", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	perm, err := svc.CreatePermission(ctx, "devices.read", "devices", "read", "Read devices")
	assert.NoError(t, err)
	assert.NotNil(t, perm)
}

func TestRBACService_CreatePermission_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO permissions`).
		WillReturnError(errors.New("db error"))

	perm, err := svc.CreatePermission(ctx, "devices.read", "devices", "read", "Read devices")
	assert.Error(t, err)
	assert.Nil(t, perm)
}

// ============================================
// GetPermission Tests
// ============================================

func TestRBACService_GetPermission_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT .* FROM permissions WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(rows)

	perm, err := svc.GetPermission(ctx, 1)
	assert.NoError(t, err)
	require.NotNil(t, perm)
	assert.Equal(t, "devices.read", perm.Name)
}

func TestRBACService_GetPermission_NotFound(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM permissions WHERE id = .*`).
		WithArgs(999).
		WillReturnError(repository.ErrPermissionNotFound)

	perm, err := svc.GetPermission(ctx, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission not found")
	assert.Nil(t, perm)
}

// ============================================
// ListPermissions Tests
// ============================================

func TestRBACService_ListPermissions_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime).
		AddRow(2, "devices.write", "devices", "write", "Write devices", testTime)

	mock.ExpectQuery(`SELECT .* FROM permissions`).
		WillReturnRows(rows)

	perms, err := svc.ListPermissions(ctx)
	assert.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestRBACService_ListPermissions_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM permissions`).
		WillReturnError(errors.New("db error"))

	perms, err := svc.ListPermissions(ctx)
	assert.Error(t, err)
	assert.Nil(t, perms)
}

// ============================================
// GetRolePermissions Tests
// ============================================

func TestRBACService_GetRolePermissions_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT .* FROM permissions .* JOIN role_permissions`).
		WithArgs(1).
		WillReturnRows(rows)

	perms, err := svc.GetRolePermissions(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestRBACService_GetRolePermissions_Error(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM permissions .* JOIN role_permissions`).
		WillReturnError(errors.New("db error"))

	perms, err := svc.GetRolePermissions(ctx, 1)
	assert.Error(t, err)
	assert.Nil(t, perms)
}

// ============================================
// GetRoleWithPermissions Tests
// ============================================

func TestRBACService_GetRoleWithPermissions_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	roleRows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "", true, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles WHERE id = .*`).
		WithArgs(1).
		WillReturnRows(roleRows)

	permRows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime)

	mock.ExpectQuery(`SELECT .* FROM permissions .* JOIN role_permissions`).
		WithArgs(1).
		WillReturnRows(permRows)

	resp, err := svc.GetRoleWithPermissions(ctx, 1)
	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "admin", resp.Role.Name)
}

// ============================================
// SeedDefaultRolesAndPermissions Tests
// ============================================

func TestRBACService_SeedDefaultRolesAndPermissions_WithRBACRepo(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// With rbacRepo, it calls rbacRepo.InitializeDefaultRBAC(ctx)
	// For each permission: GetPermissionByID -> ErrPermissionNotFound -> CreatePermission
	for range model.DefaultPermissions {
		mock.ExpectQuery(`SELECT .* FROM permissions WHERE id`).
			WillReturnError(repository.ErrPermissionNotFound)
		mock.ExpectQuery(`INSERT INTO permissions`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	// For each role: GetRoleByID -> ErrRoleNotFound -> CreateRole
	for range model.DefaultRoles {
		mock.ExpectQuery(`SELECT .* FROM roles WHERE id`).
			WillReturnError(repository.ErrRoleNotFound)
		mock.ExpectQuery(`INSERT INTO roles`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	// Permission assignments (repo-level: INSERT INTO role_permissions)
	for range []int{1, 3, 5, 7, 8, 9, 11, 13, 14} {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	for range []int{1, 2, 5, 6, 7, 9, 10, 11, 12} {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	for range []int{2, 4, 6, 7, 10, 12} {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	err := svc.SeedDefaultRolesAndPermissions(ctx, "")
	assert.NoError(t, err)
}

// ============================================
// GetUserRoleNames Tests
// ============================================

func TestRBACService_GetUserRoleNames_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// GetUserRoleNames calls GetUserRoles which does SELECT r.id, r.name, ...
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Admin", "t1", true, testTime, testTime).
		AddRow(2, "viewer", "Viewer", "t1", false, testTime, testTime)

	mock.ExpectQuery(`SELECT .* FROM roles .* JOIN user_roles`).
		WithArgs(10).
		WillReturnRows(rows)

	names, err := svc.GetUserRoleNames(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "admin")
	assert.Contains(t, names, "viewer")
}

// ============================================
// GetUserPermissionStrings Tests
// ============================================

func TestRBACService_GetUserPermissionStrings_Success(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// GetUserPermissionStrings calls GetUserPermissions
	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "devices.read", "devices", "read", "Read devices", testTime).
		AddRow(2, "devices.write", "devices", "write", "Write devices", testTime)

	mock.ExpectQuery(`SELECT DISTINCT .* FROM permissions`).
		WithArgs(10).
		WillReturnRows(rows)

	strs, err := svc.GetUserPermissionStrings(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, strs, 2)
	assert.Contains(t, strs, "devices:read")
	assert.Contains(t, strs, "devices:write")
}

// ============================================
// IsAdmin Tests
// ============================================

func TestRBACService_IsAdmin_True(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// IsAdmin calls GetUserRoles -> SQL: SELECT r.id, r.name, ... FROM roles r INNER JOIN user_roles
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(1, "admin", "Administrator", "tenant-1", true, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT r\.id, r\.name.*FROM roles r.*JOIN user_roles`).
		WithArgs(10).
		WillReturnRows(rows)

	isAdmin, err := svc.IsAdmin(ctx, 10)
	assert.NoError(t, err)
	assert.True(t, isAdmin)
}

func TestRBACService_IsAdmin_False(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow(2, "viewer", "Viewer role", "tenant-1", false, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT r\.id, r\.name.*FROM roles r.*JOIN user_roles`).
		WithArgs(10).
		WillReturnRows(rows)

	isAdmin, err := svc.IsAdmin(ctx, 10)
	assert.NoError(t, err)
	assert.False(t, isAdmin)
}

// ============================================
// HasSystemPermission Tests
// ============================================

func TestRBACService_HasSystemPermission_True(t *testing.T) {
	svc, mock, _ := newTestRBACServiceWithRBACRepo(t)
	ctx := context.Background()

	// HasSystemPermission -> GetUserPermissions -> rbacRepo.GetUserPermissions
	// SQL: SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id JOIN user_roles ur ON rp.role_id = ur.role_id WHERE ur.user_id = $1 ORDER BY p.resource, p.action
	rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
		AddRow(1, "system.manage", "system", "manage", "Manage system", testTime)

	mock.ExpectQuery(`SELECT DISTINCT p\.id.*FROM permissions p`).
		WithArgs(10).
		WillReturnRows(rows)

	hasSys, err := svc.HasSystemPermission(ctx, 10)
	assert.NoError(t, err)
	assert.True(t, hasSys)
}

// ============================================
// NilRepos Tests
// ============================================

func TestRBACService_NilRepos(t *testing.T) {
	// Verify RBACService methods don't panic with nil repos
	svc := &RBACService{}
	ctx := context.Background()

	// These methods handle nil repos gracefully (no panic, may return nil)
	_ = svc // service exists with nil repos
	assert.NotPanics(t, func() {
		svc.GetRole(ctx, 1)
	})
}

// ============================================
// Model Tests
// ============================================

func TestModel_DefaultRoles(t *testing.T) {
	roles := model.DefaultRoles
	assert.NotEmpty(t, roles)
}
