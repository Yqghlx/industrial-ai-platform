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

// FIX-014: RBAC Repository 测试

func TestRBACRepo_ListRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator role", "tenant-001", false, time.Now(), time.Now()).
			AddRow(2, "user", "Standard user role", "tenant-001", false, time.Now(), time.Now()).
			AddRow(3, "viewer", "Viewer role", "tenant-001", false, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE tenant_id`).
			WithArgs("tenant-001").
			WillReturnRows(rows)

		roles, err := repo.ListRoles(ctx, "tenant-001")
		assert.NoError(t, err)
		assert.Len(t, roles, 3)
		assert.Equal(t, "admin", roles[0].Name)
	})

	t.Run("empty_tenant", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
			WithArgs("empty-tenant").
			WillReturnRows(rows)

		roles, err := repo.ListRoles(ctx, "empty-tenant")
		assert.NoError(t, err)
		assert.Len(t, roles, 0)
	})
}

func TestRBACRepo_GetRoleByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator role", "tenant-001", false, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE id`).
			WithArgs(1).
			WillReturnRows(rows)

		role, err := repo.GetRoleByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "admin", role.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM roles WHERE id`).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		role, err := repo.GetRoleByID(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, role)
	})
}

func TestRBACRepo_CreateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		role := &model.Role{
			Name:        "operator",
			Description: "Device operator role",
			TenantID:    "tenant-001",
			IsSystem:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// CreateRole uses QueryRow with RETURNING id, not Exec
		mock.ExpectQuery(`INSERT INTO roles`).
			WithArgs(role.Name, role.Description, role.TenantID, role.IsSystem, role.CreatedAt, role.UpdatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := repo.CreateRole(ctx, role)
		assert.NoError(t, err)
		assert.Equal(t, 1, role.ID)
	})
}

func TestRBACRepo_GetRolePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(1, "device:read", "device", "read", "Read device data", time.Now()).
			AddRow(2, "device:write", "device", "write", "Write device data", time.Now()).
			AddRow(3, "device:delete", "device", "delete", "Delete devices", time.Now())

		// GetRolePermissions uses JOIN with role_permissions table
		mock.ExpectQuery(`SELECT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at FROM permissions p JOIN role_permissions`).
			WithArgs(1).
			WillReturnRows(rows)

		permissions, err := repo.GetRolePermissions(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, permissions, 3)
		assert.Equal(t, "device:read", permissions[0].Name)
	})
}

func TestRBACRepo_GetUserRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator role", "tenant-001", false, time.Now(), time.Now()).
			AddRow(2, "user", "Standard user role", "tenant-001", false, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT r\.id, r\.name, r\.description, r\.tenant_id, r\.is_system, r\.created_at, r\.updated_at FROM roles r`).
			WithArgs(1).
			WillReturnRows(rows)

		userRoles, err := repo.GetUserRoles(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, userRoles, 2)
		assert.Equal(t, "admin", userRoles[0].Name)
	})

	t.Run("no_roles", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT .* FROM roles r`).
			WithArgs(999).
			WillReturnRows(rows)

		userRoles, err := repo.GetUserRoles(ctx, 999)
		assert.NoError(t, err)
		assert.Len(t, userRoles, 0)
	})
}

func TestRBACRepo_AssignRoleToUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO user_roles`).
			WithArgs(1, 2, "tenant-001", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AssignRoleToUser(ctx, 1, 2, "tenant-001")
		assert.NoError(t, err)
	})

	t.Run("duplicate_assignment", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO user_roles`).
			WithArgs(1, 2, "tenant-001", sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value"))

		err := repo.AssignRoleToUser(ctx, 1, 2, "tenant-001")
		assert.Error(t, err)
	})
}

func TestRBACRepo_RemoveRoleFromUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM user_roles WHERE user_id`).
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveRoleFromUser(ctx, 1, 2)
		assert.NoError(t, err)
	})

	t.Run("not_assigned", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM user_roles WHERE user_id`).
			WithArgs(1, 999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveRoleFromUser(ctx, 1, 999)
		assert.NoError(t, err) // Removing non-existent role is not an error
	})
}

func TestRBACRepo_CheckPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("has_permission", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

		mock.ExpectQuery(`SELECT COUNT`).
			WithArgs(1, "device", "read").
			WillReturnRows(rows)

		hasPermission, err := repo.CheckPermission(ctx, 1, "device", "read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("no_permission", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT`).
			WithArgs(1, "device", "delete").
			WillReturnRows(rows)

		hasPermission, err := repo.CheckPermission(ctx, 1, "device", "delete")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})
}

func TestRBACRepo_ListPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(1, "device:read", "device", "read", "Read device data", time.Now()).
			AddRow(2, "device:write", "device", "write", "Write device data", time.Now())

		mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions`).
			WillReturnRows(rows)

		permissions, err := repo.ListPermissions(ctx)
		assert.NoError(t, err)
		assert.Len(t, permissions, 2)
	})
}

func TestRBACRepo_AssignPermissionToRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AssignPermissionToRole(ctx, 1, 2)
		assert.NoError(t, err)
	})

	t.Run("duplicate", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(1, 2).
			WillReturnError(errors.New("duplicate key value"))

		err := repo.AssignPermissionToRole(ctx, 1, 2)
		assert.Error(t, err)
	})
}

func TestRBACRepo_DeleteRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM roles WHERE id`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteRole(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM roles WHERE id`).
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteRole(ctx, 999)
		assert.NoError(t, err) // Deleting non-existent role returns no error
	})
}

func TestRBACRepo_UpdateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		role := &model.Role{
			ID:          1,
			Name:        "updated_admin",
			Description: "Updated description",
			UpdatedAt:   time.Now(),
		}

		mock.ExpectExec(`UPDATE roles SET`).
			WithArgs(role.Name, role.Description, role.UpdatedAt, role.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateRole(ctx, role)
		assert.NoError(t, err)
	})
}

func TestRBACRepo_GetUserPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(1, "device:read", "device", "read", "Read device data", time.Now()).
			AddRow(2, "device:write", "device", "write", "Write device data", time.Now())

		mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at`).
			WithArgs(1).
			WillReturnRows(rows)

		permissions, err := repo.GetUserPermissions(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, permissions, 2)
	})
}

func TestRBACRepo_GetRoleByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
			AddRow(1, "admin", "Administrator role", "tenant-001", false, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE name = .*`).
			WithArgs("admin").
			WillReturnRows(rows)

		role, err := repo.GetRoleByName(ctx, "admin")
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "admin", role.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM roles WHERE name = .*`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		role, err := repo.GetRoleByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrRoleNotFound, err)
		assert.Nil(t, role)
	})

	t.Run("database_error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM roles WHERE name = .*`).
			WithArgs("admin").
			WillReturnError(errors.New("connection failed"))

		role, err := repo.GetRoleByName(ctx, "admin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
		assert.Nil(t, role)
	})
}

func TestRBACRepo_CreatePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		perm := &model.Permission{
			Name:        "device:create",
			Resource:    "device",
			Action:      "create",
			Description: "Create new devices",
			CreatedAt:   time.Now(),
		}

		mock.ExpectQuery(`INSERT INTO permissions`).
			WithArgs(perm.Name, perm.Resource, perm.Action, perm.Description, perm.CreatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := repo.CreatePermission(ctx, perm)
		assert.NoError(t, err)
		assert.Equal(t, 1, perm.ID)
	})

	t.Run("error", func(t *testing.T) {
		perm := &model.Permission{
			Name:        "device:create",
			Resource:    "device",
			Action:      "create",
			Description: "Create new devices",
			CreatedAt:   time.Now(),
		}

		mock.ExpectQuery(`INSERT INTO permissions`).
			WillReturnError(errors.New("duplicate key value"))

		err := repo.CreatePermission(ctx, perm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key value")
	})
}

func TestRBACRepo_GetPermissionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(1, "device:read", "device", "read", "Read device data", time.Now())

		mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = .*`).
			WithArgs(1).
			WillReturnRows(rows)

		perm, err := repo.GetPermissionByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, perm)
		assert.Equal(t, "device:read", perm.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM permissions WHERE id = .*`).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		perm, err := repo.GetPermissionByID(ctx, 999)
		assert.Error(t, err)
		assert.Equal(t, ErrPermissionNotFound, err)
		assert.Nil(t, perm)
	})

	t.Run("database_error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM permissions WHERE id = .*`).
			WithArgs(1).
			WillReturnError(errors.New("connection failed"))

		perm, err := repo.GetPermissionByID(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
		assert.Nil(t, perm)
	})
}

func TestRBACRepo_RemovePermissionFromRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemovePermissionFromRole(ctx, 1, 2)
		assert.NoError(t, err)
	})

	t.Run("not_assigned", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
			WithArgs(1, 999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemovePermissionFromRole(ctx, 1, 999)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = .* AND permission_id = .*`).
			WillReturnError(errors.New("database error"))

		err := repo.RemovePermissionFromRole(ctx, 1, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestRBACRepo_AssignRoleToUser_AlreadyAssigned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// ON CONFLICT DO NOTHING returns 0 rows affected
	mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs(1, 2, "tenant-001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.AssignRoleToUser(ctx, 1, 2, "tenant-001")
	assert.Error(t, err)
	assert.Equal(t, ErrRoleAssigned, err)
}

func TestRBACRepo_AssignPermissionToRole_AlreadyAssigned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// ON CONFLICT DO NOTHING returns 0 rows affected
	mock.ExpectExec(`INSERT INTO role_permissions`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.AssignPermissionToRole(ctx, 1, 2)
	assert.Error(t, err)
	assert.Equal(t, ErrPermissionAssigned, err)
}

func TestRBACRepo_ListRoles_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("tenant-001").
		WillReturnError(errors.New("database error"))

	roles, err := repo.ListRoles(ctx, "tenant-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, roles)
}

func TestRBACRepo_ListRoles_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Return malformed data that will cause scan error
	rows := sqlmock.NewRows([]string{"id", "name", "description", "tenant_id", "is_system", "created_at", "updated_at"}).
		AddRow("invalid_id", "admin", "Admin role", "tenant-001", false, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM roles WHERE tenant_id`).
		WithArgs("tenant-001").
		WillReturnRows(rows)

	roles, err := repo.ListRoles(ctx, "tenant-001")
	assert.Error(t, err)
	assert.Nil(t, roles)
}

func TestRBACRepo_GetRolePermissions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at FROM permissions p JOIN role_permissions`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	permissions, err := repo.GetRolePermissions(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, permissions)
}

func TestRBACRepo_GetUserRoles_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM roles r`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	userRoles, err := repo.GetUserRoles(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, userRoles)
}

func TestRBACRepo_ListPermissions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions`).
		WillReturnError(errors.New("database error"))

	permissions, err := repo.ListPermissions(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, permissions)
}

func TestRBACRepo_CheckPermission_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnError(errors.New("database error"))

	hasPermission, err := repo.CheckPermission(ctx, 1, "device", "read")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.False(t, hasPermission)
}

func TestRBACRepo_GetUserPermissions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.name, p\.resource, p\.action, p\.description, p\.created_at`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	permissions, err := repo.GetUserPermissions(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, permissions)
}

func TestRBACRepo_UpdateRole_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	role := &model.Role{
		ID:          999,
		Name:        "updated_role",
		Description: "Updated description",
		UpdatedAt:   time.Now(),
	}

	// System role or non-existent role returns 0 rows affected
	mock.ExpectExec(`UPDATE roles SET`).
		WithArgs(role.Name, role.Description, role.UpdatedAt, role.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateRole(ctx, role)
	assert.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
}

func TestRBACRepo_DeleteRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM roles WHERE id`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	err = repo.DeleteRole(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// InitializeDefaultRBAC Tests

func TestRBACRepo_InitializeDefaultRBAC_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Mock permission checks (not found - so they will be created)
	for _, perm := range model.DefaultPermissions {
		mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = \$1`).
			WithArgs(perm.ID).
			WillReturnError(ErrPermissionNotFound)

		mock.ExpectQuery(`INSERT INTO permissions`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(perm.ID))
	}

	// Mock role checks (not found - so they will be created)
	for _, role := range model.DefaultRoles {
		mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE id = \$1`).
			WithArgs(role.ID).
			WillReturnError(ErrRoleNotFound)

		mock.ExpectQuery(`INSERT INTO roles`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(role.ID))
	}

	// Mock permission assignments for admin role (role_id = 1)
	adminPermissions := []int{1, 3, 5, 7, 8, 9, 11, 13, 14}
	for _, permID := range adminPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(1, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Mock permission assignments for operator role (role_id = 2)
	operatorPermissions := []int{1, 2, 5, 6, 7, 9, 10, 11, 12}
	for _, permID := range operatorPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(2, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// Mock permission assignments for viewer role (role_id = 3)
	viewerPermissions := []int{2, 4, 6, 7, 10, 12}
	for _, permID := range viewerPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(3, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	err = repo.InitializeDefaultRBAC(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRBACRepo_InitializeDefaultRBAC_PermissionAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Mock first permission check (already exists)
	perm := model.DefaultPermissions[0]
	mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = \$1`).
		WithArgs(perm.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at"}).
			AddRow(perm.ID, perm.Name, perm.Resource, perm.Action, perm.Description, time.Now()))

	// Mock remaining permission checks (not found)
	for i := 1; i < len(model.DefaultPermissions); i++ {
		perm := model.DefaultPermissions[i]
		mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = \$1`).
			WithArgs(perm.ID).
			WillReturnError(ErrPermissionNotFound)

		mock.ExpectQuery(`INSERT INTO permissions`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(perm.ID))
	}

	// Mock role checks
	for _, role := range model.DefaultRoles {
		mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE id = \$1`).
			WithArgs(role.ID).
			WillReturnError(ErrRoleNotFound)

		mock.ExpectQuery(`INSERT INTO roles`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(role.ID))
	}

	// Mock permission assignments
	adminPermissions := []int{1, 3, 5, 7, 8, 9, 11, 13, 14}
	for _, permID := range adminPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(1, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	operatorPermissions := []int{1, 2, 5, 6, 7, 9, 10, 11, 12}
	for _, permID := range operatorPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(2, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	viewerPermissions := []int{2, 4, 6, 7, 10, 12}
	for _, permID := range viewerPermissions {
		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(3, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	err = repo.InitializeDefaultRBAC(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRBACRepo_InitializeDefaultRBAC_CreatePermissionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Mock permission check (not found)
	perm := model.DefaultPermissions[0]
	mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = \$1`).
		WithArgs(perm.ID).
		WillReturnError(ErrPermissionNotFound)

	// Mock create permission error
	mock.ExpectQuery(`INSERT INTO permissions`).
		WillReturnError(errors.New("database error"))

	err = repo.InitializeDefaultRBAC(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRBACRepo_InitializeDefaultRBAC_CreateRoleError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRBACRepository(database.NewDBWrapper(db))
	ctx := context.Background()

	// Mock permission checks and creates
	for _, perm := range model.DefaultPermissions {
		mock.ExpectQuery(`SELECT id, name, resource, action, description, created_at FROM permissions WHERE id = \$1`).
			WithArgs(perm.ID).
			WillReturnError(ErrPermissionNotFound)

		mock.ExpectQuery(`INSERT INTO permissions`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(perm.ID))
	}

	// Mock first role check (not found)
	role := model.DefaultRoles[0]
	mock.ExpectQuery(`SELECT id, name, description, tenant_id, is_system, created_at, updated_at FROM roles WHERE id = \$1`).
		WithArgs(role.ID).
		WillReturnError(ErrRoleNotFound)

	// Mock create role error
	mock.ExpectQuery(`INSERT INTO roles`).
		WillReturnError(errors.New("database error"))

	err = repo.InitializeDefaultRBAC(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
