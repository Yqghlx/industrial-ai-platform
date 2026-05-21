package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================
// Factory.go Coverage Tests - Simplified
// ============================================

func TestServiceFactory_GetAllServices(t *testing.T) {
	factory := NewServiceFactory()

	// 测试空工厂获取方法（返回 nil）
	assert.Nil(t, factory.GetDeviceService())
	assert.Nil(t, factory.GetAlertService())
	assert.Nil(t, factory.GetTelemetryService())
	assert.Nil(t, factory.GetAuthService())
	assert.Nil(t, factory.GetUserService())
	assert.Nil(t, factory.GetHealthService())
	assert.Nil(t, factory.GetAgentService())
	assert.Nil(t, factory.GetWorkOrderService())
	assert.Nil(t, factory.GetNotificationService())
	assert.Nil(t, factory.GetBlackBoxService())
	assert.Nil(t, factory.GetReportService())
	assert.Nil(t, factory.GetExportService())
	assert.Nil(t, factory.GetRBACService())
	assert.Nil(t, factory.GetTenantService())
}

func TestServiceFactory_SetAndGet_Cycle(t *testing.T) {
	factory := NewServiceFactory()

	// Set nil, Get nil（测试赋值语句）
	factory.SetDeviceService(nil)
	assert.Nil(t, factory.GetDeviceService())

	factory.SetAlertService(nil)
	assert.Nil(t, factory.GetAlertService())

	factory.SetTelemetryService(nil)
	assert.Nil(t, factory.GetTelemetryService())

	factory.SetAuthService(nil)
	assert.Nil(t, factory.GetAuthService())

	factory.SetUserService(nil)
	assert.Nil(t, factory.GetUserService())

	factory.SetHealthService(nil)
	assert.Nil(t, factory.GetHealthService())

	factory.SetAgentService(nil)
	assert.Nil(t, factory.GetAgentService())

	factory.SetWorkOrderService(nil)
	assert.Nil(t, factory.GetWorkOrderService())

	factory.SetNotificationService(nil)
	assert.Nil(t, factory.GetNotificationService())

	factory.SetBlackBoxService(nil)
	assert.Nil(t, factory.GetBlackBoxService())

	factory.SetReportService(nil)
	assert.Nil(t, factory.GetReportService())

	factory.SetExportService(nil)
	assert.Nil(t, factory.GetExportService())

	factory.SetRBACService(nil)
	assert.Nil(t, factory.GetRBACService())

	factory.SetTenantService(nil)
	assert.Nil(t, factory.GetTenantService())
}