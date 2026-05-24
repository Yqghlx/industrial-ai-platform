package validation

import (
	"strings"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		id       string
		hasError bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", false}, // valid UUID
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", false}, // valid UUID
		{"invalid-uuid", true},
		{"", true},
		{"550e8400-e29b-41d4-a716", true},                    // too short
		{"550e8400-e29b-41d4-a716-446655440000-extra", true}, // too long
	}

	for _, tt := range tests {
		err := ValidateUUID(tt.id)
		if tt.hasError && err == nil {
			t.Errorf("ValidateUUID(%s) should return error", tt.id)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateUUID(%s) should not return error: %v", tt.id, err)
		}
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		id       string
		hasError bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", false}, // valid UUID
		{"device-123", false},                           // valid ID
		{"user_001", false},                             // valid ID
		{"test-id-underscore_123", false},               // valid ID
		{"", true},
		{"id with spaces", true},
		{"id@special!", true},
		{"a", false}, // single character is valid
	}

	for _, tt := range tests {
		err := ValidateID(tt.id)
		if tt.hasError && err == nil {
			t.Errorf("ValidateID(%s) should return error", tt.id)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateID(%s) should not return error: %v", tt.id, err)
		}
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		value    string
		field    string
		hasError bool
	}{
		{"valid value", "field", false},
		{"", "field", true},
		{"   ", "field", true},        // whitespace only
		{"  valid  ", "field", false}, // trimmed valid
	}

	for _, tt := range tests {
		err := ValidateRequired(tt.value, tt.field)
		if tt.hasError && err == nil {
			t.Errorf("ValidateRequired(%s, %s) should return error", tt.value, tt.field)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateRequired(%s, %s) should not return error: %v", tt.value, tt.field, err)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email    string
		hasError bool
	}{
		{"test@example.com", false},
		{"user.name@domain.org", false},
		{"user+tag@example.co.uk", false},
		{"", true},
		{"invalid", true},
		{"invalid@", true},
		{"@domain.com", true},
		{"user@.com", true},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if tt.hasError && err == nil {
			t.Errorf("ValidateEmail(%s) should return error", tt.email)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateEmail(%s) should not return error: %v", tt.email, err)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		minLen   int
		maxLen   int
		hasError bool
	}{
		{"password123", 8, 72, false},
		{"short", 8, 72, true},      // too short
		{"", 8, 72, true},           // empty
		{"validpass", 8, 72, false}, // exactly min length
		{"a", 1, 72, false},         // min length 1
	}

	for _, tt := range tests {
		err := ValidatePassword(tt.password, tt.minLen, tt.maxLen)
		if tt.hasError && err == nil {
			t.Errorf("ValidatePassword(%s, %d, %d) should return error", tt.password, tt.minLen, tt.maxLen)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidatePassword(%s, %d, %d) should not return error: %v", tt.password, tt.minLen, tt.maxLen, err)
		}
	}
}

func TestValidateLength(t *testing.T) {
	tests := []struct {
		value    string
		field    string
		minLen   int
		maxLen   int
		hasError bool
	}{
		{"valid", "field", 1, 10, false},
		{"", "field", 1, 10, true},             // too short
		{"toolongvalue", "field", 1, 10, true}, // too long
		{"exact10", "field", 1, 10, false},     // exactly max
	}

	for _, tt := range tests {
		err := ValidateLength(tt.value, tt.field, tt.minLen, tt.maxLen)
		if tt.hasError && err == nil {
			t.Errorf("ValidateLength(%s, %s, %d, %d) should return error", tt.value, tt.field, tt.minLen, tt.maxLen)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateLength(%s, %s, %d, %d) should not return error: %v", tt.value, tt.field, tt.minLen, tt.maxLen, err)
		}
	}
}

func TestValidatePagination(t *testing.T) {
	maxPageSize := 100

	tests := []struct {
		page       int
		pageSize   int
		expectPage int
		expectSize int
	}{
		{1, 20, 1, 20},   // valid
		{0, 20, 1, 20},   // page too small -> 1
		{-1, 20, 1, 20},  // negative page -> 1
		{1, 0, 1, 20},    // size too small -> default 20
		{1, 200, 1, 100}, // size too large -> max 100
		{5, 50, 5, 50},   // valid
	}

	for _, tt := range tests {
		result := ValidatePagination(tt.page, tt.pageSize, maxPageSize)
		if result.Page != tt.expectPage {
			t.Errorf("ValidatePagination(%d, %d) page = %d, expected %d", tt.page, tt.pageSize, result.Page, tt.expectPage)
		}
		if result.PageSize != tt.expectSize {
			t.Errorf("ValidatePagination(%d, %d) pageSize = %d, expected %d", tt.page, tt.pageSize, result.PageSize, tt.expectSize)
		}
	}
}

func TestValidateTenantPlan(t *testing.T) {
	tests := []struct {
		plan     string
		hasError bool
	}{
		{"free", false},
		{"basic", false},
		{"pro", false},
		{"enterprise", false},
		{"", true},
		{"invalid", true},
		{"FREE", true}, // case sensitive
	}

	for _, tt := range tests {
		err := ValidateTenantPlan(tt.plan)
		if tt.hasError && err == nil {
			t.Errorf("ValidateTenantPlan(%s) should return error", tt.plan)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateTenantPlan(%s) should not return error: %v", tt.plan, err)
		}
	}
}

func TestValidateTenantStatus(t *testing.T) {
	tests := []struct {
		status   string
		hasError bool
	}{
		{"active", false},
		{"suspended", false},
		{"deleted", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := ValidateTenantStatus(tt.status)
		if tt.hasError && err == nil {
			t.Errorf("ValidateTenantStatus(%s) should return error", tt.status)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateTenantStatus(%s) should not return error: %v", tt.status, err)
		}
	}
}

func TestValidateDeviceStatus(t *testing.T) {
	tests := []struct {
		status   string
		hasError bool
	}{
		{"online", false},
		{"offline", false},
		{"maintenance", false},
		{"error", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := ValidateDeviceStatus(tt.status)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDeviceStatus(%s) should return error", tt.status)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDeviceStatus(%s) should not return error: %v", tt.status, err)
		}
	}
}

func TestValidateUserRole(t *testing.T) {
	tests := []struct {
		role     string
		hasError bool
	}{
		{"admin", false},
		{"operator", false},
		{"viewer", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := ValidateUserRole(tt.role)
		if tt.hasError && err == nil {
			t.Errorf("ValidateUserRole(%s) should return error", tt.role)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateUserRole(%s) should not return error: %v", tt.role, err)
		}
	}
}

func TestValidateUserStatus(t *testing.T) {
	tests := []struct {
		status   string
		hasError bool
	}{
		{"active", false},
		{"disabled", false},
		{"pending", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := ValidateUserStatus(tt.status)
		if tt.hasError && err == nil {
			t.Errorf("ValidateUserStatus(%s) should return error", tt.status)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateUserStatus(%s) should not return error: %v", tt.status, err)
		}
	}
}

func TestValidateDeviceID(t *testing.T) {
	tests := []struct {
		id       string
		hasError bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", false}, // valid UUID
		{"device-123", false},                           // valid ID
		{"CNC-001", false},                              // valid ID
		{"DEV_001", false},                              // valid ID with underscore
		{"", true},                                      // empty
		{"id with spaces", true},                        // invalid characters
		{"id@special!", true},                           // invalid characters
		{"a", false},                                    // single character
		{makeString(101), true},                         // too long
	}

	for _, tt := range tests {
		err := ValidateDeviceID(tt.id)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDeviceID(%s) should return error", tt.id)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDeviceID(%s) should not return error: %v", tt.id, err)
		}
	}
}

func TestValidateDeviceName(t *testing.T) {
	tests := []struct {
		name     string
		hasError bool
	}{
		{"Temperature Sensor", false}, // valid
		{"CNC加工中心-1号", false},         // valid with Chinese
		{"Device 123", false},         // valid
		{"", true},                    // empty
		{"   ", true},                 // whitespace only
		{makeString(201), true},       // too long
		{"A", false},                  // single character
	}

	for _, tt := range tests {
		err := ValidateDeviceName(tt.name)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDeviceName(%s) should return error", tt.name)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDeviceName(%s) should not return error: %v", tt.name, err)
		}
	}
}

func TestValidateDeviceType(t *testing.T) {
	tests := []struct {
		deviceType string
		hasError   bool
	}{
		{"CNC", false},             // valid
		{"InjectionMolder", false}, // valid
		{"AssemblyRobot", false},   // valid
		{"Conveyor", false},        // valid
		{"Sensor", false},          // valid
		{"sensor", false},          // valid (legacy)
		{"gauge", false},           // valid
		{"PLC", false},             // valid
		{"robot", false},           // valid
		{"motor", false},           // valid
		{"pump", false},            // valid
		{"valve", false},           // valid
		{"heater", false},          // valid
		{"cooler", false},          // valid
		{"", true},                 // empty
		{"invalid_type", true},     // invalid
		{"UnknownDevice", true},    // invalid
	}

	for _, tt := range tests {
		err := ValidateDeviceType(tt.deviceType)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDeviceType(%s) should return error", tt.deviceType)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDeviceType(%s) should not return error: %v", tt.deviceType, err)
		}
	}
}

func TestValidateDeviceDescription(t *testing.T) {
	tests := []struct {
		description string
		hasError    bool
	}{
		{"This is a valid description", false}, // valid
		{"", false},                            // empty is valid
		{makeString(1001), true},               // too long
		{"Short desc", false},                  // valid short
	}

	for _, tt := range tests {
		err := ValidateDeviceDescription(tt.description)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDeviceDescription(length=%d) should return error", len(tt.description))
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDeviceDescription(length=%d) should not return error: %v", len(tt.description), err)
		}
	}
}

func TestValidateDevice(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		deviceType  string
		status      string
		description string
		hasError    bool
	}{
		{"valid device", "device-1", "sensor", "online", "A sensor device", false},
		{"valid device no desc", "device-2", "CNC", "offline", "", false},
		{"empty id", "", "sensor", "online", "", true},
		{"invalid id", "id@invalid", "sensor", "online", "", true},
		{"empty name", "device-1", "", "online", "", true}, // name required in ValidateDeviceName
		{"invalid type", "device-1", "invalid", "online", "", true},
		{"invalid status", "device-1", "sensor", "invalid_status", "", true},
		{"too long description", "device-1", "sensor", "online", makeString(1001), true},
	}

	for _, tt := range tests {
		err := ValidateDevice(tt.id, tt.name, tt.deviceType, tt.status, tt.description)
		if tt.hasError && err == nil {
			t.Errorf("ValidateDevice(%s) should return error", tt.name)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateDevice(%s) should not return error: %v", tt.name, err)
		}
	}
}

// Helper function to create a string of specific length
func makeString(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}

// FIX-P2: Password complexity tests
func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{"valid password", "Admin@123456!", false, ""},
		{"too short", "Abc@1", true, "12 characters"},
		{"missing uppercase", "admin@123456!", true, "uppercase"},
		{"missing lowercase", "ADMIN@123456!", true, "lowercase"},
		{"missing digit", "Admin@Password!", true, "digit"},
		{"missing special", "Admin12345678", true, "special"},
		{"empty password", "", true, "12 characters"},
		{"12 chars valid", "Admin@123456", false, ""},
		{"very long valid", "MyVeryLongP@ssw0rd!2024#Secure", false, ""},
		{"too long over 128", makeComplexString(129), true, "128"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordComplexity(%s) expected error, got nil", tt.password)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePasswordComplexity(%s) unexpected error: %v", tt.password, err)
				}
			}
		})
	}
}

func TestValidatePasswordWithComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		minLen   int
		maxLen   int
		wantErr  bool
	}{
		{"valid", "Admin@123456!", 8, 128, false},
		{"too short by minLen", "Admin@123456!", 20, 128, true}, // fails length check
		{"too short by complexity", "Abc@1", 1, 128, true},      // fails complexity (12 chars)
		{"too long", makeComplexString(130), 1, 128, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordWithComplexity(tt.password, tt.minLen, tt.maxLen)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPasswordComplexityError(t *testing.T) {
	err := &PasswordComplexityError{Errors: []string{"too short", "missing uppercase"}}
	expected := "password validation failed: too short, missing uppercase"
	if err.Error() != expected {
		t.Errorf("Error() = '%s', want '%s'", err.Error(), expected)
	}
}

// Helper for creating a long complex password
func makeComplexString(length int) string {
	result := "Aa1!"
	for i := 4; i < length; i++ {
		result += "x"
	}
	return result
}
