package model

import (
	"time"
)

// Role represents a user role with permissions
type Role struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	IsSystem    bool      `json:"is_system" db:"is_system"` // System roles cannot be deleted
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a specific permission
type Permission struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RolePermission associates roles with permissions
type RolePermission struct {
	RoleID       int `json:"role_id" db:"role_id"`
	PermissionID int `json:"permission_id" db:"permission_id"`
}

// UserRole associates users with roles
type UserRole struct {
	UserID     int       `json:"user_id" db:"user_id"`
	RoleID     int       `json:"role_id" db:"role_id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=200"`
	TenantID    string `json:"tenant_id"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"max=200"`
}

// AssignRoleRequest represents the request body for assigning a role to a user
type AssignRoleRequest struct {
	RoleID   int    `json:"role_id" binding:"required"`
	TenantID string `json:"tenant_id"`
}

// AssignPermissionRequest represents the request body for assigning permission to a role
type AssignPermissionRequest struct {
	PermissionID int `json:"permission_id" binding:"required"`
}

// RoleResponse represents the API response for role data
type RoleResponse struct {
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Default system roles
var DefaultRoles = []Role{
	{ID: 1, Name: "admin", Description: "Administrator with full access", IsSystem: true},
	{ID: 2, Name: "operator", Description: "Device operator with management permissions", IsSystem: true},
	{ID: 3, Name: "viewer", Description: "Read-only access to devices and reports", IsSystem: true},
}

// Default permissions
var DefaultPermissions = []Permission{
	{ID: 1, Name: "devices.manage", Resource: "devices", Action: "manage", Description: "Full device management"},
	{ID: 2, Name: "devices.read", Resource: "devices", Action: "read", Description: "Read device data"},
	{ID: 3, Name: "users.manage", Resource: "users", Action: "manage", Description: "User management"},
	{ID: 4, Name: "users.read", Resource: "users", Action: "read", Description: "Read user data"},
	{ID: 5, Name: "rules.manage", Resource: "rules", Action: "manage", Description: "Rule management"},
	{ID: 6, Name: "rules.read", Resource: "rules", Action: "read", Description: "Read rules"},
	{ID: 7, Name: "reports.read", Resource: "reports", Action: "read", Description: "Read reports"},
	{ID: 8, Name: "reports.generate", Resource: "reports", Action: "generate", Description: "Generate reports"},
	{ID: 9, Name: "alerts.manage", Resource: "alerts", Action: "manage", Description: "Alert management"},
	{ID: 10, Name: "alerts.read", Resource: "alerts", Action: "read", Description: "Read alerts"},
	{ID: 11, Name: "workorders.manage", Resource: "workorders", Action: "manage", Description: "Work order management"},
	{ID: 12, Name: "workorders.read", Resource: "workorders", Action: "read", Description: "Read work orders"},
	{ID: 13, Name: "settings.manage", Resource: "settings", Action: "manage", Description: "Settings management"},
	{ID: 14, Name: "tenants.manage", Resource: "tenants", Action: "manage", Description: "Tenant management"},
}
