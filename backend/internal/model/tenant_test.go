package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTenant_GetMaxDevices tests the GetMaxDevices method
func TestTenant_GetMaxDevices(t *testing.T) {
	tests := []struct {
		name     string
		tenant   Tenant
		expected int
	}{
		{
			name: "custom max devices",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 50,
			},
			expected: 50,
		},
		{
			name: "free plan default",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 0,
			},
			expected: 10, // PlanLimits[PlanFree].MaxDevices
		},
		{
			name: "pro plan default",
			tenant: Tenant{
				Plan:       "pro",
				MaxDevices: 0,
			},
			expected: 100, // PlanLimits[PlanPro].MaxDevices
		},
		{
			name: "enterprise plan default",
			tenant: Tenant{
				Plan:       "enterprise",
				MaxDevices: 0,
			},
			expected: -1, // unlimited
		},
		{
			name: "unknown plan fallback",
			tenant: Tenant{
				Plan:       "unknown",
				MaxDevices: 0,
			},
			expected: 10, // default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tenant.GetMaxDevices()
			if result != tt.expected {
				t.Errorf("GetMaxDevices() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestTenant_CanAddDevice tests the CanAddDevice method
func TestTenant_CanAddDevice(t *testing.T) {
	tests := []struct {
		name         string
		tenant       Tenant
		currentCount int
		expected     bool
	}{
		{
			name: "enterprise unlimited",
			tenant: Tenant{
				Plan:       "enterprise",
				MaxDevices: 0,
			},
			currentCount: 9999,
			expected:     true, // -1 means unlimited
		},
		{
			name: "free plan under limit",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 0,
			},
			currentCount: 5,
			expected:     true, // 5 < 10
		},
		{
			name: "free plan at limit",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 0,
			},
			currentCount: 10,
			expected:     false, // 10 >= 10
		},
		{
			name: "custom limit under",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 5,
			},
			currentCount: 3,
			expected:     true, // 3 < 5
		},
		{
			name: "custom limit at",
			tenant: Tenant{
				Plan:       "free",
				MaxDevices: 5,
			},
			currentCount: 5,
			expected:     false, // 5 >= 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tenant.CanAddDevice(tt.currentCount)
			if result != tt.expected {
				t.Errorf("CanAddDevice(%d) = %v, want %v", tt.currentCount, result, tt.expected)
			}
		})
	}
}

// TestTenant_GetPlanLimits tests the GetPlanLimits method
func TestTenant_GetPlanLimits(t *testing.T) {
	tests := []struct {
		name            string
		tenant          Tenant
		expectedDevices int
		expectedUsers   int
		expectedAlerts  int
	}{
		{
			name:            "free plan",
			tenant:          Tenant{Plan: "free"},
			expectedDevices: 10,
			expectedUsers:   3,
			expectedAlerts:  20,
		},
		{
			name:            "pro plan",
			tenant:          Tenant{Plan: "pro"},
			expectedDevices: 100,
			expectedUsers:   20,
			expectedAlerts:  200,
		},
		{
			name:            "enterprise plan",
			tenant:          Tenant{Plan: "enterprise"},
			expectedDevices: -1,
			expectedUsers:   -1,
			expectedAlerts:  -1,
		},
		{
			name:            "unknown plan defaults to free",
			tenant:          Tenant{Plan: "unknown"},
			expectedDevices: 10,
			expectedUsers:   3,
			expectedAlerts:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, users, alerts := tt.tenant.GetPlanLimits()
			if devices != tt.expectedDevices {
				t.Errorf("MaxDevices = %d, want %d", devices, tt.expectedDevices)
			}
			if users != tt.expectedUsers {
				t.Errorf("MaxUsers = %d, want %d", users, tt.expectedUsers)
			}
			if alerts != tt.expectedAlerts {
				t.Errorf("MaxAlerts = %d, want %d", alerts, tt.expectedAlerts)
			}
		})
	}
}

// TestTenant_JSON tests JSON serialization
func TestTenant_JSON(t *testing.T) {
	now := time.Now()
	tenant := Tenant{
		ID:         "t-1",
		Name:       "Test Tenant",
		Slug:       "test-tenant",
		Plan:       "pro",
		MaxDevices: 50,
		IsActive:   true,
		Settings:   "{}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	data, err := json.Marshal(tenant)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Tenant
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ID != tenant.ID {
		t.Errorf("ID = %s, want %s", unmarshaled.ID, tenant.ID)
	}
	if unmarshaled.Name != tenant.Name {
		t.Errorf("Name = %s, want %s", unmarshaled.Name, tenant.Name)
	}
	if unmarshaled.Plan != tenant.Plan {
		t.Errorf("Plan = %s, want %s", unmarshaled.Plan, tenant.Plan)
	}
}

// TestTenantSettings_JSON tests TenantSettings JSON
func TestTenantSettings_JSON(t *testing.T) {
	settings := TenantSettings{
		EmailNotifications: true,
		SMSAlerts:          false,
		WebhookURL:         "https://example.com/webhook",
		CustomBranding: &CustomBranding{
			LogoURL:      "https://example.com/logo.png",
			PrimaryColor: "#00FF00",
			CompanyName:  "Test Company",
		},
		Features: map[string]bool{
			"ai_agent": true,
			"reports":  true,
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TenantSettings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.EmailNotifications != settings.EmailNotifications {
		t.Errorf("EmailNotifications = %v, want %v", unmarshaled.EmailNotifications, settings.EmailNotifications)
	}
	if unmarshaled.WebhookURL != settings.WebhookURL {
		t.Errorf("WebhookURL = %s, want %s", unmarshaled.WebhookURL, settings.WebhookURL)
	}
}

// TestTenantUsage_JSON tests TenantUsage struct
func TestTenantUsage_JSON(t *testing.T) {
	now := time.Now()
	usage := TenantUsage{
		TenantID:       "t-1",
		DeviceCount:    10,
		UserCount:      5,
		AlertCount:     3,
		DataPointsDay:  1000,
		LastCalculated: now,
	}

	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TenantUsage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.TenantID != usage.TenantID {
		t.Errorf("TenantID = %s, want %s", unmarshaled.TenantID, usage.TenantID)
	}
	if unmarshaled.DeviceCount != usage.DeviceCount {
		t.Errorf("DeviceCount = %d, want %d", unmarshaled.DeviceCount, usage.DeviceCount)
	}
}

// TestCreateTenantRequest_JSON tests request struct
func TestCreateTenantRequest_JSON(t *testing.T) {
	req := CreateTenantRequest{
		Name: "New Tenant",
		Slug: "new-tenant",
		Plan: "free",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled CreateTenantRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Name != req.Name {
		t.Errorf("Name = %s, want %s", unmarshaled.Name, req.Name)
	}
	if unmarshaled.Slug != req.Slug {
		t.Errorf("Slug = %s, want %s", unmarshaled.Slug, req.Slug)
	}
}

// TestTenantUpdateRequest_JSON tests update request
func TestTenantUpdateRequest_JSON(t *testing.T) {
	req := TenantUpdateRequest{
		Name:       "Updated Name",
		Slug:       "updated-slug",
		Plan:       "pro",
		MaxDevices: 100,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TenantUpdateRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Name != req.Name {
		t.Errorf("Name = %s, want %s", unmarshaled.Name, req.Name)
	}
}

// TestTenantResponse_JSON tests response struct
func TestTenantResponse_JSON(t *testing.T) {
	now := time.Now()
	resp := TenantResponse{
		Tenant: Tenant{
			ID:        "t-1",
			Name:      "Test",
			CreatedAt: now,
		},
		Usage: TenantUsage{
			TenantID:    "t-1",
			DeviceCount: 10,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TenantResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Tenant.ID != resp.Tenant.ID {
		t.Errorf("Tenant.ID = %s, want %s", unmarshaled.Tenant.ID, resp.Tenant.ID)
	}
}
