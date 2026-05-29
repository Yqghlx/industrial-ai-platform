package service

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// FIX-038: Service 接口定义 (便于 Mock 测试)
// ============================================

// DeviceServiceInterface 设备服务接口
// 用于 Mock 测试和依赖注入
type DeviceServiceInterface interface {
	// Create 创建新设备
	Create(ctx context.Context, device *model.Device) error

	// GetByID 根据 ID 获取设备
	GetByID(ctx context.Context, id string) (*model.Device, error)

	// List 获取设备列表 (分页)
	List(ctx context.Context, page, pageSize int) ([]model.Device, int, error)

	// Update 更新设备
	Update(ctx context.Context, device *model.Device) error

	// Delete 删除设备
	Delete(ctx context.Context, id string) error

	// UpdateStatus 更新设备状态
	UpdateStatus(ctx context.Context, id, status string) error

	// AutoRegisterDevice 自动注册设备
	AutoRegisterDevice(ctx context.Context, deviceID string) (*model.Device, error)

	// GetGraph 获取设备关系图
	GetGraph(ctx context.Context) (*model.DeviceGraph, error)

	// GetDeviceStats 获取设备统计数据
	GetDeviceStats(ctx context.Context, deviceID string) (*model.DeviceStatsDetail, error)
}

// AuthServiceInterface 认证服务接口
// 用于 Mock 测试和依赖注入
// SEC-HIGH-03: 添加 DeleteUser 方法
type AuthServiceInterface interface {
	// Login 用户登录
	Login(ctx context.Context, username, password string) (*model.User, string, error)

	// Register 用户注册
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error)

	// GetUserByID 根据 ID 获取用户
	GetUserByID(ctx context.Context, id int) (*model.User, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error

	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, token string) (*Claims, error)

	// ListUsers 列出所有用户
	ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error)

	// DeleteUser 删除用户
	// SEC-HIGH-03: 新增删除用户方法
	DeleteUser(ctx context.Context, userID int) error

	// EnsureDefaultAdmin 确保默认管理员存在
	EnsureDefaultAdmin(ctx context.Context, password string) error
}

// UserServiceInterface 用户服务接口
// 用于 Mock 测试和依赖注入
type UserServiceInterface interface {
	// Authenticate 验证用户登录
	Authenticate(username, password string) (*model.User, error)

	// GetByID 根据 ID 获取用户
	GetByID(id int) (*model.User, error)

	// UpdatePassword 更新用户密码
	UpdatePassword(id int, passwordHash string) error

	// GetTokenVersion 获取用户的 Token 版本号
	GetTokenVersion(ctx context.Context, userID int) (int, error)

	// UpdateTokenVersion 递增用户的 Token 版本号
	UpdateTokenVersion(ctx context.Context, userID int) error
}

// HealthServiceInterface 健康检查服务接口
type HealthServiceInterface interface {
	CheckHealth(ctx context.Context) *HealthCheckResponse
}

// AlertServiceInterface 告警服务组合接口
// 包含规则管理、告警管理和报告三个细分接口
type AlertServiceInterface interface {
	AlertRuleServiceInterface
	AlertManagementServiceInterface
	AlertReportServiceInterface
}

// AlertRuleServiceInterface 告警规则管理接口
type AlertRuleServiceInterface interface {
	CreateRule(ctx context.Context, rule *model.AlertRule) error
	UpdateRule(ctx context.Context, rule *model.AlertRule) error
	DeleteRule(ctx context.Context, id int) error
	GetRules(ctx context.Context) ([]model.AlertRule, error)
	GetRuleByID(ctx context.Context, id int) (*model.AlertRule, error)
	ToggleRule(ctx context.Context, id int) error
	InitializeDefaultRules(ctx context.Context) error
}

// AlertManagementServiceInterface 告警管理接口
type AlertManagementServiceInterface interface {
	EvaluateRules(ctx context.Context, data *model.TelemetryData) error
	GetAlerts(ctx context.Context, status string, page, pageSize int) ([]model.Alert, int, error)
	GetAlertsWithFilter(ctx context.Context, status, severity, deviceID string, page, pageSize int) ([]model.Alert, int, error)
	GetAlertByID(ctx context.Context, id int) (*model.Alert, error)
	ResolveAlert(ctx context.Context, id int) error
	AcknowledgeAlert(ctx context.Context, id int) error
}

// AlertReportServiceInterface 告警报告接口
type AlertReportServiceInterface interface {
	GetTrendReport(ctx context.Context, period string) (*model.TrendReport, error)
	GetDeviceRanking(ctx context.Context, limit int) ([]model.DeviceRankingEntry, error)
	GetEfficiencyReport(ctx context.Context) (*model.EfficiencyReport, error)
}

// AgentServiceInterface Agent 服务接口
type AgentServiceInterface interface {
	Query(ctx context.Context, query model.AgentQuery) (*model.AgentResponse, error)
	GetDeviceContext(ctx context.Context, deviceID string) (map[string]interface{}, error)
	GetTaskLogs(ctx context.Context, limit int) ([]model.AgentTaskLog, error)
}

// TelemetryServiceInterface 遥测服务接口
type TelemetryServiceInterface interface {
	Ingest(ctx context.Context, data *model.TelemetryData) error
	GetLatest(ctx context.Context) ([]model.TelemetryData, error)
	GetLatestByDevice(ctx context.Context, deviceID string, limit int) ([]model.TelemetryData, error)
	GetByDeviceID(ctx context.Context, deviceID string, start, end time.Time, limit int) ([]model.TelemetryData, error)
	GetStats(ctx context.Context, deviceID string, start, end time.Time) (*model.DeviceStats, error)
	GetROIStats(ctx context.Context) (*model.ROIStats, error)
	GetSystemStatus(ctx context.Context) (*model.SystemStatus, error)
	GetHistoricalData(ctx context.Context, deviceID string, timeRange string, limit int) ([]model.TelemetryData, error)
}

// ============================================
// 接口实现验证 (编译时检查)
// ============================================

// 确保 DeviceService 实现 DeviceServiceInterface
var _ DeviceServiceInterface = (*DeviceService)(nil)

// 确保 AuthService 实现 AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// 确保 UserService 实现 UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

// 确保 HealthService 实现 HealthServiceInterface
var _ HealthServiceInterface = (*HealthService)(nil)

// 确保 TelemetryService 实现 TelemetryServiceInterface
var _ TelemetryServiceInterface = (*TelemetryService)(nil)

// WorkOrderServiceInterface 工单服务接口
type WorkOrderServiceInterface interface {
	Create(ctx context.Context, order *model.WorkOrder) error
	List(ctx context.Context, status, deviceID string, page, pageSize int) ([]model.WorkOrder, int, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	GetByID(ctx context.Context, id int) (*model.WorkOrder, error)
}

// NotificationServiceInterface 通知服务接口
type NotificationServiceInterface interface {
	Create(ctx context.Context, n *model.Notification) error
	List(ctx context.Context, notifType string, page, pageSize int) ([]model.Notification, int, error)
	MarkRead(ctx context.Context, id int) error
}

// BlackBoxServiceInterface 黑匣子服务接口
type BlackBoxServiceInterface interface {
	Create(ctx context.Context, record *model.BlackBoxRecord) error
	List(ctx context.Context, deviceID string, page, pageSize int) ([]model.BlackBoxRecord, int, error)
	// GetRecordByID 根据 ID 获取单条黑匣子记录
	GetRecordByID(ctx context.Context, id int64) (*model.BlackBoxRecord, error)
}

// TenantServiceInterface 租户服务接口
// FIX-003: 添加 context 支持
type TenantServiceInterface interface {
	CreateTenant(ctx context.Context, name, slug, plan string, maxDevices int) (*model.Tenant, error)
	GetTenant(ctx context.Context, id string) (*model.Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error)
	ListTenants(ctx context.Context, limit, offset int) ([]model.Tenant, error)
	UpdateTenant(ctx context.Context, id string, updates map[string]interface{}) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id string) error
	CountTenants(ctx context.Context) (int, error)
}

// RBACServiceInterface RBAC服务接口
// 组合接口，包含所有细分接口。RBACService 实现此完整接口。
// 消费者应依赖所需的细分接口而非此臃肿接口（ISP 原则）。
type RBACServiceInterface interface {
	RoleServiceInterface
	PermissionServiceInterface
	RolePermissionServiceInterface
	UserRoleServiceInterface
}

// RoleServiceInterface 角色管理接口
type RoleServiceInterface interface {
	CreateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) (*model.Role, error)
	DeleteRole(ctx context.Context, id int) error
	GetRoleByID(ctx context.Context, id int) (*model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
}

// PermissionServiceInterface 权限管理接口
type PermissionServiceInterface interface {
	CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error)
	GetPermission(ctx context.Context, id int) (*model.Permission, error)
	DeletePermission(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]model.Permission, error)
}

// RolePermissionServiceInterface 角色权限关联接口
type RolePermissionServiceInterface interface {
	AssignPermissionToRole(ctx context.Context, roleID, permID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permID int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error)
}

// UserRoleServiceInterface 用户角色管理接口
type UserRoleServiceInterface interface {
	AssignRoleToUser(ctx context.Context, userID, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	ListUserRoles(ctx context.Context, userID int) ([]model.Role, error)
	GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error)
	CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error)
}

// 确保 RBACService 实现 RBACServiceInterface
var _ RBACServiceInterface = (*RBACService)(nil)
