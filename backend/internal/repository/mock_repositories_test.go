package repository

import (
	"context"
	"testing"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// Mock Implementation Tests - Phase 2 Verification
// ============================================

func TestMockDeviceRepository_InterfaceCompliance(t *testing.T) {
	// Verify Mock implements the interface
	var _ DeviceRepositoryInterface = (*MockDeviceRepository)(nil)

	mock := new(MockDeviceRepository)
	ctx := context.Background()

	// Test Create
	device := &model.Device{
		ID:        "test-001",
		Name:      "Test Device",
		Type:      "sensor",
		Status:    "online",
		CreatedAt: time.Now(),
	}
	mock.On("Create", ctx, device).Return(nil)
	err := mock.Create(ctx, device)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockDeviceRepository_GetByID(t *testing.T) {
	mock := new(MockDeviceRepository)
	ctx := context.Background()

	expectedDevice := &model.Device{
		ID:     "test-001",
		Name:   "Test Device",
		Type:   "sensor",
		Status: "online",
	}

	mock.On("GetByID", ctx, "test-001").Return(expectedDevice, nil)

	device, err := mock.GetByID(ctx, "test-001")
	require.NoError(t, err)
	assert.Equal(t, "test-001", device.ID)
	assert.Equal(t, "Test Device", device.Name)
	mock.AssertExpectations(t)
}

func TestMockDeviceRepository_List(t *testing.T) {
	mock := new(MockDeviceRepository)
	ctx := context.Background()

	expectedDevices := []model.Device{
		{ID: "device-1", Name: "Device 1", Type: "sensor"},
		{ID: "device-2", Name: "Device 2", Type: "motor"},
	}

	mock.On("List", ctx, 1, 10).Return(expectedDevices, 2, nil)

	devices, total, err := mock.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, 2, total)
	mock.AssertExpectations(t)
}

func TestMockUserRepository_InterfaceCompliance(t *testing.T) {
	var _ UserRepositoryInterface = (*MockUserRepository)(nil)

	mock := new(MockUserRepository)
	ctx := context.Background()

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
	}

	mock.On("Create", ctx, user).Return(nil)
	err := mock.Create(ctx, user)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockAlertRepository_InterfaceCompliance(t *testing.T) {
	var _ AlertRepositoryInterface = (*MockAlertRepository)(nil)

	mock := new(MockAlertRepository)
	ctx := context.Background()

	alert := &model.Alert{
		ID:          1,
		DeviceID:    "device-001",
		Severity:    "high",
		Status:      "active",
		TriggeredAt: time.Now(),
	}

	mock.On("Create", ctx, alert).Return(nil)
	err := mock.Create(ctx, alert)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockAlertRepository_CountActive(t *testing.T) {
	mock := new(MockAlertRepository)
	ctx := context.Background()

	mock.On("CountActive", ctx).Return(5, nil)

	count, err := mock.CountActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
	mock.AssertExpectations(t)
}

func TestMockRuleRepository_InterfaceCompliance(t *testing.T) {
	var _ RuleRepositoryInterface = (*MockRuleRepository)(nil)

	mock := new(MockRuleRepository)
	ctx := context.Background()

	rule := &model.AlertRule{
		ID:       1,
		Name:     "High Temperature Alert",
		Enabled:  true,
		Severity: "critical",
	}

	mock.On("Create", ctx, rule).Return(nil)
	err := mock.Create(ctx, rule)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockTenantRepository_InterfaceCompliance(t *testing.T) {
	var _ TenantRepositoryInterface = (*MockTenantRepository)(nil)

	mock := new(MockTenantRepository)
	ctx := context.Background()

	tenant := &model.Tenant{
		ID:   "tenant-001",
		Name: "Test Tenant",
		Slug: "test-tenant",
	}

	mock.On("Create", ctx, tenant).Return(nil)
	err := mock.Create(ctx, tenant)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockRBACRepository_InterfaceCompliance(t *testing.T) {
	var _ RBACRepositoryInterface = (*MockRBACRepository)(nil)

	mock := new(MockRBACRepository)
	ctx := context.Background()

	role := &model.Role{
		ID:       1,
		Name:     "admin",
		TenantID: "tenant-001",
	}

	mock.On("CreateRole", ctx, role).Return(nil)
	err := mock.CreateRole(ctx, role)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockRBACRepository_CheckPermission(t *testing.T) {
	mock := new(MockRBACRepository)
	ctx := context.Background()

	mock.On("CheckPermission", ctx, 1, "devices", "read").Return(true, nil)

	hasPermission, err := mock.CheckPermission(ctx, 1, "devices", "read")
	require.NoError(t, err)
	assert.True(t, hasPermission)
	mock.AssertExpectations(t)
}
