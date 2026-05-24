package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/cache"
)

// ============================================
// Handler 工厂 (简化版 - 用于测试依赖注入)
// ============================================

// HandlerFactory Handler 工厂
// 简化版：主要用于测试场景，通过 ServiceFactory 注入 Service
type HandlerFactory struct {
	serviceFactory *service.ServiceFactory
	broadcastFunc  func(msg model.WSMessage)
	cache          cache.CacheService // 添加缓存
}

// NewHandlerFactory 创建 Handler 工厂
func NewHandlerFactory(sf *service.ServiceFactory, broadcastFunc func(msg model.WSMessage), cacheSvc cache.CacheService) *HandlerFactory {
	return &HandlerFactory{
		serviceFactory: sf,
		broadcastFunc:  broadcastFunc,
		cache:          cacheSvc,
	}
}

// CreateDeviceHandler 创建设备 Handler
func (f *HandlerFactory) CreateDeviceHandler() *DeviceHandlerNew {
	return NewDeviceHandlerNew(
		f.serviceFactory.GetDeviceService(),
		f.serviceFactory.GetAlertService(),
		f.serviceFactory.GetAuthService(),
		f.serviceFactory.GetTelemetryService(),
		f.broadcastFunc,
	)
}

// CreateAlertHandler 创建告警 Handler
func (f *HandlerFactory) CreateAlertHandler() *AlertHandler {
	return NewAlertHandler(
		f.serviceFactory.GetAlertService(),
		f.broadcastFunc,
	)
}

// CreateTelemetryHandler 创建遥测 Handler
func (f *HandlerFactory) CreateTelemetryHandler() *TelemetryHandlerNew {
	return NewTelemetryHandlerNew(
		f.serviceFactory.GetTelemetryService(),
		f.serviceFactory.GetAgentService(),
	)
}

// CreateAuthHandler 创建认证 Handler
func (f *HandlerFactory) CreateAuthHandler() *AuthHandlerNew {
	return NewAuthHandlerNew(
		f.serviceFactory.GetAuthService(),
		f.serviceFactory.GetUserService(),
	)
}

// CreateAdminHandler 创建管理 Handler
func (f *HandlerFactory) CreateAdminHandler() *AdminHandlerNew {
	return NewAdminHandlerNew(
		f.serviceFactory.GetAuthService(),
	)
}

// CreateHealthHandler 创建健康检查 Handler
func (f *HandlerFactory) CreateHealthHandler() *HealthHandlerNew {
	return &HealthHandlerNew{}
}

// CreateExportHandler 创建导出 Handler
func (f *HandlerFactory) CreateExportHandler() *ExportHandler {
	return NewExportHandler(
		f.serviceFactory.GetExportService(),
	)
}

// CreateBusinessHandler 创建业务 Handler
func (f *HandlerFactory) CreateBusinessHandler() *BusinessHandlerNew {
	return NewBusinessHandlerNew(
		f.serviceFactory.GetWorkOrderService(),
		f.serviceFactory.GetNotificationService(),
		f.serviceFactory.GetBlackBoxService(),
		f.serviceFactory.GetReportService(),
		f.serviceFactory.GetAlertService(),
		f.broadcastFunc,
		f.cache,
	)
}

// CreateRBACHandler 创建 RBAC Handler
// FIX-005: 接口不兼容问题 - handler.RBACServiceInterface 与 service.RBACServiceInterface 方法签名不同
// 暂时返回 nil，等待统一接口定义。使用时应检查返回值是否为 nil
func (f *HandlerFactory) CreateRBACHandler() *RBACHandler {
	// 注意: service.RBACServiceInterface 方法签名与 handler.RBACServiceInterface 不一致
	// 例如: service.AssignRoleToUser(userID, roleID) vs handler.AssignRole(userID, roleID, tenantID)
	// 需要创建适配器或统一接口定义后方可启用
	return nil
}

// CreateTenantHandler 创建租户 Handler
// FIX-005: 实现完整逻辑 - service.TenantService 已实现 handler.TenantServiceInterface
func (f *HandlerFactory) CreateTenantHandler() *TenantHandler {
	tenantSvc := f.serviceFactory.GetTenantService()
	if tenantSvc == nil {
		return nil
	}
	// tenant_handler.go 中已验证 service.TenantService 实现了 handler.TenantServiceInterface
	return NewTenantHandler(tenantSvc)
}

// ============================================
// Handler 注册器
// ============================================

// HandlerRegistry Handler 注册器
type HandlerRegistry struct {
	factory *HandlerFactory
	router  *gin.Engine
}

// NewHandlerRegistry 创建 Handler 注册器
func NewHandlerRegistry(factory *HandlerFactory, router *gin.Engine) *HandlerRegistry {
	return &HandlerRegistry{
		factory: factory,
		router:  router,
	}
}

// RegisterAll 注册所有 Handler（简化版）
func (r *HandlerRegistry) RegisterAll() {
	// 使用现有的 setupHandlers 实现
	// 此处仅做接口定义，实际路由注册在 server_new.go 中完成
}
