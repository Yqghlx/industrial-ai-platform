package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// HandlerFactory Tests
// ============================================

func TestNewHandlerFactory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))

	require.NotNil(t, factory)
	assert.NotNil(t, factory.serviceFactory)
	assert.NotNil(t, factory.broadcastFunc)
}

func TestHandlerFactory_CreateDeviceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateDeviceHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateAlertHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateAlertHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateTelemetryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateTelemetryHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateAuthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateAuthHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateAdminHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateAdminHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateHealthHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateExportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateExportHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateBusinessHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateBusinessHandler()

	assert.NotNil(t, handler)
}

func TestHandlerFactory_CreateRBACHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateRBACHandler()

	// TODO: returns nil until unified interface
	assert.Nil(t, handler)
}

func TestHandlerFactory_CreateTenantHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	handler := factory.CreateTenantHandler()

	// TODO: returns nil until unified interface
	assert.Nil(t, handler)
}

// ============================================
// HandlerRegistry Tests
// ============================================

func TestNewHandlerRegistry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	router := gin.New()

	registry := NewHandlerRegistry(factory, router)

	require.NotNil(t, registry)
	assert.Equal(t, factory, registry.factory)
	assert.Equal(t, router, registry.router)
}

func TestHandlerRegistry_RegisterAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSF := service.NewServiceFactory()
	broadcastFunc := func(msg model.WSMessage) {}

factory := NewHandlerFactory(mockSF, broadcastFunc, new(MockCache))
	router := gin.New()

	registry := NewHandlerRegistry(factory, router)

	// RegisterAll is a stub in simplified version
	registry.RegisterAll()

	// No routes registered in simplified version
	assert.Equal(t, 0, len(router.Routes()))
}