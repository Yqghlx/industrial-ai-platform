package repository

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/mock"
)

// ============================================
// Mock Implementations - Phase 2
// ============================================

// MockDeviceRepository implements DeviceRepositoryInterface for testing
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepository) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	args := m.Called(ctx, page, pageSize)
	devices := args.Get(0).([]model.Device)
	total := args.Get(1).(int)
	return devices, total, args.Error(2)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDeviceRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

// MockUserRepository implements UserRepositoryInterface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	users := args.Get(0).([]model.User)
	total := args.Get(1).(int)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

// MockAlertRepository implements AlertRepositoryInterface for testing
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
}

func (m *MockAlertRepository) CountActive(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockAlertRepository) Resolve(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAlertRepository) GetRecentByDevice(ctx context.Context, deviceID string, ruleID int, cooldownSec int) (*model.Alert, error) {
	args := m.Called(ctx, deviceID, ruleID, cooldownSec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

// MockRuleRepository implements RuleRepositoryInterface for testing
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) List(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) ListEnabled(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) ToggleEnabled(ctx context.Context, id int, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id int) (*model.AlertRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTenantRepository implements TenantRepositoryInterface for testing
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *model.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(id string) (*model.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetBySlug(slug string) (*model.Tenant, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantRepository) List(limit, offset int) ([]model.Tenant, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]model.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(tenant *model.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTenantRepository) Count() (int, error) {
	args := m.Called()
	return args.Get(0).(int), args.Error(1)
}

// MockRBACRepository implements RBACRepositoryInterface for testing
type MockRBACRepository struct {
	mock.Mock
}

func (m *MockRBACRepository) CreateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRBACRepository) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACRepository) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRBACRepository) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRBACRepository) AssignRoleToUser(ctx context.Context, userID, roleID int, tenantID string) error {
	args := m.Called(ctx, userID, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACRepository) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACRepository) CreatePermission(ctx context.Context, perm *model.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockRBACRepository) GetPermissionByID(ctx context.Context, id int) (*model.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockRBACRepository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACRepository) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRBACRepository) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) InitializeDefaultRBAC(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
