package model

import (
	"time"
)

// Additional role-related types not in rbac.go

// RoleWithPermissions represents a role with its associated permissions (alias for RoleResponse)
type RoleWithPermissions = RoleResponse

// Default system role constants
const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
	RoleUser     = "user"
)

// Default permission constants
const (
	// Device permissions
	PermissionDeviceRead   = "device:read"
	PermissionDeviceWrite  = "device:write"
	PermissionDeviceDelete = "device:delete"

	// Alert permissions
	PermissionAlertRead   = "alert:read"
	PermissionAlertWrite  = "alert:write"
	PermissionAlertDelete = "alert:delete"

	// User permissions
	PermissionUserRead   = "user:read"
	PermissionUserWrite  = "user:write"
	PermissionUserDelete = "user:delete"

	// Role permissions
	PermissionRoleRead   = "role:read"
	PermissionRoleWrite  = "role:write"
	PermissionRoleDelete = "role:delete"

	// Tenant permissions
	PermissionTenantRead  = "tenant:read"
	PermissionTenantWrite = "tenant:write"

	// Report permissions
	PermissionReportRead  = "report:read"
	PermissionReportWrite = "report:write"

	// System permissions
	PermissionSystemAdmin = "system:admin"
)

// Resource types for permissions
const (
	ResourceDevice     = "device"
	ResourceAlert      = "alert"
	ResourceUser       = "user"
	ResourceRole       = "role"
	ResourceTenant     = "tenant"
	ResourceReport     = "report"
	ResourceSystem     = "system"
	ResourceTelemetry  = "telemetry"
	ResourceWorkOrder  = "work_order"
	ResourceDashboard  = "dashboard"
	ResourceAnnotation = "annotation"
	ResourceRule       = "rule"
	ResourceSettings   = "settings"
)

// Action types for permissions
const (
	ActionCreate   = "create"
	ActionRead     = "read"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionList     = "list"
	ActionManage   = "manage"
	ActionExecute  = "execute"
	ActionExport   = "export"
	ActionGenerate = "generate"
)

// PermissionCheckRequest represents a permission check request
type PermissionCheckRequest struct {
	UserID   int    `json:"user_id" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// PermissionCheckResponse represents the result of a permission check
type PermissionCheckResponse struct {
	Allowed     bool     `json:"allowed"`
	UserID      int      `json:"user_id"`
	Resource    string   `json:"resource"`
	Action      string   `json:"action"`
	Permissions []string `json:"permissions,omitempty"`
}

// ExtendedDefaultPermissions returns a comprehensive list of default permissions to seed
func ExtendedDefaultPermissions() []Permission {
	now := time.Now()
	return []Permission{
		// Device permissions
		{Name: "devices.list", Resource: ResourceDevice, Action: ActionList, Description: "List devices", CreatedAt: now},
		{Name: "devices.read", Resource: ResourceDevice, Action: ActionRead, Description: "View device details", CreatedAt: now},
		{Name: "devices.create", Resource: ResourceDevice, Action: ActionCreate, Description: "Create devices", CreatedAt: now},
		{Name: "devices.update", Resource: ResourceDevice, Action: ActionUpdate, Description: "Update devices", CreatedAt: now},
		{Name: "devices.delete", Resource: ResourceDevice, Action: ActionDelete, Description: "Delete devices", CreatedAt: now},

		// Alert permissions
		{Name: "alerts.list", Resource: ResourceAlert, Action: ActionList, Description: "List alerts", CreatedAt: now},
		{Name: "alerts.read", Resource: ResourceAlert, Action: ActionRead, Description: "View alert details", CreatedAt: now},
		{Name: "alerts.create", Resource: ResourceAlert, Action: ActionCreate, Description: "Create alerts", CreatedAt: now},
		{Name: "alerts.update", Resource: ResourceAlert, Action: ActionUpdate, Description: "Update alerts", CreatedAt: now},
		{Name: "alerts.delete", Resource: ResourceAlert, Action: ActionDelete, Description: "Delete alerts", CreatedAt: now},

		// User permissions
		{Name: "users.list", Resource: ResourceUser, Action: ActionList, Description: "List users", CreatedAt: now},
		{Name: "users.read", Resource: ResourceUser, Action: ActionRead, Description: "View user details", CreatedAt: now},
		{Name: "users.create", Resource: ResourceUser, Action: ActionCreate, Description: "Create users", CreatedAt: now},
		{Name: "users.update", Resource: ResourceUser, Action: ActionUpdate, Description: "Update users", CreatedAt: now},
		{Name: "users.delete", Resource: ResourceUser, Action: ActionDelete, Description: "Delete users", CreatedAt: now},

		// Role permissions
		{Name: "roles.list", Resource: ResourceRole, Action: ActionList, Description: "List roles", CreatedAt: now},
		{Name: "roles.read", Resource: ResourceRole, Action: ActionRead, Description: "View role details", CreatedAt: now},
		{Name: "roles.create", Resource: ResourceRole, Action: ActionCreate, Description: "Create roles", CreatedAt: now},
		{Name: "roles.update", Resource: ResourceRole, Action: ActionUpdate, Description: "Update roles", CreatedAt: now},
		{Name: "roles.delete", Resource: ResourceRole, Action: ActionDelete, Description: "Delete roles", CreatedAt: now},

		// Tenant permissions
		{Name: "tenants.read", Resource: ResourceTenant, Action: ActionRead, Description: "View tenant details", CreatedAt: now},
		{Name: "tenants.update", Resource: ResourceTenant, Action: ActionUpdate, Description: "Update tenant settings", CreatedAt: now},

		// Report permissions
		{Name: "reports.list", Resource: ResourceReport, Action: ActionList, Description: "List reports", CreatedAt: now},
		{Name: "reports.read", Resource: ResourceReport, Action: ActionRead, Description: "View reports", CreatedAt: now},
		{Name: "reports.create", Resource: ResourceReport, Action: ActionCreate, Description: "Create reports", CreatedAt: now},
		{Name: "reports.export", Resource: ResourceReport, Action: ActionExport, Description: "Export reports", CreatedAt: now},

		// Telemetry permissions
		{Name: "telemetry.read", Resource: ResourceTelemetry, Action: ActionRead, Description: "View telemetry data", CreatedAt: now},
		{Name: "telemetry.export", Resource: ResourceTelemetry, Action: ActionExport, Description: "Export telemetry data", CreatedAt: now},

		// Work order permissions
		{Name: "work_orders.list", Resource: ResourceWorkOrder, Action: ActionList, Description: "List work orders", CreatedAt: now},
		{Name: "work_orders.read", Resource: ResourceWorkOrder, Action: ActionRead, Description: "View work order details", CreatedAt: now},
		{Name: "work_orders.create", Resource: ResourceWorkOrder, Action: ActionCreate, Description: "Create work orders", CreatedAt: now},
		{Name: "work_orders.update", Resource: ResourceWorkOrder, Action: ActionUpdate, Description: "Update work orders", CreatedAt: now},
		{Name: "work_orders.delete", Resource: ResourceWorkOrder, Action: ActionDelete, Description: "Delete work orders", CreatedAt: now},

		// Dashboard permissions
		{Name: "dashboards.read", Resource: ResourceDashboard, Action: ActionRead, Description: "View dashboards", CreatedAt: now},
		{Name: "dashboards.update", Resource: ResourceDashboard, Action: ActionUpdate, Description: "Update dashboards", CreatedAt: now},

		// Rule permissions
		{Name: "rules.manage", Resource: ResourceRule, Action: ActionManage, Description: "Manage alert rules", CreatedAt: now},
		{Name: "rules.read", Resource: ResourceRule, Action: ActionRead, Description: "Read alert rules", CreatedAt: now},

		// Settings permissions
		{Name: "settings.manage", Resource: ResourceSettings, Action: ActionManage, Description: "Manage settings", CreatedAt: now},

		// System permissions
		{Name: "system.admin", Resource: ResourceSystem, Action: ActionManage, Description: "Full system administration", CreatedAt: now},
	}
}
