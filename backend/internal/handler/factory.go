package handler

import (
	"errors"

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
		f.serviceFactory.GetTelemetryService(),
		f.serviceFactory.GetConfigService(),
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
		f.serviceFactory.GetDeviceService(),
		f.broadcastFunc,
		f.cache,
	)
}

// CreateRBACHandler 创建 RBAC Handler
// FIX-005: 使用适配器解决接口不兼容问题
// adapter 模式将 service.RBACServiceInterface 转换为 handler.RBACServiceInterface
func (f *HandlerFactory) CreateRBACHandler() *RBACHandler {
	rbacSvc := f.serviceFactory.GetRBACService()
	if rbacSvc == nil {
		return nil
	}
	// 使用适配器包装 service 层接口
	adapter := NewRBACServiceAdapter(rbacSvc)
	return NewRBACHandler(adapter)
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
// Handler 接口定义
// ============================================

// Handler 定义处理器接口，所有处理器必须实现此接口
type Handler interface {
	// RegisterRoutes 注册路由到路由组
	RegisterRoutes(router *gin.RouterGroup) error
	// Name 返回处理器名称
	Name() string
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

// RegisterAll 注册所有实现了 Handler 接口的处理器到路由引擎
// 返回注册过程中遇到的任何错误
func RegisterAll(router *gin.Engine, handlers ...Handler) error {
	if router == nil {
		return ErrNilRouter
	}

	// 创建一个根路由组
	rootGroup := router.Group("")

	for _, handler := range handlers {
		if handler == nil {
			continue
		}

		if err := handler.RegisterRoutes(rootGroup); err != nil {
			return NewHandlerError(handler.Name(), err)
		}
	}

	return nil
}

// RegisterAllWithGroup 注册所有处理器到指定的路由组
func RegisterAllWithGroup(group *gin.RouterGroup, handlers ...Handler) error {
	if group == nil {
		return ErrNilRouterGroup
	}

	for _, handler := range handlers {
		if handler == nil {
			continue
		}

		if err := handler.RegisterRoutes(group); err != nil {
			return NewHandlerError(handler.Name(), err)
		}
	}

	return nil
}

// ============================================
// 错误定义
// ============================================

var (
	// ErrNilRouter 路由器为空错误
	ErrNilRouter = errors.New("router cannot be nil")
	// ErrNilRouterGroup 路由组为空错误
	ErrNilRouterGroup = errors.New("router group cannot be nil")
)

// HandlerError 处理器注册错误
type HandlerError struct {
	HandlerName string
	Err         error
}

// NewHandlerError 创建处理器错误
func NewHandlerError(name string, err error) *HandlerError {
	return &HandlerError{
		HandlerName: name,
		Err:         err,
	}
}

// Error 实现 error 接口
func (e *HandlerError) Error() string {
	return "handler [" + e.HandlerName + "] registration failed: " + e.Err.Error()
}

// Unwrap 返回底层错误
func (e *HandlerError) Unwrap() error {
	return e.Err
}
