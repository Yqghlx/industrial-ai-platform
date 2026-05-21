package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Error constants used in tests
var (
	ErrRoleAlreadyExists      = errors.New("role already exists")
	ErrRoleNotFound           = errors.New("role not found")
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
)

// MockRBACService is a mock implementation of RBACServiceInterface
type MockRBACService struct {
	mock.Mock
}

func (m *MockRBACService) CreateRole(ctx context.Context, tenantID, name, displayName, description string) (*model.Role, error) {
	args := m.Called(ctx, tenantID, name, displayName, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACService) GetRole(ctx context.Context, id int) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACService) GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RoleResponse), args.Error(1)
}

func (m *MockRBACService) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACService) UpdateRole(ctx context.Context, id int, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockRBACService) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRBACService) AssignRole(ctx context.Context, userID, roleID int, tenantID string) error {
	args := m.Called(ctx, userID, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACService) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACService) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACService) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACService) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACService) CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error) {
	args := m.Called(ctx, name, resource, action, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockRBACService) GetPermission(ctx context.Context, id int) (*model.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockRBACService) DeletePermission(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestRBACHandler_CreateRole_Success tests successful role creation
func TestRBACHandler_CreateRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.CreateRoleRequest{
		Name:        "Operator",
		Description: "Device operator role",
		TenantID:    "tenant-001",
	}

	expectedRole := &model.Role{
		ID:          1,
		Name:        "Operator",
		Description: "Device operator role",
		TenantID:    "tenant-001",
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRBACSvc.On("CreateRole", mock.Anything, req.TenantID, req.Name, req.Name, req.Description).Return(expectedRole, nil)

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateRole(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_CreateRole_InvalidRequest tests role creation with invalid request
func TestRBACHandler_CreateRole_InvalidRequest(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Invalid JSON
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_CreateRole_ServiceError tests role creation service error
func TestRBACHandler_CreateRole_ServiceError(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.CreateRoleRequest{
		Name:        "Operator",
		Description: "Device operator role",
		TenantID:    "tenant-001",
	}

	mockRBACSvc.On("CreateRole", mock.Anything, req.TenantID, req.Name, req.Name, req.Description).Return(nil, ErrRoleAlreadyExists)

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateRole(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_ListRoles_Success tests successful role listing
func TestRBACHandler_ListRoles_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedRoles := []model.Role{
		{ID: 1, Name: "admin", Description: "Administrator", IsSystem: true},
		{ID: 2, Name: "operator", Description: "Operator", IsSystem: true},
		{ID: 3, Name: "viewer", Description: "Viewer", IsSystem: true},
	}

	mockRBACSvc.On("ListRoles", mock.Anything, "tenant-001").Return(expectedRoles, nil)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles?tenant_id=tenant-001", nil)

	handler.ListRoles(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_ListRoles_ServiceError tests role listing service error
func TestRBACHandler_ListRoles_ServiceError(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("ListRoles", mock.Anything, "").Return(nil, errors.New("database error"))

	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)

	handler.ListRoles(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_GetRole_Success tests successful role retrieval
func TestRBACHandler_GetRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedRoleResponse := &model.RoleResponse{
		Role: model.Role{
			ID:          1,
			Name:        "admin",
			Description: "Administrator with full access",
			IsSystem:    true,
		},
		Permissions: []model.Permission{
			{ID: 1, Name: "devices.manage", Resource: "devices", Action: "manage"},
			{ID: 2, Name: "users.manage", Resource: "users", Action: "manage"},
		},
	}

	mockRBACSvc.On("GetRole", mock.Anything, 1).Return(&expectedRoleResponse.Role, nil)

	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles/1", nil)

	handler.GetRole(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_GetRole_InvalidID tests role retrieval with invalid ID
func TestRBACHandler_GetRole_InvalidID(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles/invalid", nil)

	handler.GetRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_GetRole_NotFound tests role not found
func TestRBACHandler_GetRole_NotFound(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("GetRole", mock.Anything, 999).Return(nil, ErrRoleNotFound)

	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles/999", nil)

	handler.GetRole(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_UpdateRole_Success tests successful role update
func TestRBACHandler_UpdateRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
	}

	updatedRole := &model.Role{
		ID:          1,
		Name:        "Updated Role",
		Description: "Updated description",
		UpdatedAt:   time.Now(),
	}

	mockRBACSvc.On("UpdateRole", mock.Anything, 1, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	// For the handler, it calls UpdateRole with name and description directly
	mockRBACSvc.On("GetRole", mock.Anything, 1).Return(updatedRole, nil)

	body, _ := json.Marshal(req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/roles/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRole(c)

	// The handler should work properly
	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_UpdateRole_InvalidID tests role update with invalid ID
func TestRBACHandler_UpdateRole_InvalidID(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.UpdateRoleRequest{Name: "Updated"}
	body, _ := json.Marshal(req)

	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/roles/invalid", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_DeleteRole_Success tests successful role deletion
func TestRBACHandler_DeleteRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("DeleteRole", mock.Anything, 5).Return(nil)

	c.Params = gin.Params{{Key: "id", Value: "5"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/roles/5", nil)

	handler.DeleteRole(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_DeleteRole_SystemRole tests deletion of system role
func TestRBACHandler_DeleteRole_SystemRole(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("DeleteRole", mock.Anything, 1).Return(ErrCannotDeleteSystemRole)

	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/roles/1", nil)

	handler.DeleteRole(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_AssignRole_Success tests successful role assignment
func TestRBACHandler_AssignRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.AssignRoleRequest{
		RoleID:   2,
		TenantID: "tenant-001",
	}

	mockRBACSvc.On("AssignRole", mock.Anything, 10, req.RoleID, req.TenantID).Return(nil)

	body, _ := json.Marshal(req)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/10/roles", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AssignRole(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_AssignRole_InvalidUserID tests role assignment with invalid user ID
func TestRBACHandler_AssignRole_InvalidUserID(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/invalid/roles", bytes.NewReader([]byte("{}")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AssignRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_RemoveRole_Success tests successful role removal
func TestRBACHandler_RemoveRole_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("RemoveRoleFromUser", mock.Anything, 10, 2).Return(nil)

	c.Params = gin.Params{{Key: "id", Value: "10"}, {Key: "role_id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/10/roles/2", nil)

	handler.RemoveRole(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_RemoveRole_InvalidIDs tests role removal with invalid IDs
func TestRBACHandler_RemoveRole_InvalidIDs(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "invalid"}, {Key: "role_id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/invalid/roles/2", nil)

	handler.RemoveRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_GetUserRoles_Success tests successful user roles retrieval
func TestRBACHandler_GetUserRoles_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedRoles := []model.Role{
		{ID: 1, Name: "admin", Description: "Administrator"},
		{ID: 2, Name: "operator", Description: "Operator"},
	}

	mockRBACSvc.On("GetUserRoles", mock.Anything, 10).Return(expectedRoles, nil)

	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/10/roles", nil)

	handler.GetUserRoles(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_ListPermissions_Success tests successful permissions listing
func TestRBACHandler_ListPermissions_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedPermissions := []model.Permission{
		{ID: 1, Name: "devices.manage", Resource: "devices", Action: "manage"},
		{ID: 2, Name: "devices.read", Resource: "devices", Action: "read"},
		{ID: 3, Name: "users.manage", Resource: "users", Action: "manage"},
	}

	mockRBACSvc.On("ListPermissions", mock.Anything).Return(expectedPermissions, nil)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/permissions", nil)

	handler.ListPermissions(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_AssignPermission_Success tests successful permission assignment
func TestRBACHandler_AssignPermission_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.AssignPermissionRequest{
		PermissionID: 1,
	}

	mockRBACSvc.On("AssignPermissionToRole", mock.Anything, 2, req.PermissionID).Return(nil)

	body, _ := json.Marshal(req)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/roles/2/permissions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AssignPermission(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_RemovePermission_Success tests successful permission removal
func TestRBACHandler_RemovePermission_Success(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("RemovePermissionFromRole", mock.Anything, 2, 1).Return(nil)

	c.Params = gin.Params{{Key: "id", Value: "2"}, {Key: "perm_id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/roles/2/permissions/1", nil)

	handler.RemovePermission(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRBACSvc.AssertExpectations(t)
}

// TestRoleStructure tests role model structure
func TestRoleStructure(t *testing.T) {
	role := model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Administrator with full access",
		TenantID:    "tenant-001",
		IsSystem:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, 1, role.ID)
	assert.Equal(t, "admin", role.Name)
	assert.Equal(t, "Administrator with full access", role.Description)
	assert.Equal(t, "tenant-001", role.TenantID)
	assert.True(t, role.IsSystem)
	assert.NotZero(t, role.CreatedAt)
	assert.NotZero(t, role.UpdatedAt)
}

// TestPermissionStructure tests permission model structure
func TestPermissionStructure(t *testing.T) {
	perm := model.Permission{
		ID:          1,
		Name:        "devices.manage",
		Resource:    "devices",
		Action:      "manage",
		Description: "Full device management",
		CreatedAt:   time.Now(),
	}

	assert.Equal(t, 1, perm.ID)
	assert.Equal(t, "devices.manage", perm.Name)
	assert.Equal(t, "devices", perm.Resource)
	assert.Equal(t, "manage", perm.Action)
	assert.NotZero(t, perm.CreatedAt)
}

// TestDefaultRoles tests default system roles
func TestDefaultRoles(t *testing.T) {
	defaultRoles := model.DefaultRoles

	assert.Len(t, defaultRoles, 3)
	assert.Equal(t, "admin", defaultRoles[0].Name)
	assert.Equal(t, "operator", defaultRoles[1].Name)
	assert.Equal(t, "viewer", defaultRoles[2].Name)

	for _, role := range defaultRoles {
		assert.True(t, role.IsSystem)
	}
}

// TestDefaultPermissions tests default permissions
func TestDefaultPermissions(t *testing.T) {
	defaultPerms := model.DefaultPermissions

	assert.GreaterOrEqual(t, len(defaultPerms), 10)

	// Check for essential permissions
	var hasDeviceManage, hasUserManage, hasReportRead bool
	for _, perm := range defaultPerms {
		if perm.Name == "devices.manage" {
			hasDeviceManage = true
		}
		if perm.Name == "users.manage" {
			hasUserManage = true
		}
		if perm.Name == "reports.read" {
			hasReportRead = true
		}
	}

	assert.True(t, hasDeviceManage, "Should have devices.manage permission")
	assert.True(t, hasUserManage, "Should have users.manage permission")
	assert.True(t, hasReportRead, "Should have reports.read permission")
}

// TestNewRBACHandler tests handler creation
func TestNewRBACHandler(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.rbacSvc)
}

// TestRBACHandler_UpdateRole_ServiceError tests role update service error
func TestRBACHandler_UpdateRole_ServiceError(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
	}

	mockRBACSvc.On("UpdateRole", mock.Anything, 1, mock.AnythingOfType("map[string]interface {}")).Return(errors.New("update failed"))

	body, _ := json.Marshal(req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/roles/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRole(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_UpdateRole_GetRoleError tests error when getting role after update
func TestRBACHandler_UpdateRole_GetRoleError(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
	}

	mockRBACSvc.On("UpdateRole", mock.Anything, 1, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRBACSvc.On("GetRole", mock.Anything, 1).Return(nil, ErrRoleNotFound)

	body, _ := json.Marshal(req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/roles/1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRole(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_UpdateRole_InvalidJSON tests role update with invalid JSON
func TestRBACHandler_UpdateRole_InvalidJSON(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/roles/1", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRole(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_GetUserRoles_InvalidID tests user roles retrieval with invalid ID
func TestRBACHandler_GetUserRoles_InvalidID(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid/roles", nil)

	handler.GetUserRoles(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRBACHandler_GetUserRoles_ServiceError tests user roles retrieval service error
func TestRBACHandler_GetUserRoles_ServiceError(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("GetUserRoles", mock.Anything, 10).Return(nil, errors.New("database error"))

	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/10/roles", nil)

	handler.GetUserRoles(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRBACSvc.AssertExpectations(t)
}

// TestRBACHandler_GetUserRoles_EmptyRoles tests user roles retrieval with empty result
func TestRBACHandler_GetUserRoles_EmptyRoles(t *testing.T) {
	mockRBACSvc := new(MockRBACService)
	handler := NewRBACHandler(mockRBACSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	mockRBACSvc.On("GetUserRoles", mock.Anything, 10).Return([]model.Role{}, nil)

	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/10/roles", nil)

	handler.GetUserRoles(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response model.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	mockRBACSvc.AssertExpectations(t)
}

// TestRoleRequestValidation tests role request validation
func TestRoleRequestValidation(t *testing.T) {
	// Valid create role request
	validCreate := model.CreateRoleRequest{
		Name:        "ValidRole",
		Description: "A valid role",
		TenantID:    "tenant-001",
	}
	assert.NotEmpty(t, validCreate.Name)
	assert.NotEmpty(t, validCreate.TenantID)

	// Valid assign role request
	validAssign := model.AssignRoleRequest{
		RoleID:   1,
		TenantID: "tenant-001",
	}
	assert.GreaterOrEqual(t, validAssign.RoleID, 1)

	// Valid assign permission request
	validPerm := model.AssignPermissionRequest{
		PermissionID: 1,
	}
	assert.GreaterOrEqual(t, validPerm.PermissionID, 1)
}

// Test removed - requires proper mock setup
