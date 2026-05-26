package handler

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/stretchr/testify/mock"
)

// MockDeviceRepository 模拟设备仓库 (共享 Mock)
type MockDeviceRepository struct {
	mock.Mock
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
	devices, _ := args.Get(0).([]model.Device)
	return devices, args.Get(1).(int), args.Error(2)
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
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

func (m *MockDeviceRepository) WithTx(tx database.TransactionInterface) repository.DeviceRepositoryInterface {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.DeviceRepositoryInterface)
}

// MockDeviceService 模拟设备服务 (共享 Mock)
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) AutoRegisterDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceService) Update(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceService) GetGraph(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDeviceService) GetByID(ctx context.Context, id string) (*model.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) List(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	args := m.Called(ctx, page, pageSize)
	devices, _ := args.Get(0).([]model.Device)
	return devices, args.Get(1).(int), args.Error(2)
}

func (m *MockDeviceService) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockUserService 模拟用户服务 (用于 Auth Handler)
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Authenticate(username, password string) (*model.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetByID(id int) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(id int, passwordHash string) error {
	args := m.Called(id, passwordHash)
	return args.Error(0)
}

func (m *MockUserService) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockUserService) UpdateTokenVersion(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockTelemetryService 模拟遥测服务
type MockTelemetryService struct {
	mock.Mock
}

func (m *MockTelemetryService) Ingest(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTelemetryService) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryService) GetSystemStatus(ctx context.Context) (*model.SystemStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemStatus), args.Error(1)
}

func (m *MockTelemetryService) GetLatestByDevice(ctx context.Context, deviceID string, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TelemetryData), args.Error(1)
}

func (m *MockTelemetryService) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, start, end, limit)
	data, _ := args.Get(0).([]model.TelemetryData)
	return data, args.Error(1)
}

func (m *MockTelemetryService) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	args := m.Called(ctx, deviceID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DeviceStats), args.Error(1)
}

func (m *MockTelemetryService) GetROIStats(ctx context.Context) (*model.ROIStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ROIStats), args.Error(1)
}

func (m *MockTelemetryService) GetHistoricalData(ctx context.Context, deviceID string, timeRange string, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, timeRange, limit)
	data, _ := args.Get(0).([]model.TelemetryData)
	return data, args.Error(1)
}

// MockAgentService 模拟 Agent 服务
type MockAgentService struct {
	mock.Mock
}

func (m *MockAgentService) Query(ctx context.Context, query model.AgentQuery) (*model.AgentResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AgentResponse), args.Error(1)
}

func (m *MockAgentService) GetDeviceContext(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAgentService) GetTaskLogs(ctx context.Context, limit int) ([]model.AgentTaskLog, error) {
	args := m.Called(ctx, limit)
	logs, _ := args.Get(0).([]model.AgentTaskLog)
	return logs, args.Error(1)
}

// MockTelemetryRepository 模拟遥测仓库
type MockTelemetryRepository struct {
	mock.Mock
}

func (m *MockTelemetryRepository) GetLatest(ctx context.Context) ([]model.TelemetryData, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]model.TelemetryData)
	return data, args.Error(1)
}

func (m *MockTelemetryRepository) GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error) {
	args := m.Called(ctx, deviceID, start, end, limit)
	data, _ := args.Get(0).([]model.TelemetryData)
	return data, args.Error(1)
}

func (m *MockTelemetryRepository) GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error) {
	args := m.Called(ctx, deviceID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DeviceStats), args.Error(1)
}

// MockTenantService 模拟租户服务 (用于 Tenant Handler)
// FIX-003: 更新方法签名添加 context.Context 参数
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) CreateTenant(ctx context.Context, name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	args := m.Called(ctx, name, slug, plan, maxDevices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenant(ctx context.Context, id string) (*model.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) ListTenants(ctx context.Context, limit, offset int) ([]model.Tenant, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Tenant), args.Error(1)
}

func (m *MockTenantService) UpdateTenant(ctx context.Context, id string, updates map[string]interface{}) (*model.Tenant, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) DeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantService) CountTenants(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

// MockRuleRepository 模拟规则仓库 (用于 Device Handler)
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id int) (*model.AlertRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRule), args.Error(1)
}

func (m *MockRuleRepository) List(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	rules, _ := args.Get(0).([]model.AlertRule)
	return rules, args.Error(1)
}

func (m *MockRuleRepository) ListEnabled(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	rules, _ := args.Get(0).([]model.AlertRule)
	return rules, args.Error(1)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuleRepository) ToggleEnabled(ctx context.Context, id int, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

// MockAlertService 模拟告警服务 (用于 Device Handler)
type MockAlertService struct {
	mock.Mock
}

func (m *MockAlertService) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertService) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertService) DeleteRule(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetRules(ctx context.Context) ([]model.AlertRule, error) {
	args := m.Called(ctx)
	rules, _ := args.Get(0).([]model.AlertRule)
	return rules, args.Error(1)
}

func (m *MockAlertService) EvaluateRules(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockAlertService) GetAlerts(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	alerts, _ := args.Get(0).([]model.Alert)
	return alerts, args.Get(1).(int), args.Error(2)
}

// P0-03: Mock method for GetAlertsWithFilter
func (m *MockAlertService) GetAlertsWithFilter(ctx context.Context, status, severity, deviceID string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, severity, deviceID, page, pageSize)
	alerts, _ := args.Get(0).([]model.Alert)
	return alerts, args.Get(1).(int), args.Error(2)
}

func (m *MockAlertService) InitializeDefaultRules(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAlertService) GetAlertByID(ctx context.Context, id int) (*model.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

func (m *MockAlertService) ResolveAlert(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) AcknowledgeAlert(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetRuleByID(ctx context.Context, id int) (*model.AlertRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRule), args.Error(1)
}

func (m *MockAlertService) ToggleRule(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertService) GetTrendReport(ctx context.Context, period string) (map[string]interface{}, error) {
	args := m.Called(ctx, period)
	data, _ := args.Get(0).(map[string]interface{})
	return data, args.Error(1)
}

func (m *MockAlertService) GetDeviceRanking(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	args := m.Called(ctx, limit)
	data, _ := args.Get(0).([]map[string]interface{})
	return data, args.Error(1)
}

func (m *MockAlertService) GetEfficiencyReport(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).(map[string]interface{})
	return data, args.Error(1)
}

// MockAlertRepository 模拟告警仓库
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	alerts, _ := args.Get(0).([]model.Alert)
	return alerts, args.Get(1).(int), args.Error(2)
}

// P0-03: Mock method for ListWithFilter
func (m *MockAlertRepository) ListWithFilter(ctx context.Context, filter repository.AlertFilter, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, filter, page, pageSize)
	alerts, _ := args.Get(0).([]model.Alert)
	return alerts, args.Get(1).(int), args.Error(2)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id int) (*model.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *model.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetRecentByDevice(ctx context.Context, deviceID string, ruleID, cooldownSec int) (*model.Alert, error) {
	args := m.Called(ctx, deviceID, ruleID, cooldownSec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Alert), args.Error(1)
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

// MockWorkOrderRepository 模拟工单仓库
type MockWorkOrderRepository struct {
	mock.Mock
}

func (m *MockWorkOrderRepository) Create(ctx context.Context, wo *model.WorkOrder) error {
	args := m.Called(ctx, wo)
	return args.Error(0)
}

func (m *MockWorkOrderRepository) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WorkOrder), args.Error(1)
}

func (m *MockWorkOrderRepository) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	args := m.Called(ctx, status, deviceID, page, pageSize)
	orders, _ := args.Get(0).([]model.WorkOrder)
	return orders, args.Get(1).(int), args.Error(2)
}

func (m *MockWorkOrderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockNotificationRepository 模拟通知仓库
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, n *model.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNotificationRepository) List(ctx context.Context, notifType string, unreadOnly bool, page, pageSize int) ([]model.Notification, int, error) {
	args := m.Called(ctx, notifType, unreadOnly, page, pageSize)
	notifications, _ := args.Get(0).([]model.Notification)
	return notifications, args.Get(1).(int), args.Error(2)
}

func (m *MockNotificationRepository) MarkRead(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockBlackBoxRepository 模拟黑匣子仓库
type MockBlackBoxRepository struct {
	mock.Mock
}

func (m *MockBlackBoxRepository) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockBlackBoxRepository) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	records, _ := args.Get(0).([]model.BlackBoxRecord)
	return records, args.Get(1).(int), args.Error(2)
}

// MockReportRepository 模拟报告仓库
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, report *model.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) List(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error) {
	args := m.Called(ctx, reportType, page, pageSize)
	reports, _ := args.Get(0).([]model.Report)
	return reports, args.Get(1).(int), args.Error(2)
}

// MockReportService 模拟报告服务
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GenerateReport(ctx context.Context, reportType string, deviceID string) (*model.Report, error) {
	args := m.Called(ctx, reportType, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportService) GetROIStats(ctx context.Context) (*model.ROIStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ROIStats), args.Error(1)
}

func (m *MockReportService) ListReports(ctx context.Context, reportType string, page, pageSize int) ([]model.Report, int, error) {
	args := m.Called(ctx, reportType, page, pageSize)
	reports, _ := args.Get(0).([]model.Report)
	return reports, args.Get(1).(int), args.Error(2)
}

func (m *MockReportService) GetReportByID(ctx context.Context, id int) (*model.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportService) DeleteReport(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockWorkOrderService 模拟工单服务
type MockWorkOrderService struct {
	mock.Mock
}

func (m *MockWorkOrderService) Create(ctx context.Context, order *model.WorkOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockWorkOrderService) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	args := m.Called(ctx, status, deviceID, page, pageSize)
	orders, _ := args.Get(0).([]model.WorkOrder)
	return orders, args.Get(1).(int), args.Error(2)
}

func (m *MockWorkOrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockWorkOrderService) GetByID(ctx context.Context, id int) (*model.WorkOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WorkOrder), args.Error(1)
}

// MockNotificationService 模拟通知服务
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) Create(ctx context.Context, n *model.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNotificationService) List(ctx context.Context, notifType string, page, pageSize int) ([]model.Notification, int, error) {
	args := m.Called(ctx, notifType, page, pageSize)
	notifications, _ := args.Get(0).([]model.Notification)
	return notifications, args.Get(1).(int), args.Error(2)
}

func (m *MockNotificationService) MarkRead(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockBlackBoxService 模拟黑匣子服务
type MockBlackBoxService struct {
	mock.Mock
}

func (m *MockBlackBoxService) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockBlackBoxService) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	records, _ := args.Get(0).([]model.BlackBoxRecord)
	return records, args.Get(1).(int), args.Error(2)
}

// MockExportService 模拟导出服务
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) Export(ctx context.Context, req *service.ExportRequest) (*service.ExportResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ExportResult), args.Error(1)
}

// MockUserRepository 模拟用户仓库
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
	users, _ := args.Get(0).([]model.User)
	return users, args.Get(1).(int), args.Error(2)
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
	return args.Get(0).(int), args.Error(1)
}

func (m *MockUserRepository) UpdateTokenVersion(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) WithTx(tx database.TransactionInterface) repository.UserRepositoryInterface {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.UserRepositoryInterface)
}

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.User), args.String(1), args.Error(2)
}

// FIX-016/017: 更新 ValidateToken 和 RefreshToken 方法签名以匹配 AuthServiceInterface
func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*service.Claims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPair), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, tokenString string) error {
	args := m.Called(ctx, tokenString)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// FIX-016/017: 新增 AuthServiceInterface 方法
func (m *MockAuthService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int), args.Error(2)
}
