package service

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// 统一 Mock 实现 (使用 testify/mock)
// ============================================

// MockDeviceService 设备服务 Mock
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

// MockAuthService 认证服务 Mock
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

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockAuthService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int), args.Error(2)
}

// MockUserService 用户服务 Mock
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

// MockHealthService 健康检查服务 Mock
type MockHealthService struct {
	mock.Mock
}

func (m *MockHealthService) CheckHealth(ctx context.Context) *HealthCheckResponse {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*HealthCheckResponse)
}

// MockWorkOrderService 工单服务 Mock
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

// MockNotificationService 通知服务 Mock
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

// MockBlackBoxService 黑匣子服务 Mock
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

// MockReportService 报告服务 Mock
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

// MockExportService 导出服务 Mock
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) Export(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExportResult), args.Error(1)
}

// MockRBACService RBAC 服务 Mock
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

// MockTenantService 租户服务 Mock
type MockTenantService struct {
	mock.Mock
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) CreateTenant(ctx context.Context, name, slug, plan string, maxDevices int) (*model.Tenant, error) {
	args := m.Called(ctx, name, slug, plan, maxDevices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) GetTenant(ctx context.Context, id string) (*model.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) ListTenants(ctx context.Context, limit, offset int) ([]model.Tenant, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数，使用 updates map
func (m *MockTenantService) UpdateTenant(ctx context.Context, id string, updates map[string]interface{}) (*model.Tenant, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) DeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FIX-003: 添加 context 参数
func (m *MockTenantService) CountTenants(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}
