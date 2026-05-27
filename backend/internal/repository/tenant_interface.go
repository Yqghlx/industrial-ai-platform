package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
)

// TenantRepositoryInterface defines the interface for tenant repository
// FIX-003: 添加 context 支持
type TenantRepositoryInterface interface {
	Create(ctx context.Context, tenant *model.Tenant) error
	GetByID(ctx context.Context, id string) (*model.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*model.Tenant, error)
	List(ctx context.Context, limit, offset int) ([]model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}
