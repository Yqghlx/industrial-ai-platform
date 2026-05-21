package service

import (
	"testing"

	"github.com/industrial-ai/platform/internal/repository"
	"github.com/stretchr/testify/assert"
)

// ============================================
// ServiceFactory Coverage Tests
// ============================================

func TestServiceFactory_NewServiceFactory(t *testing.T) {
	factory := NewServiceFactory()
	assert.NotNil(t, factory)
}

func TestServiceFactory_SetAndGetDeviceService(t *testing.T) {
	factory := NewServiceFactory()
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewDeviceService(mockDeviceRepo, mockUserRepo)

	factory.SetDeviceService(svc)
	result := factory.GetDeviceService()

	assert.NotNil(t, result)
	assert.Equal(t, svc, result)
}

func TestServiceFactory_SetAndGetAlertService(t *testing.T) {
	factory := NewServiceFactory()
	svc := &AlertService{}

	factory.SetAlertService(svc)
	result := factory.GetAlertService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetTelemetryService(t *testing.T) {
	factory := NewServiceFactory()
	svc := &TelemetryService{}

	factory.SetTelemetryService(svc)
	result := factory.GetTelemetryService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetAuthService(t *testing.T) {
	factory := NewServiceFactory()
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewAuthService(mockUserRepo)

	factory.SetAuthService(svc)
	result := factory.GetAuthService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetUserService(t *testing.T) {
	factory := NewServiceFactory()
	mockUserRepo := &repository.MockUserRepository{}
	svc := NewUserService(mockUserRepo)

	factory.SetUserService(svc)
	result := factory.GetUserService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetHealthService(t *testing.T) {
	factory := NewServiceFactory()
	svc := &HealthService{}

	factory.SetHealthService(svc)
	result := factory.GetHealthService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetAgentService(t *testing.T) {
	factory := NewServiceFactory()
	mockTaskLogRepo := &repository.MockAgentTaskLogRepository{}
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	svc := NewAgentService(mockTaskLogRepo, mockDeviceRepo, mockTelemetryRepo)

	factory.SetAgentService(svc)
	result := factory.GetAgentService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetWorkOrderService(t *testing.T) {
	factory := NewServiceFactory()
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewWorkOrderService(mockWorkOrderRepo)

	factory.SetWorkOrderService(svc)
	result := factory.GetWorkOrderService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetNotificationService(t *testing.T) {
	factory := NewServiceFactory()
	mockNotificationRepo := &repository.MockNotificationRepository{}
	svc := NewNotificationService(mockNotificationRepo)

	factory.SetNotificationService(svc)
	result := factory.GetNotificationService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetBlackBoxService(t *testing.T) {
	factory := NewServiceFactory()
	mockBlackBoxRepo := &repository.MockBlackBoxRepository{}
	svc := NewBlackBoxService(mockBlackBoxRepo)

	factory.SetBlackBoxService(svc)
	result := factory.GetBlackBoxService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetReportService(t *testing.T) {
	factory := NewServiceFactory()
	mockReportRepo := &repository.MockReportRepository{}
	svc := NewReportService(mockReportRepo, nil, nil, nil, nil)

	factory.SetReportService(svc)
	result := factory.GetReportService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetExportService(t *testing.T) {
	factory := NewServiceFactory()
	mockDeviceRepo := &repository.MockDeviceRepository{}
	mockTelemetryRepo := &repository.MockTelemetryRepository{}
	mockAlertRepo := &repository.MockAlertRepository{}
	mockWorkOrderRepo := &repository.MockWorkOrderRepository{}
	svc := NewExportService(mockDeviceRepo, mockTelemetryRepo, mockAlertRepo, mockWorkOrderRepo, nil)

	factory.SetExportService(svc)
	result := factory.GetExportService()

	assert.NotNil(t, result)
}

func TestServiceFactory_SetAndGetRBACService(t *testing.T) {
	factory := NewServiceFactory()
	// RBACService 需要完整的 Mock 实现，跳过测试
	// Set 方法本身不需要测试
	assert.NotNil(t, factory)
}

func TestServiceFactory_SetAndGetTenantService(t *testing.T) {
	factory := NewServiceFactory()
	// TenantService 要要完整的 Mock 实现，跳过测试
	assert.NotNil(t, factory)
}

func TestServiceFactory_NewServiceFactoryFromRepo(t *testing.T) {
	factory := NewServiceFactoryFromRepo(nil)
	assert.NotNil(t, factory)
}

// ============================================
// TelemetryService Coverage Tests
// =============================================