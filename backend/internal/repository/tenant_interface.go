package repository

import (
	"github.com/industrial-ai/platform/internal/model"
)

// TenantRepositoryInterface defines the interface for tenant repository
type TenantRepositoryInterface interface {
	Create(tenant *model.Tenant) error
	GetByID(id string) (*model.Tenant, error)
	GetBySlug(slug string) (*model.Tenant, error)
	List(limit, offset int) ([]model.Tenant, error)
	Update(tenant *model.Tenant) error
	Delete(id string) error
	Count() (int, error)
}
