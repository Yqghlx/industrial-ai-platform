package repository

import (
	"database/sql"
	"errors"

	"github.com/industrial-ai/platform/internal/model"
)

var ErrTenantNotFound = errors.New("tenant not found")

type TenantRepo struct {
	db *sql.DB
}

func NewTenantRepo(db *sql.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) Create(tenant *model.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, plan, max_devices, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query,
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

func (r *TenantRepo) GetByID(id string) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	tenant := &model.Tenant{}
	err := r.db.QueryRow(query, id).Scan(
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

func (r *TenantRepo) GetBySlug(slug string) (*model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`
	tenant := &model.Tenant{}
	err := r.db.QueryRow(query, slug).Scan(
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

func (r *TenantRepo) List(limit, offset int) ([]model.Tenant, error) {
	query := `
		SELECT id, name, slug, plan, max_devices, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, limit, offset)
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

func (r *TenantRepo) Update(tenant *model.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, slug = $3, plan = $4, max_devices = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.db.Exec(query,
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

func (r *TenantRepo) Delete(id string) error {
	query := `DELETE FROM tenants WHERE id = $1`
	result, err := r.db.Exec(query, id)
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

func (r *TenantRepo) Count() (int, error) {
	query := `SELECT COUNT(*) FROM tenants`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
