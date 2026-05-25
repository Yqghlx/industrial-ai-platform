package service

import (
	"github.com/industrial-ai/platform/internal/repository"
)

// ============================================
// Service 工厂 (工厂模式 + 依赖注入)
// ============================================

// ServiceFactory Service 工厂
// 统一管理所有 Service 的创建和依赖注入
type ServiceFactory struct {
	// 已注入的 Service 实例（用于测试和灵活配置）
	deviceService       DeviceServiceInterface
	alertService        AlertServiceInterface
	telemetryService    TelemetryServiceInterface
	authService         AuthServiceInterface
	userService         UserServiceInterface
	healthService       HealthServiceInterface
	agentService        AgentServiceInterface
	workOrderService    WorkOrderServiceInterface
	notificationService NotificationServiceInterface
	blackBoxService     BlackBoxServiceInterface
	reportService       ReportServiceInterface
	exportService       ExportServiceInterface
	rbacService         RBACServiceInterface
	tenantService       TenantServiceInterface
}

// NewServiceFactory 创建 Service 工厂（空工厂）
// 用于测试场景，通过 Set 方法注入 Mock
//
// 注意：此工厂返回空实例，所有 Service 字段为 nil。
// 调用方需要通过 Set* 方法手动注入 Service 实现，
// 或使用 NewServiceFactoryFromRepo 创建完整工厂。
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// NewServiceFactoryFromRepo 创建 Service 工厂（从 Repository 创建）
//
// DESIGN DECISION: 当前返回空工厂，原因如下：
// 1. 服务依赖复杂，需要逐步完善初始化逻辑
// 2. 避免循环依赖，某些服务需要其他服务作为依赖
// 3. 生产环境应使用完整的依赖注入框架（如 wire、dig）
//
// TODO: 实现完整的 Service 初始化
// - 实例化基础服务（无依赖）
// - 按依赖顺序初始化其他服务
// - 或引入 DI 框架自动管理依赖
func NewServiceFactoryFromRepo(repoFactory *repository.RepositoryFactory) *ServiceFactory {
	return &ServiceFactory{}
}

// ============================================
// Service 设置方法（用于依赖注入）
// ============================================

func (f *ServiceFactory) SetDeviceService(svc DeviceServiceInterface)       { f.deviceService = svc }
func (f *ServiceFactory) SetAlertService(svc AlertServiceInterface)         { f.alertService = svc }
func (f *ServiceFactory) SetTelemetryService(svc TelemetryServiceInterface) { f.telemetryService = svc }
func (f *ServiceFactory) SetAuthService(svc AuthServiceInterface)           { f.authService = svc }
func (f *ServiceFactory) SetUserService(svc UserServiceInterface)           { f.userService = svc }
func (f *ServiceFactory) SetHealthService(svc HealthServiceInterface)       { f.healthService = svc }
func (f *ServiceFactory) SetAgentService(svc AgentServiceInterface)         { f.agentService = svc }
func (f *ServiceFactory) SetWorkOrderService(svc WorkOrderServiceInterface) { f.workOrderService = svc }
func (f *ServiceFactory) SetNotificationService(svc NotificationServiceInterface) {
	f.notificationService = svc
}
func (f *ServiceFactory) SetBlackBoxService(svc BlackBoxServiceInterface) { f.blackBoxService = svc }
func (f *ServiceFactory) SetReportService(svc ReportServiceInterface)     { f.reportService = svc }
func (f *ServiceFactory) SetExportService(svc ExportServiceInterface)     { f.exportService = svc }
func (f *ServiceFactory) SetRBACService(svc RBACServiceInterface)         { f.rbacService = svc }
func (f *ServiceFactory) SetTenantService(svc TenantServiceInterface)     { f.tenantService = svc }

// ============================================
// Service 获取方法（单例模式）
// ============================================

func (f *ServiceFactory) GetDeviceService() DeviceServiceInterface       { return f.deviceService }
func (f *ServiceFactory) GetAlertService() AlertServiceInterface         { return f.alertService }
func (f *ServiceFactory) GetTelemetryService() TelemetryServiceInterface { return f.telemetryService }
func (f *ServiceFactory) GetAuthService() AuthServiceInterface           { return f.authService }
func (f *ServiceFactory) GetUserService() UserServiceInterface           { return f.userService }
func (f *ServiceFactory) GetHealthService() HealthServiceInterface       { return f.healthService }
func (f *ServiceFactory) GetAgentService() AgentServiceInterface         { return f.agentService }
func (f *ServiceFactory) GetWorkOrderService() WorkOrderServiceInterface { return f.workOrderService }
func (f *ServiceFactory) GetNotificationService() NotificationServiceInterface {
	return f.notificationService
}
func (f *ServiceFactory) GetBlackBoxService() BlackBoxServiceInterface { return f.blackBoxService }
func (f *ServiceFactory) GetReportService() ReportServiceInterface     { return f.reportService }
func (f *ServiceFactory) GetExportService() ExportServiceInterface     { return f.exportService }
func (f *ServiceFactory) GetRBACService() RBACServiceInterface         { return f.rbacService }
func (f *ServiceFactory) GetTenantService() TenantServiceInterface     { return f.tenantService }

// ============================================
// ReportServiceInterface 定义
// ============================================

// 注意：ReportServiceInterface 已在 report_service.go 中定义
// 这里不重复定义

// ============================================
// ExportServiceInterface 定义（避免重复）
// ============================================

// 注意：ExportServiceInterface 已在 export_service.go 中定义
// 这里不重复定义，使用 export_service.go 中的定义
