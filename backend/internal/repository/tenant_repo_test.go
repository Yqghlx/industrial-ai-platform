package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FIX-013: Tenant Repository 测试

func TestTenantRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("success", func(t *testing.T) {
		tenant := &model.Tenant{
			ID:         "tenant-001",
			Name:       "Test Tenant",
			Slug:       "test-tenant",
			Plan:       "basic",
			MaxDevices: 10,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec(`INSERT INTO tenants`).
			WithArgs(tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.MaxDevices, tenant.CreatedAt, tenant.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(tenant)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate_slug", func(t *testing.T) {
		tenant := &model.Tenant{
			ID:         "tenant-002",
			Name:       "Duplicate",
			Slug:       "test-tenant",
			Plan:       "basic",
			MaxDevices: 10,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec(`INSERT INTO tenants`).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		err := repo.Create(tenant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})
}

func TestTenantRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
			AddRow("tenant-001", "Test Tenant", "test-tenant", "basic", 10, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, name, slug, plan, max_devices, created_at, updated_at FROM tenants WHERE id`).
			WithArgs("tenant-001").
			WillReturnRows(rows)

		tenant, err := repo.GetByID("tenant-001")
		assert.NoError(t, err)
		assert.Equal(t, "tenant-001", tenant.ID)
		assert.Equal(t, "Test Tenant", tenant.Name)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM tenants WHERE id`).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		tenant, err := repo.GetByID("non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrTenantNotFound, err)
		assert.Nil(t, tenant)
	})
}

func TestTenantRepo_GetBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
			AddRow("tenant-001", "Test Tenant", "test-tenant", "basic", 10, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug`).
			WithArgs("test-tenant").
			WillReturnRows(rows)

		tenant, err := repo.GetBySlug("test-tenant")
		assert.NoError(t, err)
		assert.Equal(t, "test-tenant", tenant.Slug)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM tenants WHERE slug`).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		tenant, err := repo.GetBySlug("non-existent")
		assert.Error(t, err)
		assert.Nil(t, tenant)
	})
}

func TestTenantRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("success", func(t *testing.T) {
		tenant := &model.Tenant{
			ID:         "tenant-001",
			Name:       "Updated Name",
			Slug:       "updated-slug",
			Plan:       "premium",
			MaxDevices: 100,
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec(`UPDATE tenants SET`).
			WithArgs(tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.MaxDevices, tenant.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(tenant)
		assert.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		tenant := &model.Tenant{
			ID:         "non-existent",
			Name:       "Update",
			Slug:       "update-slug",
			Plan:       "basic",
			MaxDevices: 10,
			UpdatedAt:  time.Now(),
		}

		mock.ExpectExec(`UPDATE tenants SET`).
			WithArgs(tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.MaxDevices, tenant.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(tenant)
		assert.Error(t, err)
		assert.Equal(t, ErrTenantNotFound, err)
	})
}

func TestTenantRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM tenants WHERE id`).
			WithArgs("tenant-001").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete("tenant-001")
		assert.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM tenants WHERE id`).
			WithArgs("non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete("non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrTenantNotFound, err)
	})
}

func TestTenantRepo_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	t.Run("with_pagination", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"}).
			AddRow("tenant-001", "Tenant 1", "tenant-1", "basic", 10, time.Now(), time.Now()).
			AddRow("tenant-002", "Tenant 2", "tenant-2", "premium", 100, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT .* FROM tenants ORDER BY created_at DESC LIMIT`).
			WithArgs(10, 0).
			WillReturnRows(rows)

		tenants, err := repo.List(10, 0)
		assert.NoError(t, err)
		assert.Len(t, tenants, 2)
	})

	t.Run("empty_result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "slug", "plan", "max_devices", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT .* FROM tenants`).
			WithArgs(10, 20).
			WillReturnRows(rows)

		tenants, err := repo.List(10, 20)
		assert.NoError(t, err)
		assert.Len(t, tenants, 0)
	})
}

func TestTenantRepo_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTenantRepo(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(rows)

	count, err := repo.Count()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
