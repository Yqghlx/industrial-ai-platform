package repository

import (
	"context"

	"database/sql"
	"errors"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
)

var ErrTenantNotFound = errors.New("tenant not found")

type TenantRepo struct {
	db database.QueryExecutor
}

func NewTenantRepo(db database.QueryExecutor) *TenantRepo {
	return &TenantRepo{db: db}
}

// WithTx returns a new repository that uses the given transaction
func (r *TenantRepo) WithTx(tx database.TransactionInterface) *TenantRepo {
	return &TenantRepo{db: tx}
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) Create(ctx context.Context, tenant *model.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, plan, max_devices, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Plan,
		tenant.MaxDevices,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	return err
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	tenant := &model.Tenant{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Plan,
		&tenant.MaxDevices,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`
	tenant := &model.Tenant{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Plan,
		&tenant.MaxDevices,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) List(ctx context.Context, limit, offset int) ([]model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []model.Tenant
	for rows.Next() {
		tenant := model.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Plan,
			&tenant.MaxDevices,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) Update(ctx context.Context, tenant *model.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, slug = $3, plan = $4, max_devices = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Plan,
		tenant.MaxDevices,
		tenant.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenants WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// FIX-003: 添加 context 参数，替换 context.Background()
func (r *TenantRepo) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tenants`
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}