// Package validation provides input validation utilities
// FIX-027: 输入验证UUID
package validation

import (
	"errors"
	"regexp"
	"strings"
)

const (
	// UUID格式: 8-4-4-4-12 hex digits
	UUIDPattern = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	// 简单ID格式: alphanumeric, dash, underscore
	IDPattern = `^[a-zA-Z0-9_-]+$`
	// Email格式
	EmailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var (
	uuidRegex  = regexp.MustCompile(UUIDPattern)
	idRegex    = regexp.MustCompile(IDPattern)
	emailRegex = regexp.MustCompile(EmailPattern)
)

// ValidateUUID validates UUID format
func ValidateUUID(id string) error {
	if id == "" {
		return errors.New("UUID is required")
	}
	if !uuidRegex.MatchString(id) {
		return errors.New("invalid UUID format")
	}
	return nil
}

// ValidateID validates general ID format (UUID or safe string)
func ValidateID(id string) error {
	if id == "" {
		return errors.New("ID is required")
	}
	// Allow UUID format
	if uuidRegex.MatchString(id) {
		return nil
	}
	// Allow safe ID format
	if len(id) < 1 || len(id) > 100 {
		return errors.New("ID length must be between 1 and 100")
	}
	if !idRegex.MatchString(id) {
		return errors.New("ID contains invalid characters")
	}
	return nil
}

// ValidateRequired validates required field
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// Password complexity error types
// FIX-P2: 增强密码复杂度验证
type PasswordComplexityError struct {
	Errors []string
}

func (e *PasswordComplexityError) Error() string {
	return "password validation failed: " + strings.Join(e.Errors, ", ")
}

// ValidatePassword validates password length (basic validation)
// For stronger validation, use ValidatePasswordComplexity
func ValidatePassword(password string, minLen, maxLen int) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < minLen {
		return errors.New("password is too short")
	}
	if len(password) > maxLen {
		return errors.New("password is too long")
	}
	return nil
}

// ValidatePasswordComplexity validates password with complexity requirements
// FIX-P2: 增强密码复杂度验证 - 检查大小写字母、数字、特殊字符
// Requirements:
//   - Minimum length of 12 characters
//   - At least one uppercase letter
//   - At least one lowercase letter
//   - At least one digit
//   - At least one special character
func ValidatePasswordComplexity(password string) error {
	var errors []string

	// Check minimum length (12 characters minimum for security)
	if len(password) < 12 {
		errors = append(errors, "must be at least 12 characters long")
	}

	// Check maximum length (prevent DoS)
	if len(password) > 128 {
		errors = append(errors, "must be at most 128 characters long")
	}

	// Check for uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		errors = append(errors, "must contain at least one uppercase letter")
	}

	// Check for lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		errors = append(errors, "must contain at least one lowercase letter")
	}

	// Check for digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		errors = append(errors, "must contain at least one digit")
	}

	// Check for special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?` + "`" + `~]`).MatchString(password)
	if !hasSpecial {
		errors = append(errors, "must contain at least one special character (e.g., !@#$%^&*)")
	}

	if len(errors) > 0 {
		return &PasswordComplexityError{Errors: errors}
	}

	return nil
}

// ValidatePasswordWithComplexity validates password with both length and complexity requirements
// FIX-P2: 组合长度和复杂度验证
func ValidatePasswordWithComplexity(password string, minLen, maxLen int) error {
	// First validate length
	if err := ValidatePassword(password, minLen, maxLen); err != nil {
		return err
	}
	// Then validate complexity
	return ValidatePasswordComplexity(password)
}

// ValidateLength validates string length
func ValidateLength(value, fieldName string, minLen, maxLen int) error {
	if len(value) < minLen {
		return errors.New(fieldName + " is too short")
	}
	if len(value) > maxLen {
		return errors.New(fieldName + " is too long")
	}
	return nil
}

// Pagination validation result
type PaginationParams struct {
	Page     int
	PageSize int
}

// ValidatePagination validates and normalizes pagination params
func ValidatePagination(page, pageSize int, maxPageSize int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // default
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// TenantPlan validation
func ValidateTenantPlan(plan string) error {
	if plan == "" {
		return errors.New("tenant plan is required")
	}
	validPlans := []string{"free", "basic", "pro", "enterprise"}
	for _, valid := range validPlans {
		if plan == valid {
			return nil
		}
	}
	return errors.New("invalid tenant plan")
}

// TenantStatus validation
func ValidateTenantStatus(status string) error {
	if status == "" {
		return errors.New("tenant status is required")
	}
	validStatuses := []string{"active", "suspended", "deleted"}
	for _, valid := range validStatuses {
		if status == valid {
			return nil
		}
	}
	return errors.New("invalid tenant status")
}

// DeviceStatus validation
func ValidateDeviceStatus(status string) error {
	if status == "" {
		return errors.New("device status is required")
	}
	validStatuses := []string{"online", "offline", "maintenance", "error"}
	for _, valid := range validStatuses {
		if status == valid {
			return nil
		}
	}
	return errors.New("invalid device status")
}

// UserRole validation
func ValidateUserRole(role string) error {
	if role == "" {
		return errors.New("user role is required")
	}
	validRoles := []string{"admin", "operator", "viewer"}
	for _, valid := range validRoles {
		if role == valid {
			return nil
		}
	}
	return errors.New("invalid user role")
}

// UserStatus validation
func ValidateUserStatus(status string) error {
	if status == "" {
		return errors.New("user status is required")
	}
	validStatuses := []string{"active", "disabled", "pending"}
	for _, valid := range validStatuses {
		if status == valid {
			return nil
		}
	}
	return errors.New("invalid user status")
}

// Device ID validation constants
const (
	MinDeviceIDLength          = 1
	MaxDeviceIDLength          = 100
	MinDeviceNameLength        = 1
	MaxDeviceNameLength        = 200
	MaxDeviceDescriptionLength = 1000
)

// ValidDeviceTypes contains the list of valid device types
var ValidDeviceTypes = []string{
	"CNC",             // CNC加工中心
	"InjectionMolder", // 注塑机
	"AssemblyRobot",   // 装配机器人
	"Conveyor",        // 输送带
	"Sensor",          // 传感器节点
	"sensor",          // 传感器 (legacy)
	"gauge",           // 仪表
	"PLC",             // 可编程逻辑控制器
	"robot",           // 机器人
	"motor",           // 电机
	"pump",            // 泵
	"valve",           // 阀门
	"heater",          // 加热器
	"cooler",          // 冷却器
}

// ValidateDeviceID validates device ID format (length and characters)
func ValidateDeviceID(id string) error {
	if id == "" {
		return errors.New("device ID is required")
	}
	if len(id) < MinDeviceIDLength {
		return errors.New("device ID is too short")
	}
	if len(id) > MaxDeviceIDLength {
		return errors.New("device ID is too long (max 100 characters)")
	}
	// Allow UUID format
	if uuidRegex.MatchString(id) {
		return nil
	}
	// Allow safe ID format: alphanumeric, dash, underscore
	if !idRegex.MatchString(id) {
		return errors.New("device ID contains invalid characters (only alphanumeric, dash, underscore allowed)")
	}
	return nil
}

// ValidateDeviceName validates device name length
func ValidateDeviceName(name string) error {
	if name == "" {
		return errors.New("device name is required")
	}
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < MinDeviceNameLength {
		return errors.New("device name is too short")
	}
	if len(trimmed) > MaxDeviceNameLength {
		return errors.New("device name is too long (max 200 characters)")
	}
	return nil
}

// ValidateDeviceType validates device type against allowed enum values
func ValidateDeviceType(deviceType string) error {
	if deviceType == "" {
		return errors.New("device type is required")
	}
	for _, valid := range ValidDeviceTypes {
		if deviceType == valid {
			return nil
		}
	}
	return errors.New("invalid device type: " + deviceType)
}

// ValidateDeviceDescription validates device description length
func ValidateDeviceDescription(description string) error {
	if len(description) > MaxDeviceDescriptionLength {
		return errors.New("device description is too long (max 1000 characters)")
	}
	return nil
}

// ValidateDevice validates all device fields
func ValidateDevice(id, name, deviceType, status, description string) error {
	if err := ValidateDeviceID(id); err != nil {
		return err
	}
	if err := ValidateDeviceName(name); err != nil {
		return err
	}
	if err := ValidateDeviceType(deviceType); err != nil {
		return err
	}
	if err := ValidateDeviceStatus(status); err != nil {
		return err
	}
	if description != "" {
		if err := ValidateDeviceDescription(description); err != nil {
			return err
		}
	}
	return nil
}
