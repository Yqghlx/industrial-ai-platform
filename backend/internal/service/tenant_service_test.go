package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/database"
)

func newTestTenantService(t *testing.T) (*TenantService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := repository.NewTenantRepo(database.NewDBWrapper(db))
	svc := NewTenantService(repo)
	return svc, mock
}

func TestNewTenantService(t *testing.T) {
	svc, _ := newTestTenantService(t)
	assert.NotNil(t, svc)
}

func TestTenantService_CreateTenant_Success(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	// Slug check - not found
	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("test-tenant").
		WillReturnError(sql.ErrNoRows)

	// Insert
	mock.ExpectExec(`INSERT INTO tenants`).
		WithArgs(sqlmock.AnyArg(), "Test Tenant", "test-tenant", "free", 10, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tenant, err := svc.CreateTenant(ctx, "Test Tenant", "test-tenant", "free", 0)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "Test Tenant", tenant.Name)
	assert.Equal(t, "test-tenant", tenant.Slug)
	assert.Equal(t, "free", tenant.Plan)
	assert.Equal(t, 10, tenant.MaxDevices)
}

func TestTenantService_CreateTenant_SlugExists(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	// Slug check - found
	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("existing-id", "Existing", "test-tenant", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("test-tenant").
		WillReturnRows(rows)

	tenant, err := svc.CreateTenant(ctx, "Test Tenant", "test-tenant", "free", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tenant slug already exists")
	assert.Nil(t, tenant)
}

func TestTenantService_CreateTenant_InvalidPlan(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	// Slug check - not found
	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("test-tenant").
		WillReturnError(sql.ErrNoRows)

	// Insert - plan should default to "free"
	mock.ExpectExec(`INSERT INTO tenants`).
		WithArgs(sqlmock.AnyArg(), "Test", "test-tenant", "free", 10, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tenant, err := svc.CreateTenant(ctx, "Test", "test-tenant", "invalid-plan", 0)
	assert.NoError(t, err)
	assert.Equal(t, "free", tenant.Plan)
}

func TestTenantService_CreateTenant_ProPlan(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("pro-tenant").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO tenants`).
		WithArgs(sqlmock.AnyArg(), "Pro Tenant", "pro-tenant", "pro", 100, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tenant, err := svc.CreateTenant(ctx, "Pro Tenant", "pro-tenant", "pro", 0)
	assert.NoError(t, err)
	assert.Equal(t, "pro", tenant.Plan)
	assert.Equal(t, 100, tenant.MaxDevices)
}

func TestTenantService_CreateTenant_CustomMaxDevices(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("custom-tenant").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO tenants`).
		WithArgs(sqlmock.AnyArg(), "Custom", "custom-tenant", "free", 50, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tenant, err := svc.CreateTenant(ctx, "Custom", "custom-tenant", "free", 50)
	assert.NoError(t, err)
	assert.Equal(t, 50, tenant.MaxDevices)
}

func TestTenantService_CreateTenant_DBError(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("test").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO tenants`).
		WillReturnError(errors.New("db error"))

	tenant, err := svc.CreateTenant(ctx, "Test", "test", "free", 0)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestTenantService_GetTenant(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "test", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	tenant, err := svc.GetTenant(ctx, "t-1")
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "t-1", tenant.ID)
}

func TestTenantService_GetTenant_NotFound(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("nonexistent").
		WillReturnError(repository.ErrTenantNotFound)

	tenant, err := svc.GetTenant(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestTenantService_GetTenantBySlug(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "test-slug", "pro", 100, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("test-slug").
		WillReturnRows(rows)

	tenant, err := svc.GetTenantBySlug(ctx, "test-slug")
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "test-slug", tenant.Slug)
}

func TestTenantService_ListTenants(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Tenant 1", "slug-1", "free", 10, time.Now(), time.Now()).
		AddRow("t-2", "Tenant 2", "slug-2", "pro", 100, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants ORDER BY created_at DESC LIMIT .* OFFSET .*`).
		WithArgs(100, 0).
		WillReturnRows(rows)

	tenants, err := svc.ListTenants(ctx, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, tenants, 2)
}

func TestTenantService_ListTenants_DefaultLimit(t *testing.T) {
	svc, mock := newTestTenantService(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT .* FROM tenants ORDER BY created_at DESC LIMIT .* OFFSET .*`).
		WithArgs(100, 0).
		WillReturnRows(rows)

	tenants, err := svc.ListTenants(ctx, 0, 0)
	assert.NoError(t, err)
	assert.Empty(t, tenants)
}

func TestTenantService_UpdateTenant(t *testing.T) {
	svc, mock := newTestTenantService(t)

	// GetByID
	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Old Name", "old-slug", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	// Update
	mock.ExpectExec(`UPDATE tenants SET`).
		WithArgs("t-1", "New Name", "old-slug", "free", 10, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updates := map[string]interface{}{
		"name": "New Name",
	}
	tenant, err := svc.UpdateTenant(context.Background(), "t-1", updates)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "New Name", tenant.Name)
}

func TestTenantService_UpdateTenant_Slug(t *testing.T) {
	svc, mock := newTestTenantService(t)

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "old-slug", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	// Slug check - not found (no conflict)
	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("new-slug").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`UPDATE tenants SET`).
		WithArgs("t-1", "Test", "new-slug", "free", 10, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updates := map[string]interface{}{
		"slug": "new-slug",
	}
	tenant, err := svc.UpdateTenant(context.Background(), "t-1", updates)
	assert.NoError(t, err)
	assert.Equal(t, "new-slug", tenant.Slug)
}

func TestTenantService_UpdateTenant_SlugConflict(t *testing.T) {
	svc, mock := newTestTenantService(t)

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "old-slug", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	// Slug check - found for another tenant
	existingRows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-2", "Other", "new-slug", "pro", 100, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug = .*`).
		WithArgs("new-slug").
		WillReturnRows(existingRows)

	updates := map[string]interface{}{
		"slug": "new-slug",
	}
	tenant, err := svc.UpdateTenant(context.Background(), "t-1", updates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tenant slug already exists")
	assert.Nil(t, tenant)
}

func TestTenantService_UpdateTenant_Plan(t *testing.T) {
	svc, mock := newTestTenantService(t)

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "test", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	mock.ExpectExec(`UPDATE tenants SET`).
		WithArgs("t-1", "Test", "test", "pro", 10, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updates := map[string]interface{}{
		"plan": "pro",
	}
	tenant, err := svc.UpdateTenant(context.Background(), "t-1", updates)
	assert.NoError(t, err)
	assert.Equal(t, "pro", tenant.Plan)
}

func TestTenantService_UpdateTenant_MaxDevices(t *testing.T) {
	svc, mock := newTestTenantService(t)

	rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
		AddRow("t-1", "Test", "test", "free", 10, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnRows(rows)

	mock.ExpectExec(`UPDATE tenants SET`).
		WithArgs("t-1", "Test", "test", "free", 200, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updates := map[string]interface{}{
		"max_devices": 200,
	}
	tenant, err := svc.UpdateTenant(context.Background(), "t-1", updates)
	assert.NoError(t, err)
	assert.Equal(t, 200, tenant.MaxDevices)
}

func TestTenantService_UpdateTenant_NotFound(t *testing.T) {
	svc, mock := newTestTenantService(t)

	mock.ExpectQuery(`SELECT .* FROM tenants WHERE id = .*`).
		WithArgs("nonexistent").
		WillReturnError(repository.ErrTenantNotFound)

	updates := map[string]interface{}{"name": "New"}
	tenant, err := svc.UpdateTenant(context.Background(), "nonexistent", updates)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestTenantService_DeleteTenant(t *testing.T) {
	svc, mock := newTestTenantService(t)

	mock.ExpectExec(`DELETE FROM tenants WHERE id = .*`).
		WithArgs("t-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteTenant(context.Background(), "t-1")
	assert.NoError(t, err)
}

func TestTenantService_DeleteTenant_NotFound(t *testing.T) {
	svc, mock := newTestTenantService(t)

	mock.ExpectExec(`DELETE FROM tenants WHERE id = .*`).
		WithArgs("nonexistent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.DeleteTenant(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestTenantService_CountTenants(t *testing.T) {
	svc, mock := newTestTenantService(t)

	mock.ExpectQuery(`SELECT COUNT\(\\*\) FROM tenants`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := svc.CountTenants(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 42, count)
}

func TestTenantService_getDefaultMaxDevices(t *testing.T) {
	svc := &TenantService{}

	tests := []struct {
		plan     string
		expected int
	}{
		{"free", 10},
		{"pro", 100},
		{"enterprise", 10000},
		{"unknown", 10},
	}

	for _, tt := range tests {
		t.Run(tt.plan, func(t *testing.T) {
			result := svc.getDefaultMaxDevices(tt.plan)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Compile-time interface check
func TestTenantModel(t *testing.T) {
	tenant := &model.Tenant{
		ID:         "test",
		Name:       "Test",
		Slug:       "test",
		Plan:       "free",
		MaxDevices: 10,
	}
	assert.Equal(t, "test", tenant.ID)
}
