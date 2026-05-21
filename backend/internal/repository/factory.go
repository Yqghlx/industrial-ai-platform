package repository

import (
	"database/sql"

	"github.com/industrial-ai/platform/pkg/database"
)

// ============================================
// Repository 工厂 (简化版 - 用于测试依赖注入)
// ============================================

// RepositoryFactory Repository 工厂
// 简化版：主要用于测试场景，通过 Set 方法注入 Mock
type RepositoryFactory struct {
	db database.DatabaseInterface
}

// NewRepositoryFactory 创建 Repository 工厂
func NewRepositoryFactory(db *sql.DB) *RepositoryFactory {
	return &RepositoryFactory{db: database.NewDBWrapper(db)}
}

// GetDeviceRepository 获取设备 Repository
func (f *RepositoryFactory) GetDeviceRepository() *DeviceRepository {
	return NewDeviceRepository(f.db)
}

// GetTelemetryRepository 获取遥测 Repository
func (f *RepositoryFactory) GetTelemetryRepository() *TelemetryRepository {
	return NewTelemetryRepository(f.db)
}

// GetAlertRepository 获取告警 Repository
func (f *RepositoryFactory) GetAlertRepository() *AlertRepository {
	return NewAlertRepository(f.db)
}

// GetRuleRepository 获取规则 Repository
func (f *RepositoryFactory) GetRuleRepository() *RuleRepository {
	return NewRuleRepository(f.db)
}

// GetUserRepository 获取用户 Repository
func (f *RepositoryFactory) GetUserRepository() *UserRepository {
	return NewUserRepository(f.db)
}

// GetWorkOrderRepository 获取工单 Repository
func (f *RepositoryFactory) GetWorkOrderRepository() *WorkOrderRepository {
	return NewWorkOrderRepository(f.db)
}

// GetNotificationRepository 获取通知 Repository
func (f *RepositoryFactory) GetNotificationRepository() *NotificationRepository {
	return NewNotificationRepository(f.db)
}

// GetBlackBoxRepository 获取黑匣子 Repository
func (f *RepositoryFactory) GetBlackBoxRepository() *BlackBoxRepository {
	return NewBlackBoxRepository(f.db)
}

// GetReportRepository 获取报告 Repository
func (f *RepositoryFactory) GetReportRepository() *ReportRepository {
	return NewReportRepository(f.db)
}

// GetTenantRepository 获取租户 Repository
// TODO: 实现后返回具体类型
func (f *RepositoryFactory) GetTenantRepository() interface{} {
	return nil
}

// GetRoleRepository 获取角色 Repository
// TODO: 实现后返回具体类型
func (f *RepositoryFactory) GetRoleRepository() interface{} {
	return nil
}

// GetPermissionRepository 获取权限 Repository
// TODO: 实现后返回具体类型
func (f *RepositoryFactory) GetPermissionRepository() interface{} {
	return nil
}

// GetUserRoleRepository 获取用户角色 Repository
// TODO: 实现后返回具体类型
func (f *RepositoryFactory) GetUserRoleRepository() interface{} {
	return nil
}
