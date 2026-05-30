package service

import (
	"database/sql"

	"github.com/industrial-ai/platform/pkg/cache"
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
	configService       ConfigServiceInterface
	repoFactory         *repository.RepositoryFactory
}

// NewServiceFactory 创建 Service 工厂（空工厂）
// 用于测试场景，通过 Set 方法注入 Mock
//
// 注意：此工厂返回空实例，所有 Service 字段为 nil。
// 调用方需要通过 Set* 方法手动注入 Service 实现，
// 或使用 NewServiceFactoryFromDB 创建完整工厂。
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// NewServiceFactoryFromDB 从数据库连接创建完整 Service 工厂
// Handler 层只需传入 *sql.DB，不需要了解 Repository 层细节
func NewServiceFactoryFromDB(db *sql.DB) *ServiceFactory {
	return NewServiceFactoryFromRepo(repository.NewRepositoryFactory(db))
}

// NewServiceFactoryFromRepo 创建 Service 工厂（从 Repository 创建）
// P1-09: 实现完整的Service初始化逻辑（简化版）
// 注意：部分服务因依赖复杂或返回值类型不匹配，保持nil，可通过Set方法注入
func NewServiceFactoryFromRepo(repoFactory *repository.RepositoryFactory) *ServiceFactory {
	if repoFactory == nil {
		return &ServiceFactory{}
	}

	factory := &ServiceFactory{repoFactory: repoFactory}

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
	reportSvc := NewReportService(
		repoFactory.GetReportRepository(),
		repoFactory.GetTelemetryRepository(),
		repoFactory.GetDeviceRepository(),
		repoFactory.GetWorkOrderRepository(),
		repoFactory.GetNotificationRepository(),
	)
	factory.reportService = reportSvc
	factory.exportService = NewExportService(
		repoFactory.GetDeviceRepository(),
		repoFactory.GetTelemetryRepository(),
		repoFactory.GetAlertRepository(),
		repoFactory.GetWorkOrderRepository(),
		reportSvc,
	)

	// 5. 工单、通知、黑匣子服务（仅依赖 Repository）
	factory.workOrderService = NewWorkOrderService(repoFactory.GetWorkOrderRepository())
	factory.notificationService = NewNotificationService(repoFactory.GetNotificationRepository())
	factory.blackBoxService = NewBlackBoxService(repoFactory.GetBlackBoxRepository())

	// healthService 需要 *sql.DB 和 HTTP client 配置，通过 SetHealthService 注入
	// agentService 需要 CacheService，通过 SetAgentService 注入

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

// InitializeAgentService 创建并注入 AgentService（需要 CacheService 外部依赖）
// AgentService 的创建需要在 Handler 层提供 CacheService 后调用
func (f *ServiceFactory) InitializeAgentService(cacheSvc cache.CacheService) {
	if f.repoFactory == nil || cacheSvc == nil {
		return
	}
	f.agentService = NewAgentService(
		f.repoFactory.GetAgentTaskLogRepository(),
		f.repoFactory.GetDeviceRepository(),
		f.repoFactory.GetTelemetryRepository(),
		cacheSvc,
	)
}

// InitializeConfigService 创建并注入 ConfigService（需在 AgentService 之后调用）
func (f *ServiceFactory) InitializeConfigService() {
	if f.repoFactory == nil {
		return
	}
	var agentSvc *AgentService
	if impl, ok := f.agentService.(*AgentService); ok {
		agentSvc = impl
	}
	f.configService = NewConfigService(
		f.repoFactory.GetSystemConfigRepo(),
		f.repoFactory.GetLLMConfigRepo(),
		agentSvc,
	)
}

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
func (f *ServiceFactory) GetConfigService() ConfigServiceInterface     { return f.configService }
func (f *ServiceFactory) SetConfigService(svc ConfigServiceInterface)   { f.configService = svc }

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
