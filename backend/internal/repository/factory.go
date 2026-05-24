package repository

import (
	"database/sql"
	"errors"

	"github.com/industrial-ai/platform/pkg/database"
)

// ErrDatabaseNotInitialized 数据库未初始化错误
// FIX-001: 添加错误定义，防止 nil 返回导致的 panic
var ErrDatabaseNotInitialized = errors.New("database not initialized")

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

// GetTenantRepo 获取租户 Repository
// FIX-001: 返回具体类型而非 interface{}，添加错误检查防止 panic
func (f *RepositoryFactory) GetTenantRepo() (*TenantRepo, error) {
	if f.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	return NewTenantRepo(f.db), nil
}

// GetRoleRepo 获取角色 Repository
// FIX-001: 返回具体类型而非 interface{}，添加错误检查防止 panic
func (f *RepositoryFactory) GetRoleRepo() (*RoleRepo, error) {
	if f.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	return NewRoleRepo(f.db), nil
}

// GetPermissionRepo 获取权限 Repository
// FIX-001: 返回具体类型而非 interface{}，添加错误检查防止 panic
func (f *RepositoryFactory) GetPermissionRepo() (*PermissionRepo, error) {
	if f.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	return NewPermissionRepo(f.db), nil
}

// GetRBACRepository 获取 RBAC Repository（综合角色、权限、用户角色管理）
// FIX-001: 返回具体类型而非 interface{}，添加错误检查防止 panic
func (f *RepositoryFactory) GetRBACRepository() (*RBACRepository, error) {
	if f.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	return NewRBACRepository(f.db), nil
}
