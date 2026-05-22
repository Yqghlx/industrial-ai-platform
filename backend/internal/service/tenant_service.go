package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/constants"
	"github.com/industrial-ai/platform/pkg/errors"
)

type TenantService struct {
	repo repository.TenantRepositoryInterface
}

func NewTenantService(repo repository.TenantRepositoryInterface) *TenantService {
	return &TenantService{repo: repo}
}

// FIX-003: 添加 context 参数
// BE-P2-02: 使用常量替换魔法数字
func (s *TenantService) CreateTenant(ctx context.Context, name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	// Check if slug exists
	existing, err := s.repo.GetBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, errors.NewAppError(errors.ErrCodeConflict, "Tenant slug already exists", slug)
	}

	// Validate plan
	validPlans := map[string]bool{"free": true, "pro": true, "enterprise": true}
	if !validPlans[plan] {
		plan = "free"
	}

	// Set default max devices based on plan
	if maxDevices <= 0 {
		maxDevices = s.getDefaultMaxDevices(plan)
	}

	tenant := &model.Tenant{
		ID:         uuid.New().String(),
		Name:       name,
		Slug:       slug,
		Plan:       plan,
		MaxDevices: maxDevices,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.repo.Create(ctx, tenant)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	return tenant, nil
}

// FIX-003: 添加 context 参数
func (s *TenantService) GetTenant(ctx context.Context, id string) (*model.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// FIX-003: 添加 context 参数
func (s *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.NewTenantNotFoundError(slug)
	}
	return tenant, nil
}

// FIX-003: 添加 context 参数
// BE-P2-02: 使用常量替换魔法数字
func (s *TenantService) ListTenants(ctx context.Context, limit, offset int) ([]model.Tenant, error) {
	if limit <= 0 {
		limit = constants.MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	tenants, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return tenants, nil
}

// FIX-003: 添加 context 参数
func (s *TenantService) UpdateTenant(ctx context.Context, id string, updates map[string]interface{}) (*model.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewTenantNotFoundError(id)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		tenant.Name = name
	}
	if slug, ok := updates["slug"].(string); ok && slug != "" {
		// Check if new slug exists for another tenant
		existing, err := s.repo.GetBySlug(ctx, slug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.NewAppError(errors.ErrCodeConflict, "Tenant slug already exists", slug)
		}
		tenant.Slug = slug
	}
	if plan, ok := updates["plan"].(string); ok {
		validPlans := map[string]bool{"free": true, "pro": true, "enterprise": true}
		if validPlans[plan] {
			tenant.Plan = plan
		}
	}
	if maxDevices, ok := updates["max_devices"].(int); ok && maxDevices > 0 {
		tenant.MaxDevices = maxDevices
	}

	tenant.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, tenant)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	return tenant, nil
}

// FIX-003: 添加 context 参数
func (s *TenantService) DeleteTenant(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

// FIX-003: 添加 context 参数
func (s *TenantService) CountTenants(ctx context.Context) (int, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, errors.NewDatabaseError(err.Error())
	}
	return count, nil
}