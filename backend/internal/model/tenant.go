package model

import (
	"time"
)

// TenantPlan defines the subscription plan types
type TenantPlan string

const (
	PlanFree       TenantPlan = "free"
	PlanPro        TenantPlan = "pro"
	PlanEnterprise TenantPlan = "enterprise"
)

// PlanLimits defines the limits for each plan
var PlanLimits = map[TenantPlan]struct {
	MaxDevices int
	MaxUsers   int
	MaxAlerts  int
}{
	PlanFree:       {MaxDevices: 10, MaxUsers: 3, MaxAlerts: 20},
	PlanPro:        {MaxDevices: 100, MaxUsers: 20, MaxAlerts: 200},
	PlanEnterprise: {MaxDevices: -1, MaxUsers: -1, MaxAlerts: -1}, // -1 means unlimited
}

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Slug       string    `json:"slug" db:"slug"`
	Plan       string    `json:"plan" db:"plan"` // free, pro, enterprise
	MaxDevices int       `json:"max_devices" db:"max_devices"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	Settings   string    `json:"settings" db:"settings"` // JSON string for tenant-specific settings
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// TenantSettings represents the JSON structure for tenant settings
type TenantSettings struct {
	EmailNotifications bool            `json:"email_notifications"`
	SMSAlerts          bool            `json:"sms_alerts"`
	WebhookURL         string          `json:"webhook_url"`
	CustomBranding     *CustomBranding `json:"custom_branding,omitempty"`
	Features           map[string]bool `json:"features,omitempty"`
}

// CustomBranding represents tenant-specific branding settings
type CustomBranding struct {
	LogoURL      string `json:"logo_url,omitempty"`
	PrimaryColor string `json:"primary_color,omitempty"`
	CompanyName  string `json:"company_name,omitempty"`
}

// TenantUsage tracks the current usage for a tenant
type TenantUsage struct {
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	DeviceCount    int       `json:"device_count" db:"device_count"`
	UserCount      int       `json:"user_count" db:"user_count"`
	AlertCount     int       `json:"alert_count" db:"alert_count"`
	DataPointsDay  int64     `json:"data_points_day" db:"data_points_day"` // data points in last 24h
	LastCalculated time.Time `json:"last_calculated" db:"last_calculated"`
}

// CreateTenantRequest represents the request body for creating a tenant
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Slug string `json:"slug" binding:"required,min=2,max=50"` // URL-friendly identifier (removed 'slug' custom validator)
	Plan string `json:"plan" binding:"omitempty,oneof=free pro enterprise"`
}

// TenantCreateRequest alias for CreateTenantRequest
type TenantCreateRequest = CreateTenantRequest

// UpdateTenantRequest represents the request body for updating a tenant
type UpdateTenantRequest struct {
	Name       *string `json:"name" binding:"omitempty,min=2,max=100"`
	Plan       *string `json:"plan" binding:"omitempty,oneof=free pro enterprise"`
	MaxDevices *int    `json:"max_devices" binding:"omitempty,min=0"`
	IsActive   *bool   `json:"is_active"`
	Settings   *string `json:"settings" binding:"omitempty"` // JSON string
}

// TenantUpdateRequest alias for UpdateTenantRequest (simplified version for handler)
type TenantUpdateRequest struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Plan       string `json:"plan"`
	MaxDevices int    `json:"max_devices"`
}

// TenantResponse represents the API response for tenant data
type TenantResponse struct {
	Tenant Tenant      `json:"tenant"`
	Usage  TenantUsage `json:"usage"`
}

// GetMaxDevices returns the device limit for the tenant
func (t *Tenant) GetMaxDevices() int {
	if t.MaxDevices > 0 {
		return t.MaxDevices
	}
	if limits, ok := PlanLimits[TenantPlan(t.Plan)]; ok {
		return limits.MaxDevices
	}
	return 10 // Default fallback
}

// CanAddDevice checks if the tenant can add more devices
func (t *Tenant) CanAddDevice(currentCount int) bool {
	maxDevices := t.GetMaxDevices()
	return maxDevices < 0 || currentCount < maxDevices // -1 means unlimited
}

// GetPlanLimits returns the plan limits for this tenant
func (t *Tenant) GetPlanLimits() (maxDevices, maxUsers, maxAlerts int) {
	if limits, ok := PlanLimits[TenantPlan(t.Plan)]; ok {
		return limits.MaxDevices, limits.MaxUsers, limits.MaxAlerts
	}
	return 10, 3, 20 // Default to free plan limits
}

// TenantUpdates 租户更新参数（指针字段表示可选更新）
type TenantUpdates struct {
	Name       *string `json:"name,omitempty"`
	Slug       *string `json:"slug,omitempty"`
	Plan       *string `json:"plan,omitempty"`
	MaxDevices *int    `json:"max_devices,omitempty"`
}
