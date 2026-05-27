package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRBACServiceInternal is a mock implementation of rbacServiceInternal interface
type MockRBACServiceInternal struct {
	mock.Mock
}

func (m *MockRBACServiceInternal) CreateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACServiceInternal) UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACServiceInternal) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRBACServiceInternal) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACServiceInternal) ListRoles(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACServiceInternal) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACServiceInternal) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACServiceInternal) ListUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACServiceInternal) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACServiceInternal) AssignPermissionToRole(ctx context.Context, roleID, permID int) error {
	args := m.Called(ctx, roleID, permID)
	return args.Error(0)
}

func (m *MockRBACServiceInternal) RemovePermissionFromRole(ctx context.Context, roleID, permID int) error {
	args := m.Called(ctx, roleID, permID)
	return args.Error(0)
}

// ============== CreateRole Tests ==============

func TestAdapter_CreateRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRole := &model.Role{
		ID:          1,
		TenantID:    "tenant-001",
		Name:        "Operator",
		Description: "Device operator role",
	}

	mockSvc.On("CreateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(expectedRole, nil)

	role, err := adapter.CreateRole(ctx, "tenant-001", "Operator", "Operator", "Device operator role")

	require.NoError(t, err)
	assert.Equal(t, expectedRole, role)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_CreateRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role already exists")

	mockSvc.On("CreateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil, testError)

	role, err := adapter.CreateRole(ctx, "tenant-001", "Operator", "Operator", "Device operator role")

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, role)
	mockSvc.AssertExpectations(t)
}

// ============== GetRole Tests ==============

func TestAdapter_GetRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Administrator",
	}

	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(expectedRole, nil)

	role, err := adapter.GetRole(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedRole, role)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_GetRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role not found")

	mockSvc.On("GetRoleByID", mock.Anything, 999).Return(nil, testError)

	role, err := adapter.GetRole(ctx, 999)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, role)
	mockSvc.AssertExpectations(t)
}

// ============== GetRoleWithPermissions Tests ==============

func TestAdapter_GetRoleWithPermissions_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Administrator",
	}

	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(expectedRole, nil)

	roleResponse, err := adapter.GetRoleWithPermissions(ctx, 1)

	require.NoError(t, err)
	assert.NotNil(t, roleResponse)
	assert.Equal(t, *expectedRole, roleResponse.Role)
	assert.Equal(t, []model.Permission{}, roleResponse.Permissions)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_GetRoleWithPermissions_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role not found")

	mockSvc.On("GetRoleByID", mock.Anything, 999).Return(nil, testError)

	roleResponse, err := adapter.GetRoleWithPermissions(ctx, 999)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, roleResponse)
	mockSvc.AssertExpectations(t)
}

// ============== ListRoles Tests ==============

func TestAdapter_ListRoles_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRoles := []model.Role{
		{ID: 1, Name: "admin", TenantID: "tenant-001"},
		{ID: 2, Name: "operator", TenantID: "tenant-001"},
		{ID: 3, Name: "viewer", TenantID: "tenant-002"},
	}

	mockSvc.On("ListRoles", mock.Anything).Return(expectedRoles, nil)

	roles, err := adapter.ListRoles(ctx, "")

	require.NoError(t, err)
	assert.Len(t, roles, 3)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_ListRoles_WithTenantFilter(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRoles := []model.Role{
		{ID: 1, Name: "admin", TenantID: "tenant-001"},
		{ID: 2, Name: "operator", TenantID: "tenant-001"},
		{ID: 3, Name: "viewer", TenantID: "tenant-002"},
	}

	mockSvc.On("ListRoles", mock.Anything).Return(expectedRoles, nil)

	roles, err := adapter.ListRoles(ctx, "tenant-001")

	require.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "tenant-001", roles[0].TenantID)
	assert.Equal(t, "tenant-001", roles[1].TenantID)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_ListRoles_EmptyResult(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("ListRoles", mock.Anything).Return([]model.Role{}, nil)

	roles, err := adapter.ListRoles(ctx, "")

	require.NoError(t, err)
	assert.Empty(t, roles)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_ListRoles_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("database error")

	mockSvc.On("ListRoles", mock.Anything).Return(nil, testError)

	roles, err := adapter.ListRoles(ctx, "")

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, roles)
	mockSvc.AssertExpectations(t)
}

// ============== UpdateRole Tests ==============

func TestAdapter_UpdateRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	existingRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Old description",
	}
	updatedRole := &model.Role{
		ID:          1,
		Name:        "UpdatedAdmin",
		Description: "New description",
	}

	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(existingRole, nil)
	mockSvc.On("UpdateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(updatedRole, nil)

	updates := map[string]interface{}{
		"name":        "UpdatedAdmin",
		"description": "New description",
	}

	err := adapter.UpdateRole(ctx, 1, updates)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_UpdateRole_GetRoleError(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role not found")

	mockSvc.On("GetRoleByID", mock.Anything, 999).Return(nil, testError)

	updates := map[string]interface{}{
		"name": "UpdatedAdmin",
	}

	err := adapter.UpdateRole(ctx, 999, updates)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_UpdateRole_UpdateError(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	existingRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Old description",
	}
	testError := errors.New("update failed")

	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(existingRole, nil)
	mockSvc.On("UpdateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil, testError)

	updates := map[string]interface{}{
		"name": "UpdatedAdmin",
	}

	err := adapter.UpdateRole(ctx, 1, updates)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_UpdateRole_PartialUpdate(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	existingRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Old description",
	}
	updatedRole := &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "New description only",
	}

	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(existingRole, nil)
	mockSvc.On("UpdateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(updatedRole, nil)

	// Only update description, not name
	updates := map[string]interface{}{
		"description": "New description only",
	}

	err := adapter.UpdateRole(ctx, 1, updates)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

// ============== DeleteRole Tests ==============

func TestAdapter_DeleteRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("DeleteRole", mock.Anything, 1).Return(nil)

	err := adapter.DeleteRole(ctx, 1)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_DeleteRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("cannot delete system role")

	mockSvc.On("DeleteRole", mock.Anything, 1).Return(testError)

	err := adapter.DeleteRole(ctx, 1)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

// ============== AssignRole Tests ==============

func TestAdapter_AssignRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("AssignRoleToUser", mock.Anything, 10, 2).Return(nil)

	err := adapter.AssignRole(ctx, 10, 2, "tenant-001")

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_AssignRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role assignment failed")

	mockSvc.On("AssignRoleToUser", mock.Anything, 10, 2).Return(testError)

	err := adapter.AssignRole(ctx, 10, 2, "tenant-001")

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

// ============== RemoveRoleFromUser Tests ==============

func TestAdapter_RemoveRoleFromUser_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("RemoveRoleFromUser", mock.Anything, 10, 2).Return(nil)

	err := adapter.RemoveRoleFromUser(ctx, 10, 2)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_RemoveRoleFromUser_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("role removal failed")

	mockSvc.On("RemoveRoleFromUser", mock.Anything, 10, 2).Return(testError)

	err := adapter.RemoveRoleFromUser(ctx, 10, 2)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

// ============== GetUserRoles Tests ==============

func TestAdapter_GetUserRoles_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedRoles := []model.Role{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "operator"},
	}

	mockSvc.On("ListUserRoles", mock.Anything, 10).Return(expectedRoles, nil)

	roles, err := adapter.GetUserRoles(ctx, 10)

	require.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, expectedRoles, roles)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_GetUserRoles_Empty(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("ListUserRoles", mock.Anything, 10).Return([]model.Role{}, nil)

	roles, err := adapter.GetUserRoles(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, roles)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_GetUserRoles_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("user not found")

	mockSvc.On("ListUserRoles", mock.Anything, 999).Return(nil, testError)

	roles, err := adapter.GetUserRoles(ctx, 999)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, roles)
	mockSvc.AssertExpectations(t)
}

// ============== GetUserPermissions Tests ==============

func TestAdapter_GetUserPermissions_ReturnsEmpty(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// GetUserPermissions returns empty list as per implementation
	permissions, err := adapter.GetUserPermissions(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, permissions)
	assert.Equal(t, []model.Permission{}, permissions)
}

// ============== CheckPermission Tests ==============

func TestAdapter_CheckPermission_ReturnsFalse(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// CheckPermission returns false as per implementation
	allowed, err := adapter.CheckPermission(ctx, 10, "devices", "read")

	require.NoError(t, err)
	assert.False(t, allowed)
}

// ============== CreatePermission Tests ==============

func TestAdapter_CreatePermission_ReturnsNotImplemented(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// CreatePermission returns ErrNotImplemented
	permission, err := adapter.CreatePermission(ctx, "test.perm", "test", "read", "Test permission")

	require.Error(t, err)
	assert.Equal(t, ErrNotImplemented, err)
	assert.Nil(t, permission)
}

// ============== GetPermission Tests ==============

func TestAdapter_GetPermission_ReturnsNotImplemented(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// GetPermission returns ErrNotImplemented
	permission, err := adapter.GetPermission(ctx, 1)

	require.Error(t, err)
	assert.Equal(t, ErrNotImplemented, err)
	assert.Nil(t, permission)
}

// ============== ListPermissions Tests ==============

func TestAdapter_ListPermissions_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	expectedPermissions := []model.Permission{
		{ID: 1, Name: "devices.read", Resource: "devices", Action: "read"},
		{ID: 2, Name: "devices.manage", Resource: "devices", Action: "manage"},
	}

	mockSvc.On("ListPermissions", mock.Anything).Return(expectedPermissions, nil)

	permissions, err := adapter.ListPermissions(ctx)

	require.NoError(t, err)
	assert.Len(t, permissions, 2)
	assert.Equal(t, expectedPermissions, permissions)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_ListPermissions_Empty(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("ListPermissions", mock.Anything).Return([]model.Permission{}, nil)

	permissions, err := adapter.ListPermissions(ctx)

	require.NoError(t, err)
	assert.Empty(t, permissions)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_ListPermissions_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("database error")

	mockSvc.On("ListPermissions", mock.Anything).Return(nil, testError)

	permissions, err := adapter.ListPermissions(ctx)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Nil(t, permissions)
	mockSvc.AssertExpectations(t)
}

// ============== DeletePermission Tests ==============

func TestAdapter_DeletePermission_ReturnsNotImplemented(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// DeletePermission returns ErrNotImplemented
	err := adapter.DeletePermission(ctx, 1)

	require.Error(t, err)
	assert.Equal(t, ErrNotImplemented, err)
}

// ============== AssignPermissionToRole Tests ==============

func TestAdapter_AssignPermissionToRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("AssignPermissionToRole", mock.Anything, 1, 10).Return(nil)

	err := adapter.AssignPermissionToRole(ctx, 1, 10)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_AssignPermissionToRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("permission assignment failed")

	mockSvc.On("AssignPermissionToRole", mock.Anything, 1, 10).Return(testError)

	err := adapter.AssignPermissionToRole(ctx, 1, 10)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

// ============== RemovePermissionFromRole Tests ==============

func TestAdapter_RemovePermissionFromRole_Success(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	mockSvc.On("RemovePermissionFromRole", mock.Anything, 1, 10).Return(nil)

	err := adapter.RemovePermissionFromRole(ctx, 1, 10)

	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAdapter_RemovePermissionFromRole_Error(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()
	testError := errors.New("permission removal failed")

	mockSvc.On("RemovePermissionFromRole", mock.Anything, 1, 10).Return(testError)

	err := adapter.RemovePermissionFromRole(ctx, 1, 10)

	require.Error(t, err)
	assert.Equal(t, testError, err)
	mockSvc.AssertExpectations(t)
}

// ============== GetRolePermissions Tests ==============

func TestAdapter_GetRolePermissions_ReturnsEmpty(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// GetRolePermissions returns empty list as per implementation
	permissions, err := adapter.GetRolePermissions(ctx, 1)

	require.NoError(t, err)
	assert.Empty(t, permissions)
	assert.Equal(t, []model.Permission{}, permissions)
}

// ============== NewRBACServiceAdapter Tests ==============

func TestNewRBACServiceAdapter_NilService(t *testing.T) {
	adapter := NewRBACServiceAdapter(nil)

	// Adapter is created but with nil internal service
	// Note: The adapter allows nil service, callers should handle nil service gracefully
	assert.NotNil(t, adapter)
}

func TestNewRBACServiceAdapter_ValidService(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc)

	assert.NotNil(t, adapter)

	// Verify it's the correct type
	_, ok := adapter.(*rbacServiceAdapter)
	assert.True(t, ok)
}

// ============== Integration-style Tests ==============

func TestAdapter_FullWorkflow(t *testing.T) {
	mockSvc := new(MockRBACServiceInternal)
	adapter := NewRBACServiceAdapter(mockSvc).(*rbacServiceAdapter)

	ctx := context.Background()

	// Test CreateRole
	createdRole := &model.Role{ID: 1, Name: "test-role", TenantID: "tenant-001"}
	mockSvc.On("CreateRole", mock.Anything, mock.AnythingOfType("*model.Role")).Return(createdRole, nil)

	role, err := adapter.CreateRole(ctx, "tenant-001", "test-role", "test-role", "Test role")
	require.NoError(t, err)
	assert.Equal(t, createdRole, role)

	// Test GetRole
	mockSvc.On("GetRoleByID", mock.Anything, 1).Return(createdRole, nil)

	role, err = adapter.GetRole(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, createdRole, role)

	// Test AssignRole
	mockSvc.On("AssignRoleToUser", mock.Anything, 100, 1).Return(nil)

	err = adapter.AssignRole(ctx, 100, 1, "tenant-001")
	require.NoError(t, err)

	// Test GetUserRoles
	mockSvc.On("ListUserRoles", mock.Anything, 100).Return([]model.Role{*createdRole}, nil)

	roles, err := adapter.GetUserRoles(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, roles, 1)

	// Test RemoveRoleFromUser
	mockSvc.On("RemoveRoleFromUser", mock.Anything, 100, 1).Return(nil)

	err = adapter.RemoveRoleFromUser(ctx, 100, 1)
	require.NoError(t, err)

	// Test DeleteRole
	mockSvc.On("DeleteRole", mock.Anything, 1).Return(nil)

	err = adapter.DeleteRole(ctx, 1)
	require.NoError(t, err)

	mockSvc.AssertExpectations(t)
}
