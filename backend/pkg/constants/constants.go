// Package constants provides application-wide constants
// FIX-026: 魔法数字常量化
package constants

import "time"

// Pagination constants
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// Tenant constants
const (
	TenantPlanFree       = "free"
	TenantPlanBasic      = "basic"
	TenantPlanPro        = "pro"
	TenantPlanEnterprise = "enterprise"

	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusDeleted   = "deleted"
)

// Device constants
const (
	DeviceStatusOnline      = "online"
	DeviceStatusOffline     = "offline"
	DeviceStatusMaintenance = "maintenance"
	DeviceStatusError       = "error"
)

// User constants
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"

	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
	UserStatusPending  = "pending"
)

// Token constants
const (
	AccessTokenDurationHours  = 24
	RefreshTokenDurationDays  = 7
	RefreshTokenDurationHours = 168
	TokenMinSecretLength      = 32
)

// Password constants
const (
	PasswordMinLength = 8
	PasswordMaxLength = 72
)

// Timeout constants
const (
	DefaultRequestTimeout = 15 * time.Second
	MaxRequestTimeout     = 30 * time.Second
	DBQueryTimeout        = 10 * time.Second
)

// HTTP status codes (standardized)
const (
	HTTPStatusOK                  = 200
	HTTPStatusCreated             = 201
	HTTPStatusNoContent           = 204
	HTTPStatusBadRequest          = 400
	HTTPStatusUnauthorized        = 401
	HTTPStatusForbidden           = 403
	HTTPStatusNotFound            = 404
	HTTPStatusConflict            = 409
	HTTPStatusInternalServerError = 500
	HTTPStatusServiceUnavailable  = 503
)

// Bearer token prefix
const BearerTokenPrefix = "Bearer "

// Validation helper functions
func IsValidTenantPlan(plan string) bool {
	return plan == TenantPlanFree || plan == TenantPlanBasic ||
		plan == TenantPlanPro || plan == TenantPlanEnterprise
}

func IsValidTenantStatus(status string) bool {
	return status == TenantStatusActive || status == TenantStatusSuspended ||
		status == TenantStatusDeleted
}

func IsValidDeviceStatus(status string) bool {
	return status == DeviceStatusOnline || status == DeviceStatusOffline ||
		status == DeviceStatusMaintenance || status == DeviceStatusError
}

func IsValidUserRole(role string) bool {
	return role == RoleAdmin || role == RoleOperator || role == RoleViewer
}

func HasPermission(role, requiredRole string) bool {
	level := GetRoleLevel(role)
	requiredLevel := GetRoleLevel(requiredRole)
	return level >= requiredLevel
}

func GetRoleLevel(role string) int {
	switch role {
	case RoleAdmin:
		return 3
	case RoleOperator:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}
