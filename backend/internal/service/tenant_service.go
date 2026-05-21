package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/errors"
)

type TenantService struct {
	repo *repository.TenantRepo
}

func NewTenantService(repo *repository.TenantRepo) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) CreateTenant(name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	// Check if slug exists
	existing, err := s.repo.GetBySlug(slug)
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

	err = s.repo.Create(tenant)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	return tenant, nil
}

func (s *TenantService) getDefaultMaxDevices(plan string) int {
	limits := map[string]int{
		"free":       10,
		"pro":        100,
		"enterprise": 1000,
	}
	return limits[plan]
}

func (s *TenantService) GetTenant(id string) (*model.Tenant, error) {
	return s.repo.GetByID(id)
}

func (s *TenantService) GetTenantBySlug(slug string) (*model.Tenant, error) {
	tenant, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, errors.NewTenantNotFoundError(slug)
	}
	return tenant, nil
}

func (s *TenantService) ListTenants(limit, offset int) ([]model.Tenant, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	tenants, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}
	return tenants, nil
}

func (s *TenantService) UpdateTenant(id string, updates map[string]interface{}) (*model.Tenant, error) {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.NewTenantNotFoundError(id)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		tenant.Name = name
	}
	if slug, ok := updates["slug"].(string); ok && slug != "" {
		// Check if new slug exists for another tenant
		existing, err := s.repo.GetBySlug(slug)
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

	err = s.repo.Update(tenant)
	if err != nil {
		return nil, errors.NewDatabaseError(err.Error())
	}

	return tenant, nil
}

func (s *TenantService) DeleteTenant(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return errors.NewDatabaseError(err.Error())
	}
	return nil
}

func (s *TenantService) CountTenants() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, errors.NewDatabaseError(err.Error())
	}
	return count, nil
}
