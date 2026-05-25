package repository

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/mock"
)

// ============================================
// Mock Implementations - Phase 2
// ============================================

// MockDeviceRepository implements DeviceRepositoryInterface for testing
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepository) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	args := m.Called(ctx, page, pageSize)
	devices := args.Get(0).([]model.Device)
	total := args.Get(1).(int)
	return devices, total, args.Error(2)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDeviceRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

// MockUserRepository implements UserRepositoryInterface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	users := args.Get(0).([]model.User)
	total := args.Get(1).(int)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) UpdateTokenVersion(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) WithTx(tx database.TransactionInterface) UserRepositoryInterface {
	return m
}

// MockAlertRepository implements AlertRepositoryInterface for testing
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
}

// P0-03: Mock method for ListWithFilter
func (m *MockAlertRepository) ListWithFilter(ctx context.Context, filter AlertFilter, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, filter, page, pageSize)
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
}

func (m *MockAlertRepository) CountActive(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockAlertRepository) Resolve(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAlertRepository) GetRecentByDevice(ctx context.Context, deviceID string, ruleID int, cooldownSec int) (*model.Alert, error) {
	args := m.Called(ctx, deviceID, ruleID, cooldownSec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

// FIX-P1-01: N+1 查询优化 - 新增批量查询 mock 方法
func (m *MockAlertRepository) GetRecentAlertsByDeviceBatch(ctx context.Context, deviceID string, ruleIDs []int, cooldownSec int) (map[int]*model.Alert, error) {
	args := m.Called(ctx, deviceID, ruleIDs, cooldownSec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]*model.Alert), args.Error(1)
}

// P2-002: 告警历史管理 - 新增归档 mock 方法
func (m *MockAlertRepository) ArchiveOldAlerts(ctx context.Context, daysOld int) (int, error) {
	args := m.Called(ctx, daysOld)
	return args.Int(0), args.Error(1)
}

func (m *MockAlertRepository) GetArchivedAlerts(ctx context.Context, deviceID string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
}

func (m *MockAlertRepository) DeleteArchivedAlerts(ctx context.Context, daysOld int) (int, error) {
	args := m.Called(ctx, daysOld)
	return args.Int(0), args.Error(1)
}

func (m *MockAlertRepository) GetAlertStatistics(ctx context.Context) (*AlertStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AlertStatistics), args.Error(1)
}

// MockRuleRepository implements RuleRepositoryInterface for testing
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) List(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) ListEnabled(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) ToggleEnabled(ctx context.Context, id int, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id int) (*model.AlertRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTenantRepository implements TenantRepositoryInterface for testing
type MockTenantRepository struct {
	mock.Mock
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) List(ctx context.Context, limit, offset int) ([]model.Tenant, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FIX-003: 添加 context 参数
func (m *MockTenantRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

// MockRBACRepository implements RBACRepositoryInterface for testing
type MockRBACRepository struct {
	mock.Mock
}

func (m *MockRBACRepository) CreateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRBACRepository) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACRepository) ListRoles(ctx context.Context, tenantID string) ([]model.Role, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRBACRepository) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRBACRepository) AssignRoleToUser(ctx context.Context, userID, roleID int, tenantID string) error {
	args := m.Called(ctx, userID, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACRepository) GetUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACRepository) CreatePermission(ctx context.Context, perm *model.Permission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockRBACRepository) GetPermissionByID(ctx context.Context, id int) (*model.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockRBACRepository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRBACRepository) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRBACRepository) GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACRepository) InitializeDefaultRBAC(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockTelemetryRepository implements TelemetryRepositoryInterface
type MockTelemetryRepository struct {
	mock.Mock
}

func (m *MockTelemetryRepository) Insert(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTelemetryRepository) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryRepository) GetByDevice(ctx context.Context, deviceID string) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryRepository) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	args := m.Called(ctx, deviceID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DeviceStats), args.Error(1)
}

// MockWorkOrderRepository implements WorkOrderRepositoryInterface
type MockWorkOrderRepository struct {
	mock.Mock
}

func (m *MockWorkOrderRepository) Create(ctx context.Context, order *model.WorkOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockWorkOrderRepository) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	args := m.Called(ctx, status, deviceID, page, pageSize)
	return args.Get(0).([]model.WorkOrder), args.Get(1).(int), args.Error(2)
}

func (m *MockWorkOrderRepository) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WorkOrder), args.Error(1)
}

func (m *MockWorkOrderRepository) Update(ctx context.Context, order *model.WorkOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockWorkOrderRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockNotificationRepository implements NotificationRepositoryInterface
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) List(ctx context.Context, status string, unreadOnly bool, page, pageSize int) ([]model.Notification, int, error) {
	args := m.Called(ctx, status, unreadOnly, page, pageSize)
	return args.Get(0).([]model.Notification), args.Get(1).(int), args.Error(2)
}

func (m *MockNotificationRepository) MarkRead(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockBlackBoxRepository implements BlackBoxRepositoryInterface
type MockBlackBoxRepository struct {
	mock.Mock
}

func (m *MockBlackBoxRepository) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockBlackBoxRepository) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	return args.Get(0).([]model.BlackBoxRecord), args.Get(1).(int), args.Error(2)
}

func (m *MockBlackBoxRepository) GetByID(ctx context.Context, id int) (*model.BlackBoxRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.BlackBoxRecord), args.Error(1)
}

// MockReportRepository implements ReportRepositoryInterface
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, report *model.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) List(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error) {
	args := m.Called(ctx, reportType, page, pageSize)
	return args.Get(0).([]model.Report), args.Get(1).(int), args.Error(2)
}

func (m *MockReportRepository) GetByID(ctx context.Context, id int) (*model.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAgentTaskLogRepository implements AgentTaskLogRepositoryInterface
type MockAgentTaskLogRepository struct {
	mock.Mock
}

// NewMockAgentTaskLogRepository creates a new MockAgentTaskLogRepository
func NewMockAgentTaskLogRepository() *MockAgentTaskLogRepository {
	return &MockAgentTaskLogRepository{}
}

func (m *MockAgentTaskLogRepository) Create(ctx context.Context, log *model.AgentTaskLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAgentTaskLogRepository) List(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.AgentTaskLog), args.Error(1)
}

// ============================================
// Additional Mock Methods
// ============================================

func (m *MockDeviceRepository) WithTx(tx database.TransactionInterface) DeviceRepositoryInterface {
	return m
}

func (m *MockTelemetryRepository) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, start, end, limit)
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockWorkOrderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockWorkOrderRepository) WithTx(tx database.TransactionInterface) WorkOrderRepositoryInterface {
	return m
}

func (m *MockNotificationRepository) WithTx(tx database.TransactionInterface) NotificationRepositoryInterface {
	return m
}

func (m *MockBlackBoxRepository) WithTx(tx database.TransactionInterface) BlackBoxRepositoryInterface {
	return m
}

func (m *MockReportRepository) WithTx(tx database.TransactionInterface) ReportRepositoryInterface {
	return m
}

func (m *MockTelemetryRepository) WithTx(tx database.TransactionInterface) TelemetryRepositoryInterface {
	return m
}

func (m *MockAgentTaskLogRepository) WithTx(tx database.TransactionInterface) AgentTaskLogRepositoryInterface {
	return m
}
