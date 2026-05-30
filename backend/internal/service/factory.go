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
// P1-09: 实现完整的Service初始化逻辑（简化版）
// 注意：部分服务因依赖复杂或返回值类型不匹配，保持nil，可通过Set方法注入
func NewServiceFactoryFromRepo(repoFactory *repository.RepositoryFactory) *ServiceFactory {
	if repoFactory == nil {
		return &ServiceFactory{}
	}

	factory := &ServiceFactory{}

	// 按依赖顺序初始化服务
	// 1. 基础服务（无其他服务依赖）
	factory.deviceService = NewDeviceService(
		repoFactory.GetDeviceRepository(),
		repoFactory.GetUserRepository(),
	)
	factory.userService = NewUserService(repoFactory.GetUserRepository())
	factory.authService = NewAuthService(repoFactory.GetUserRepository())

	// TenantService: GetTenantRepo 返回 (*TenantRepo, error)，需要错误处理
	tenantRepo, err := repoFactory.GetTenantRepo()
	if err == nil && tenantRepo != nil {
		factory.tenantService = NewTenantService(tenantRepo)
	}

		// RBACService: 使用 RBACRepositoryInterface 初始化
		rbacRepo, rbacErr := repoFactory.GetRBACRepository()
		if rbacErr == nil && rbacRepo != nil {
			factory.rbacService = NewRBACService(rbacRepo)
		}

	// 2. AlertService（带配置）
	alertSvc := NewAlertService(
		repoFactory.GetRuleRepository(),
		repoFactory.GetAlertRepository(),
		repoFactory.GetNotificationRepository(),
		repoFactory.GetWorkOrderRepository(),
		repoFactory.GetBlackBoxRepository(),
		repoFactory.GetTelemetryRepository(),
		repoFactory.GetDeviceRepository(),
		AlertServiceConfig{},
	)
	factory.alertService = alertSvc

	// 3. TelemetryService（依赖AlertService，使用具体类型）
	factory.telemetryService = NewTelemetryService(
		repoFactory.GetTelemetryRepository(),
		repoFactory.GetDeviceRepository(),
		repoFactory.GetAlertRepository(),
		repoFactory.GetWorkOrderRepository(),
		alertSvc, // 使用具体类型 *AlertService
	)

	// 4. 报告和导出服务
	factory.reportService = NewReportService(
		repoFactory.GetReportRepository(),
		repoFactory.GetTelemetryRepository(),
		repoFactory.GetDeviceRepository(),
		repoFactory.GetWorkOrderRepository(),
		repoFactory.GetNotificationRepository(),
	)
	factory.exportService = NewExportService(
		repoFactory.GetDeviceRepository(),
		nil,
		repoFactory.GetAlertRepository(),
		nil,
		nil,
	)

	// 5. 其他服务（暂未实现或需要额外依赖）
	// - agentService: 需要 CacheService 接口
	// - workOrderService: 需要独立初始化
	// - notificationService: 需要独立初始化
	// - blackBoxService: 需要独立初始化
	// - healthService: 需要 HTTP client 配置
	// 这些服务可通过 Set* 方法手动注入

	return factory
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
