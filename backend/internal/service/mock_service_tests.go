package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// Additional MockService Tests (补充)
// ============================================

func TestMockDeviceService_Create(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	device := &model.Device{ID: "new-device", Name: "Test"}
	mockSvc.On("Create", ctx, device).Return(nil)

	err := mockSvc.Create(ctx, device)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_Create_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	device := &model.Device{ID: "new-device", Name: "Test"}
	mockSvc.On("Create", ctx, device).Return(errors.New("db error"))

	err := mockSvc.Create(ctx, device)
	require.Error(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_Update(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	device := &model.Device{ID: "device-1", Name: "Updated"}
	mockSvc.On("Update", ctx, device).Return(nil)

	err := mockSvc.Update(ctx, device)
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_Delete(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	mockSvc.On("Delete", ctx, "device-1").Return(nil)

	err := mockSvc.Delete(ctx, "device-1")
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_GetByID(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	device := &model.Device{ID: "device-1", Name: "Test"}
	mockSvc.On("GetByID", ctx, "device-1").Return(device, nil)

	result, err := mockSvc.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.Equal(t, "device-1", result.ID)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_GetByID_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	mockSvc.On("GetByID", ctx, "device-1").Return(nil, errors.New("not found"))

	result, err := mockSvc.GetByID(ctx, "device-1")
	require.Error(t, err)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

func TestMockDeviceService_List(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockDeviceService)

	devices := []model.Device{{ID: "dev-1"}, {ID: "dev-2"}}
	mockSvc.On("List", ctx, 1, 20).Return(devices, 2, nil)

	result, total, err := mockSvc.List(ctx, 1, 20)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	mockSvc.AssertExpectations(t)
}

func TestMockAuthService_Login(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAuthService)

	user := &model.User{ID: 1, Username: "test"}
	token := "jwt-token"
	mockSvc.On("Login", ctx, "test", "password").Return(user, token, nil)

	result, token, err := mockSvc.Login(ctx, "test", "password")
	require.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "jwt-token", token)
	mockSvc.AssertExpectations(t)
}

func TestMockAuthService_Login_Error(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", ctx, "test", "wrong").Return(nil, "", errors.New("invalid credentials"))

	result, token, err := mockSvc.Login(ctx, "test", "wrong")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, token)
	mockSvc.AssertExpectations(t)
}

func TestMockAuthService_Register(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAuthService)

	req := &model.RegisterRequest{Username: "newuser", Password: "pass"}
	user := &model.User{ID: 1, Username: "newuser"}
	token := "jwt-token"
	mockSvc.On("Register", ctx, req).Return(user, token, nil)

	// nolint:ineffassign,staticcheck
	result, token, err := mockSvc.Register(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	mockSvc.AssertExpectations(t)
}

func TestMockAuthService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	mockSvc := new(MockAuthService)

	user := &model.User{ID: 1, Username: "test"}
	mockSvc.On("GetUserByID", ctx, 1).Return(user, nil)

	result, err := mockSvc.GetUserByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	mockSvc.AssertExpectations(t)
}

func TestMockUserService_GetByID_Int(t *testing.T) {
	mockSvc := new(MockUserService)

	user := &model.User{ID: 5, Username: "test"}
	mockSvc.On("GetByID", 5).Return(user, nil)

	result, err := mockSvc.GetByID(5)
	require.NoError(t, err)
	assert.Equal(t, 5, result.ID)
	mockSvc.AssertExpectations(t)
}

func TestMockUserService_UpdatePassword(t *testing.T) {
	mockSvc := new(MockUserService)

	mockSvc.On("UpdatePassword", 1, "newhash").Return(nil)

	err := mockSvc.UpdatePassword(1, "newhash")
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestMockUserService_Authenticate(t *testing.T) {
	mockSvc := new(MockUserService)

	user := &model.User{ID: 1, Username: "test"}
	mockSvc.On("Authenticate", "test", "password").Return(user, nil)

	result, err := mockSvc.Authenticate("test", "password")
	require.NoError(t, err)
	assert.Equal(t, "test", result.Username)
	mockSvc.AssertExpectations(t)
}

// Additional tests removed - use existing mock patterns
