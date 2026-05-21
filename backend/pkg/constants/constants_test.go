package constants

import "testing"

func TestIsValidTenantPlan(t *testing.T) {
	tests := []struct {
		plan     string
		expected bool
	}{
		{"free", true},
		{"basic", true},
		{"pro", true},
		{"enterprise", true},
		{"invalid", false},
		{"", false},
		{"FREE", false}, // case sensitive
	}

	for _, tt := range tests {
		result := IsValidTenantPlan(tt.plan)
		if result != tt.expected {
			t.Errorf("IsValidTenantPlan(%s) = %v, expected %v", tt.plan, result, tt.expected)
		}
	}
}

func TestIsValidTenantStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"active", true},
		{"suspended", true},
		{"deleted", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidTenantStatus(tt.status)
		if result != tt.expected {
			t.Errorf("IsValidTenantStatus(%s) = %v, expected %v", tt.status, result, tt.expected)
		}
	}
}

func TestIsValidDeviceStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"online", true},
		{"offline", true},
		{"maintenance", true},
		{"error", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidDeviceStatus(tt.status)
		if result != tt.expected {
			t.Errorf("IsValidDeviceStatus(%s) = %v, expected %v", tt.status, result, tt.expected)
		}
	}
}

func TestIsValidUserRole(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{"admin", true},
		{"operator", true},
		{"viewer", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidUserRole(tt.role)
		if result != tt.expected {
			t.Errorf("IsValidUserRole(%s) = %v, expected %v", tt.role, result, tt.expected)
		}
	}
}

func TestGetRoleLevel(t *testing.T) {
	tests := []struct {
		role     string
		expected int
	}{
		{"admin", 3},
		{"operator", 2},
		{"viewer", 1},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := GetRoleLevel(tt.role)
		if result != tt.expected {
			t.Errorf("GetRoleLevel(%s) = %v, expected %v", tt.role, result, tt.expected)
		}
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		role     string
		required string
		expected bool
	}{
		{"admin", "viewer", true},     // admin(3) >= viewer(1)
		{"admin", "operator", true},   // admin(3) >= operator(2)
		{"admin", "admin", true},      // admin(3) >= admin(3)
		{"operator", "viewer", true},  // operator(2) >= viewer(1)
		{"operator", "admin", false},  // operator(2) < admin(3)
		{"viewer", "operator", false}, // viewer(1) < operator(2)
		{"viewer", "viewer", true},    // viewer(1) >= viewer(1)
		{"unknown", "viewer", false},  // unknown(0) < viewer(1)
	}

	for _, tt := range tests {
		result := HasPermission(tt.role, tt.required)
		if result != tt.expected {
			t.Errorf("HasPermission(%s, %s) = %v, expected %v", tt.role, tt.required, result, tt.expected)
		}
	}
}

func TestPaginationConstants(t *testing.T) {
	if DefaultPageSize <= 0 {
		t.Error("DefaultPageSize should be positive")
	}
	if MaxPageSize <= DefaultPageSize {
		t.Error("MaxPageSize should be greater than DefaultPageSize")
	}
	if MinPageSize <= 0 {
		t.Error("MinPageSize should be positive")
	}
}

func TestTokenDurationConstants(t *testing.T) {
	if AccessTokenDurationHours <= 0 {
		t.Error("AccessTokenDurationHours should be positive")
	}
	if RefreshTokenDurationHours <= AccessTokenDurationHours {
		t.Error("RefreshTokenDurationHours should be greater than AccessTokenDurationHours")
	}
	if TokenMinSecretLength < 32 {
		t.Error("TokenMinSecretLength should be at least 32")
	}
}

func TestPasswordConstants(t *testing.T) {
	if PasswordMinLength < 8 {
		t.Error("PasswordMinLength should be at least 8")
	}
	if PasswordMaxLength < PasswordMinLength {
		t.Error("PasswordMaxLength should be greater than PasswordMinLength")
	}
}
