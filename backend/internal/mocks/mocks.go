package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// Mock DeviceService
// ============================================

type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) Create(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.Device), args.Get(1).(int), args.Error(2)
}

func (m *MockDeviceService) Update(ctx context.Context, device *model.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceService) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockDeviceService) AutoRegisterDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) GetGraph(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// ============================================
// Mock AuthService
// ============================================

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Get(1).(string), args.Error(2)
	}
	return args.Get(0).(*model.User), args.Get(1).(string), args.Error(2)
}

func (m *MockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(string), args.Error(2)
	}
	return args.Get(0).(*model.User), args.Get(1).(string), args.Error(2)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// ============================================
// Mock AlertService
// ============================================

type MockAlertService struct {
	mock.Mock
}

func (m *MockAlertService) EvaluateRules(ctx context.Context, data *model.TelemetryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AlertRule), args.Error(1)
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

func (m *MockAlertService) GetAlerts(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	alerts := args.Get(0).([]model.Alert)
	total := args.Get(1).(int)
	return alerts, total, args.Error(2)
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

func (m *MockAlertService) InitializeDefaultRules(ctx context.Context) error {
	args := m.Called(ctx)
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

// ============================================
// Mock UserService
// ============================================

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

// ============================================
// Mock HealthService
// ============================================

type MockHealthService struct {
	mock.Mock
}

func (m *MockHealthService) CheckHealth(ctx context.Context) *service.HealthCheckResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*service.HealthCheckResponse)
}

// ============================================
// Mock AgentService
// ============================================

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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AgentTaskLog), args.Error(1)
}

// ============================================
// Mock WorkOrderService
// ============================================

type MockWorkOrderService struct {
	mock.Mock
}

func (m *MockWorkOrderService) Create(ctx context.Context, order *model.WorkOrder) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockWorkOrderService) List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error) {
	args := m.Called(ctx, status, deviceID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.WorkOrder), args.Get(1).(int), args.Error(2)
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

// ============================================
// Mock NotificationService
// ============================================

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) Create(ctx context.Context, n *model.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNotificationService) List(ctx context.Context, notifType string, page, pageSize int) ([]model.Notification, int, error) {
	args := m.Called(ctx, notifType, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.Notification), args.Get(1).(int), args.Error(2)
}

func (m *MockNotificationService) MarkRead(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ============================================
// Mock BlackBoxService
// ============================================

type MockBlackBoxService struct {
	mock.Mock
}

func (m *MockBlackBoxService) Create(ctx context.Context, record *model.BlackBoxRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockBlackBoxService) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error) {
	args := m.Called(ctx, deviceID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.BlackBoxRecord), args.Get(1).(int), args.Error(2)
}

// ============================================
// Mock ReportService
// ============================================

type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) Generate(ctx context.Context, reportType string, params map[string]interface{}) (*model.Report, error) {
	args := m.Called(ctx, reportType, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportService) List(ctx context.Context, page, pageSize int) ([]model.Report, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.Report), args.Get(1).(int), args.Error(2)
}

func (m *MockReportService) GetByID(ctx context.Context, id int) (*model.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

// ============================================
// Mock ExportService
// ============================================

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

// ============================================
// Mock RBACService
// ============================================

type MockRBACService struct {
	mock.Mock
}

func (m *MockRBACService) CreateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACService) UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACService) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRBACService) GetRoleByID(ctx context.Context, id int) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRBACService) ListRoles(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACService) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRBACService) ListUserRoles(ctx context.Context, userID int) ([]model.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRBACService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRBACService) AssignPermissionToRole(ctx context.Context, roleID, permID int) error {
	args := m.Called(ctx, roleID, permID)
	return args.Error(0)
}

func (m *MockRBACService) RemovePermissionFromRole(ctx context.Context, roleID, permID int) error {
	args := m.Called(ctx, roleID, permID)
	return args.Error(0)
}

// ============================================
// Mock TenantService
// ============================================

type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) CreateTenant(name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	args := m.Called(name, slug, plan, maxDevices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) UpdateTenant(id, name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	args := m.Called(id, name, slug, plan, maxDevices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenantByID(id string) (*model.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantService) ListTenants(page, pageSize int) ([]model.Tenant, int, error) {
	args := m.Called(page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]model.Tenant), args.Get(1).(int), args.Error(2)
}

func (m *MockTenantService) DeleteTenant(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ============================================
// Mock JWTService
// ============================================

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(user *model.User) (string, error) {
	args := m.Called(user)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockJWTService) ParseToken(token string) (*service.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

func (m *MockJWTService) RefreshToken(token string) (string, error) {
	args := m.Called(token)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockJWTService) ValidateToken(token string) (*service.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}